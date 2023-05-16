package key

import (
	"sync"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type SwitchingKeyLoader func(uint64) *rlwe.SwitchingKey

type CustomKeyGenerator struct {
	ckksParams 		 ckks.Parameters
	params           rlwe.Parameters
	keyGenerator	 rlwe.KeyGenerator
}

func NewKeyGenerator(params ckks.Parameters, keyGen rlwe.KeyGenerator) *CustomKeyGenerator {

	return &CustomKeyGenerator{
		ckksParams:		  params,
		params:           params.Parameters,
		keyGenerator: 	  keyGen,
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

func (keygen *CustomKeyGenerator) GenRotationKeysForRotations(ks []int, includeConjugate bool, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey) error) {
	galEls := keygen.GetGalEl(ks, includeConjugate)
	if includeConjugate {
		galEls = append(galEls, keygen.params.GaloisElementForRowRotation())
	}
	keygen.GenRotationKeys(galEls, sk, callback)
}

func (keygen *CustomKeyGenerator) GenRotationKeys(galEls []uint64, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey) error) []error {
	errs := make([]error, len(galEls))

	for i, galEl := range galEls {
		swk := keygen.keyGenerator.GenSwitchingKeyForGalois(galEl, sk)
		errs[i] = callback(galEl, swk)
	}

	return errs
}

func (keygen *CustomKeyGenerator) GenRotationKeysConcurrent(galEls []uint64, sk *rlwe.SecretKey, callback func(galEl uint64, swk *rlwe.SwitchingKey) error) []error {

	var wg sync.WaitGroup
	errs := make([]error, len(galEls))

	for i, galEl := range galEls {

		wg.Add(1)
		kg := ckks.NewKeyGenerator(keygen.ckksParams)

		go func(galElGoRoutine uint64, kgRoutine *rlwe.KeyGenerator, index int){
			
			defer wg.Done()

			swk := (*kgRoutine).GenSwitchingKeyForGalois(galElGoRoutine, sk)
			errs[index] = callback(galElGoRoutine, swk)

		}(galEl, &kg, i)
		
	}

	wg.Wait()

	return errs
}