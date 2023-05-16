package activations

import (
	"math"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					SOFTMAX
//=================================================

type Softmax struct {
	U utility.Utils
}

func NewSoftmax(u utility.Utils) Softmax {

	newUtils := u.CopyWithNewScale(math.Pow(2, 30))
	return Softmax{newUtils}

}

func (s Softmax) Forward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	// Homomorphic friendly softmax function
	// e^x / (e^x1 + e^x2 + ... + e^xn)
	// Cost 7 level if rescale is false and 8 if rescale is true

	// Encode filter if doesn't exist in cache
	sum := s.U.Encrypt(s.U.GenerateFilledArray(0.0))

	//create array that will contain the result, but for the first loop, contain the e^x of each input
	arrexp := make([]*rlwe.Ciphertext, len(input))

	// Exponentiate input and get sum
	for i := range input {
		arrexp[i] = s.U.ExpNew(input[i], inputLength)
		s.U.Add(&sum, arrexp[i], &sum)
	}

	// Declare stretch scale as 1/40
	stretchScale := (float64(1) / float64(20))
	plainStretch := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(stretchScale, inputLength))

	// Calculate inverse of sum of e^input
	inverseSum := s.U.InverseApproxNew(&sum, stretchScale, inputLength) // Level input - 4

	output := make([]*rlwe.Ciphertext, len(arrexp))
	outputChannels := make([]chan *rlwe.Ciphertext, len(arrexp))

	//multiply the arrexp with stretchscale and the inverse, which will be the result that the function return
	for i := range arrexp {

		outputChannels[i] = make(chan *rlwe.Ciphertext)

		go func(inputEach *rlwe.Ciphertext, utils utility.Utils, c chan *rlwe.Ciphertext) {

			result := utils.MultiplyPlainNew(inputEach, plainStretch, true, false) // Level input - 3
			s.U.Multiply(result, inverseSum, result, false, false)
			c <- result

		}(arrexp[i], s.U.CopyWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels {
		output[i] = <-outputChannels[i]
	}

	return output // Level input - 5

}

func (s Softmax) Backward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	// Not implemented, won't be used
	return input

}

func (s Softmax) GetForwardLevelConsumption() int {

	return 8

}

func (s Softmax) GetBackwardLevelConsumption() int {

	return 0

}

func (s Softmax) GetType() string {
	return "softmax"
}
