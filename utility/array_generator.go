package utility

import (

	"math/rand"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/rlwe"
)

func (u Utils) Get2PowRotationEvaluator() ckks.Evaluator {

	return u.Evaluator.WithKey(rlwe.EvaluationKey{Rlk: u.KeyChain.RelinKey, Rtks: u.KeyChain.GaloisKey})

}

func (u Utils) Float64ToComplex128(value []float64) []complex128 {

	cmplx := make([]complex128, len(value))
	for i := range value {
		cmplx[i] = complex(value[i], 0)
	}
	return cmplx

}

func (u Utils) Complex128ToFloat64(value []complex128) []float64 {

	flt := make([]float64, len(value))
	for i := range value {
		flt[i] = real(value[i])
	}
	return flt

}

func (u Utils) GenerateFilledArray(value float64) []float64 {

	arr := make([]float64, u.Params.Slots())
	for i := range arr {
		arr[i] = value
	}

	return arr

}

func (u Utils) GenerateFilledArraySize(value float64, size int) []float64 {

	arr := make([]float64, u.Params.Slots())
	for i := 0; i < size; i++ {
		arr[i] = value
	}

	return arr

}

func (u Utils) GenerateRandomNormalArray(length int, stdDev float64) []float64 {

	randomArr := make([]float64, u.Params.Slots())
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < length; i++ {
		randomArr[i] = rand.NormFloat64() * stdDev
	}

	return randomArr

}

func (u Utils) GenerateRandomNormalArraySeed(length int, stdDev float64, seed int) []float64 {

	randomArr := make([]float64, u.Params.Slots())
	rand.Seed(int64(seed))

	for i := 0; i < length; i++ {
		randomArr[i] = rand.NormFloat64() * stdDev
	}

	return randomArr

}
