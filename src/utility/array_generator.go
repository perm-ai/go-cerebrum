package utility

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/rlwe"
)

func getSumElementsKs(logSlots int) []int {

	ks := []int{}

	for i := 0; i <= logSlots; i++{
		positive := int(math.Pow(2, float64(i)))
		ks = append(ks, positive)
		ks = append(ks, (-1 * positive))
	}

	sort.Ints(ks[:])

	return ks

}

func (u Utils) Get2PowRotationEvaluator() ckks.Evaluator {

	return u.Evaluator.WithKey(rlwe.EvaluationKey{Rlk: &u.RelinKey, Rtks: &u.GaloisKey})

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

func (u Utils) GenerateRandomNormalArray(length int) []float64 {

	randomArr := make([]float64, u.Params.Slots())
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < length; i++ {
		randomArr[i] = rand.NormFloat64()
	}

	return randomArr

}