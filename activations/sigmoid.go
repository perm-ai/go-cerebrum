package activations

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					SIGMOID
//=================================================

type Sigmoid struct {
	U utility.Utils
}

func (s Sigmoid) Forward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	outputChannels := make([]chan *rlwe.Ciphertext, len(input))
	output := make([]*rlwe.Ciphertext, len(input))

	deg3Coeff := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(-0.004, inputLength))
	deg1Coeff := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(0.197, inputLength))
	deg0 := *s.U.Encoder.EncodeNew(s.U.Float64ToComplex128(s.U.GenerateFilledArraySize(0.5, inputLength)), s.U.Params.MaxLevel(), s.U.Params.NewScale(s.U.Scale), s.U.Params.LogSlots())

	for i := range input {

		outputChannels[i] = make(chan *rlwe.Ciphertext)

		go func(inputEach *rlwe.Ciphertext, utils utility.Utils, c chan *rlwe.Ciphertext) {

			// y := 0.5 + 0.197x - 0.004x^3
			// Calculate degree three
			xSquared := utils.MultiplyNew(inputEach.CopyNew(), inputEach.CopyNew(), true, false)
			deg3 := utils.MultiplyPlainNew(inputEach.CopyNew(), deg3Coeff, true, false)
			utils.Multiply(xSquared, deg3, deg3, true, false)

			// Calculate degree one
			deg1 := utils.MultiplyPlainNew(inputEach.CopyNew(), deg1Coeff, true, false)

			// Add all degree together
			result := utils.AddNew(deg3, deg1)
			utils.AddPlain(result, &deg0, result)
			c <- result

		}(input[i], s.U.CopyWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels {
		output[i] = <-outputChannels[i]
	}

	return output

}

func (s Sigmoid) Backward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	outputChannels := make([]chan *rlwe.Ciphertext, len(input))
	output := make([]*rlwe.Ciphertext, len(input))
	
	deg2Coeff := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(0.012, inputLength))
	deg0 := *s.U.Encoder.EncodeNew(s.U.Float64ToComplex128(s.U.GenerateFilledArraySize(0.197, inputLength)), s.U.Params.MaxLevel(), s.U.Params.NewScale(s.U.Scale), s.U.Params.LogSlots())

	for i := range input {
		// 0.012x^2 + 0.197

		outputChannels[i] = make(chan *rlwe.Ciphertext)

		go func(inputEach *rlwe.Ciphertext, utils utility.Utils, c chan *rlwe.Ciphertext) {

			// Calculate degree three
			xSquared := utils.MultiplyNew(inputEach.CopyNew(), inputEach.CopyNew(), true, false)
			deg2 := utils.MultiplyPlainNew(xSquared, deg2Coeff, true, false)

			// Add all degree together
			result := utils.AddPlainNew(deg2, &deg0)

			c <- result

		}(input[i], s.U.CopyWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels {
		output[i] = <-outputChannels[i]
	}

	return output

}

func (s Sigmoid) GetForwardLevelConsumption() int {
	return 2
}

func (s Sigmoid) GetBackwardLevelConsumption() int {
	return 2
}

func (s Sigmoid) GetType() string {
	return "sigmoid"
}
