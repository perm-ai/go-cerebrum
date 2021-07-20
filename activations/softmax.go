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
	U          		utility.Utils
	zeroEliminator 	map[int]ckks.Plaintext
}

func NewSoftmax(u utility.Utils) Softmax {

	eliminator := make(map[int]ckks.Plaintext)

	return Softmax{u, eliminator}

}

func (s Softmax) Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// Homomorphic friendly softmax function
	// e^x / (e^x1 + e^x2 + ... + e^xn)
	// Cost 7 level if rescale is false and 8 if rescale is true

	// Encode filter if doesn't exist in cache
	if _, ok := s.zeroEliminator[inputLength]; !ok {

		arr := s.U.GenerateFilledArray(1)

		for i := 0; i < inputLength; i++ {
			arr[i] = 0
		}

		s.zeroEliminator[inputLength] = *s.U.Encoder.EncodeNTTNew(s.U.Float64ToComplex128(arr), s.U.Params.LogSlots())
	}

	// Exponentiate input
	exp := s.U.ExpNew(&input) // Level input - 2

	if exp.Scale() > math.Pow(2, 40.1) && exp.Level() == 6 {
		s.U.Evaluator.Rescale(exp, math.Pow(2, 35), exp)
	}

	// Filter 1 (since e^0 = 1)
	s.U.SubPlain(*exp, s.zeroEliminator[inputLength], exp)

	expSum := s.U.SumElementsNew(*exp)

	// Declare stretch scale as 1/40
	stretchScale := (float64(1) / float64(40))

	// Calculate inverse of sum of e^input
	inverseSum := s.U.InverseApproxNew(&expSum, stretchScale) // Level input - 4

	// Apply stretch scale to exponentiated input
	s.U.MultiplyConstArray(exp, s.U.GenerateFilledArraySize(stretchScale, inputLength), exp, true, false) // Level input - 3

	return s.U.MultiplyNew(*exp, *inverseSum, false, false) // Level input - 5

}

func (s Softmax) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// Not implemented, won't be used
	return input

}
