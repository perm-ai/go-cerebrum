package activations

import (
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					SOFTMAX
//=================================================

type Softmax struct {
	U              utility.Utils
	zeroEliminator map[int]ckks.Plaintext
}

func NewSoftmax(u utility.Utils) Softmax {

	eliminator := make(map[int]ckks.Plaintext)

	return Softmax{u, eliminator}

}

func (s Softmax) Forward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	// Homomorphic friendly softmax function
	// e^x / (e^x1 + e^x2 + ... + e^xn)
	// Cost 7 level if rescale is false and 8 if rescale is true

	// Encode filter if doesn't exist in cache
	sum := s.U.Encrypt(s.U.GenerateFilledArray(0.0))

	if _, ok := s.zeroEliminator[inputLength]; !ok {

		arr := s.U.GenerateFilledArray(1)

		for i := 0; i < inputLength; i++ {
			arr[i] = 0
		}

		s.zeroEliminator[inputLength] = *s.U.Encoder.EncodeNTTNew(s.U.Float64ToComplex128(arr), s.U.Params.LogSlots())
	}
	//create array that will contain the result, but for the first loop, contain the e^x of each input
	arrexp := make([]*ckks.Ciphertext, len(input))
	// Exponentiate input and get sum
	for i, p := range input {
		arrexp[i] = s.U.ExpNew(p)

		s.U.SubPlain(*arrexp[i], s.zeroEliminator[inputLength], arrexp[i])
		// Level input - 2
		if arrexp[i].Scale > math.Pow(2, 40.1) && arrexp[i].Level() == 6 {
			s.U.Evaluator.Rescale(arrexp[i], math.Pow(2, 35), arrexp[i])
		}
		s.U.Add(sum, *arrexp[i], &sum)
	}

	// Declare stretch scale as 1/40
	stretchScale := (float64(1) / float64(40))
	plainStretch := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(stretchScale, inputLength))
	// Calculate inverse of sum of e^input
	inverseSum := s.U.InverseApproxNew(&sum, stretchScale) // Level input - 4

	//multiply the arrexp with stretchscale and the inverse, which will be the result that the function return
	for i := range arrexp {
		s.U.MultiplyPlain(arrexp[i], &plainStretch, arrexp[i], true, false) // Level input - 3
		s.U.Multiply(*arrexp[i], *inverseSum, arrexp[i], false, false)
	}

	return arrexp // Level input - 5

}

func (s Softmax) Backward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	// Not implemented, won't be used
	return input

}

func (s Softmax) GetForwardLevelConsumption() int {

	return 7

}

func (s Softmax) GetBackwardLevelConsumption() int {

	return 0

}
