package key

import (
	"math/big"

	"github.com/ldsec/lattigo/v2/utils"
	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/rlwe"
)

type SwitchingKeyLoader func(uint64) *rlwe.SwitchingKey

type CustomKeyGenerator struct {
	params          rlwe.Parameters
	ringQP          *ring.Ring
	pBigInt         *big.Int
	polypool        [2]*ring.Poly
	gaussianSampler *ring.GaussianSampler
	uniformSampler  *ring.UniformSampler
}

func NewKeyGenerator(params rlwe.Parameters) *CustomKeyGenerator {

	ringQP := params.RingQP()

	prng, err := utils.NewPRNG()
	if err != nil {
		panic(err)
	}

	return &CustomKeyGenerator{
		params:          params,
		ringQP:          ringQP,
		pBigInt:         params.PBigInt(),
		polypool:        [2]*ring.Poly{ringQP.NewPoly(), ringQP.NewPoly()},
		gaussianSampler: ring.NewGaussianSampler(prng, ringQP, params.Sigma(), int(6*params.Sigma())),
		uniformSampler:  ring.NewUniformSampler(prng, ringQP),
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

func (keygen *CustomKeyGenerator) GenRotationKeysForRotations(ks []int, includeConjugate bool, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey)) {
	galEls := make([]uint64, len(ks), len(ks)+1)
	for i, k := range ks {
		galEls[i] = keygen.params.GaloisElementForColumnRotationBy(k)
	}
	if includeConjugate {
		galEls = append(galEls, keygen.params.GaloisElementForRowRotation())
	}
	keygen.GenRotationKeys(galEls, sk, callback)
}

func (keygen *CustomKeyGenerator) GenRotationKeys(galEls []uint64, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey)){
	for _, galEl := range galEls {
		switchingKey := rlwe.NewSwitchingKey(keygen.params)
		keygen.genrotKey(sk.Value, keygen.params.InverseGaloisElement(galEl), switchingKey)
		callback(galEl, switchingKey)
	}
}


func (keygen *CustomKeyGenerator) genrotKey(sk *ring.Poly, galEl uint64, swk *rlwe.SwitchingKey) {

	skIn := sk
	skOut := keygen.polypool[1]

	index := ring.PermuteNTTIndex(galEl, uint64(keygen.ringQP.N))
	ring.PermuteNTTWithIndexLvl(keygen.params.QPCount()-1, skIn, index, skOut)

	keygen.newSwitchingKey(skIn, skOut, swk)

	keygen.polypool[0].Zero()
	keygen.polypool[1].Zero()
}

func (keygen *CustomKeyGenerator) newSwitchingKey(skIn, skOut *ring.Poly, swk *rlwe.SwitchingKey) {

	ringQP := keygen.ringQP

	// Computes P * skIn
	ringQP.MulScalarBigint(skIn, keygen.pBigInt, keygen.polypool[0])

	alpha := keygen.params.PCount()
	beta := keygen.params.Beta()

	var index int
	for i := 0; i < beta; i++ {

		// e

		keygen.gaussianSampler.Read(swk.Value[i][0])
		ringQP.NTTLazy(swk.Value[i][0], swk.Value[i][0])
		ringQP.MForm(swk.Value[i][0], swk.Value[i][0])

		// a (since a is uniform, we consider we already sample it in the NTT and Montgomery domain)
		keygen.uniformSampler.Read(swk.Value[i][1])

		// e + (skIn * P) * (q_star * q_tild) mod QP
		//
		// q_prod = prod(q[i*alpha+j])
		// q_star = Q/qprod
		// q_tild = q_star^-1 mod q_prod
		//
		// Therefore : (skIn * P) * (q_star * q_tild) = sk*P mod q[i*alpha+j], else 0
		for j := 0; j < alpha; j++ {

			index = i*alpha + j

			qi := ringQP.Modulus[index]
			p0tmp := keygen.polypool[0].Coeffs[index]
			p1tmp := swk.Value[i][0].Coeffs[index]

			for w := 0; w < ringQP.N; w++ {
				p1tmp[w] = ring.CRed(p1tmp[w]+p0tmp[w], qi)
			}

			// It handles the case where nb pj does not divide nb qi
			if index >= keygen.params.QCount() {
				break
			}
		}

		// (skIn * P) * (q_star * q_tild) - a * skOut + e mod QP
		ringQP.MulCoeffsMontgomeryAndSub(swk.Value[i][1], skOut, swk.Value[i][0])
	}
}
