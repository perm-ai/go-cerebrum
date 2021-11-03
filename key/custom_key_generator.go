package key

import (
	"math"
	"math/big"

	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/rlwe"
	"github.com/ldsec/lattigo/v2/utils"
)

type SwitchingKeyLoader func(uint64) *rlwe.SwitchingKey

type CustomKeyGenerator struct {
	params           rlwe.Parameters
	poolQ            *ring.Poly
	poolQP           rlwe.PolyQP
	gaussianSamplerQ *ring.GaussianSampler
	uniformSamplerQ  *ring.UniformSampler
	uniformSamplerP  *ring.UniformSampler
}

func NewKeyGenerator(params rlwe.Parameters) *CustomKeyGenerator {

	prng, err := utils.NewPRNG()
	if err != nil {
		panic(err)
	}

	return &CustomKeyGenerator{
		params:           params,
		poolQ:            params.RingQ().NewPoly(),
		poolQP:           params.RingQP().NewPoly(),
		gaussianSamplerQ: ring.NewGaussianSampler(prng, params.RingQ(), params.Sigma(), int(6*params.Sigma())),
		uniformSamplerQ:  ring.NewUniformSampler(prng, params.RingQ()),
		uniformSamplerP:  ring.NewUniformSampler(prng, params.RingP()),
	}
}

func (keygen *CustomKeyGenerator) LoadRotationKeySet (ks []int, includeConjugate bool, loader SwitchingKeyLoader) *rlwe.RotationKeySet{

	galEls := make([]uint64, len(ks), len(ks)+1)
	for i, k := range ks {
		galEls[i] = keygen.params.GaloisElementForColumnRotationBy(k)
	}
	if includeConjugate {
		galEls = append(galEls, keygen.params.GaloisElementForRowRotation())
	}

	rlks := rlwe.NewRotationKeySet(keygen.params, galEls)

	for _, galEl := range galEls{
		rlks.Keys[galEl] = loader(galEl)
	}

	return rlks
}

func (keygen *CustomKeyGenerator) GetGalEl(ks []int, includeConjugate bool) []uint64 {
	galEls := make([]uint64, len(ks), len(ks)+1)
	for i, k := range ks {
		galEls[i] = keygen.params.GaloisElementForColumnRotationBy(k)
	}
	return galEls
}

func (keygen *CustomKeyGenerator) GenRotationKeysForRotations(ks []int, includeConjugate bool, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey)) {
	galEls := keygen.GetGalEl(ks, includeConjugate)
	if includeConjugate {
		galEls = append(galEls, keygen.params.GaloisElementForRowRotation())
	}
	keygen.GenRotationKeys(galEls, sk, callback)
}

func (keygen *CustomKeyGenerator) GenRotationKeys(galEls []uint64, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey)){
	for _, galEl := range galEls {
		switchingKey := rlwe.NewSwitchingKey(keygen.params, keygen.params.QCount()-1, keygen.params.PCount()-1)
		keygen.genrotKey(sk.Value, keygen.params.InverseGaloisElement(galEl), switchingKey)
		callback(galEl, switchingKey)
	}
}

func (keygen *CustomKeyGenerator) GenRotationKeysConcurrent(galEls []uint64, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey)){
	for _, galEl := range galEls {

		go func(galElGoRouting uint64){
			switchingKey := rlwe.NewSwitchingKey(keygen.params, keygen.params.QCount()-1, keygen.params.PCount()-1)
			keygen.genrotKey(sk.Value, keygen.params.InverseGaloisElement(galElGoRouting), switchingKey)
			callback(galElGoRouting, switchingKey)
		}(galEl)
		
	}
}


func (keygen *CustomKeyGenerator) genrotKey(sk rlwe.PolyQP, galEl uint64, swk *rlwe.SwitchingKey) {

	skIn := sk
	skOut := keygen.poolQP

	index := ring.PermuteNTTIndex(galEl, uint64(keygen.params.N()))
	ring.PermuteNTTWithIndexLvl(keygen.params.QCount()-1, skIn.Q, index, skOut.Q)
	ring.PermuteNTTWithIndexLvl(keygen.params.PCount()-1, skIn.P, index, skOut.P)

	keygen.genSwitchingKey(skIn.Q, skOut, swk)
}

func (keygen *CustomKeyGenerator) genSwitchingKey(skIn *ring.Poly, skOut rlwe.PolyQP, swk *rlwe.SwitchingKey) {

	ringQ := keygen.params.RingQ()
	ringQP := keygen.params.RingQP()

	levelQ := len(swk.Value[0][0].Q.Coeffs) - 1
	levelP := len(swk.Value[0][0].P.Coeffs) - 1

	var pBigInt *big.Int
	if levelP == keygen.params.PCount()-1 {
		pBigInt = keygen.params.RingP().ModulusBigint
	} else {
		P := keygen.params.RingP().Modulus
		pBigInt = new(big.Int).SetUint64(P[0])
		for i := 1; i < levelP+1; i++ {
			pBigInt.Mul(pBigInt, ring.NewUint(P[i]))
		}
	}

	// Computes P * skIn
	ringQ.MulScalarBigintLvl(levelQ, skIn, pBigInt, keygen.poolQ)

	alpha := levelP + 1
	beta := int(math.Ceil(float64(levelQ+1) / float64(levelP+1)))

	var index int
	for i := 0; i < beta; i++ {

		// e
		keygen.gaussianSamplerQ.ReadLvl(levelQ, swk.Value[i][0].Q)
		ringQP.ExtendBasisSmallNormAndCenter(swk.Value[i][0].Q, levelP, nil, swk.Value[i][0].P)
		ringQP.NTTLazyLvl(levelQ, levelP, swk.Value[i][0], swk.Value[i][0])
		ringQP.MFormLvl(levelQ, levelP, swk.Value[i][0], swk.Value[i][0])

		// a (since a is uniform, we consider we already sample it in the NTT and Montgomery domain)
		keygen.uniformSamplerQ.ReadLvl(levelQ, swk.Value[i][1].Q)
		keygen.uniformSamplerP.ReadLvl(levelP, swk.Value[i][1].P)

		// e + (skIn * P) * (q_star * q_tild) mod QP
		//
		// q_prod = prod(q[i*alpha+j])
		// q_star = Q/qprod
		// q_tild = q_star^-1 mod q_prod
		//
		// Therefore : (skIn * P) * (q_star * q_tild) = sk*P mod q[i*alpha+j], else 0
		for j := 0; j < alpha; j++ {

			index = i*alpha + j

			// It handles the case where nb pj does not divide nb qi
			if index >= levelQ+1 {
				break
			}

			qi := ringQ.Modulus[index]
			p0tmp := keygen.poolQ.Coeffs[index]
			p1tmp := swk.Value[i][0].Q.Coeffs[index]

			for w := 0; w < ringQ.N; w++ {
				p1tmp[w] = ring.CRed(p1tmp[w]+p0tmp[w], qi)
			}
		}

		// (skIn * P) * (q_star * q_tild) - a * skOut + e mod QP
		ringQP.MulCoeffsMontgomeryAndSubLvl(levelQ, levelP, swk.Value[i][1], skOut, swk.Value[i][0])
	}
}

