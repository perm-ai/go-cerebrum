package activations

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					SIGMOID
//=================================================

type Sigmoid struct {
	U            utility.Utils
}

func (s Sigmoid) Forward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	outputChannels := make([]chan *ckks.Ciphertext, len(input))
	output := make([]*ckks.Ciphertext, len(input))

	for i := range input {

		outputChannels[i] = make(chan *ckks.Ciphertext)

		go func(inputEach *ckks.Ciphertext, utils utility.Utils, c chan *ckks.Ciphertext){

			// y := 0.5 + 0.197x - 0.004x^3
			// Calculate degree three
			xSquared := utils.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
			deg3Coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize(-0.004, inputLength))
			deg3 := utils.MultiplyPlainNew(inputEach.CopyNew(), &deg3Coeff, true, false)
			utils.Multiply(xSquared, deg3, &deg3, true, false)

			// Calculate degree one
			deg1Coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize(0.197, inputLength))
			deg1 := utils.MultiplyPlainNew(inputEach.CopyNew(), &deg1Coeff, true, false)

			// Encode deg0 as plaintext
			deg0 := *utils.Encoder.EncodeNTTNew(utils.Float64ToComplex128(utils.GenerateFilledArraySize(0.5, inputLength)), utils.Params.LogSlots())

			// Add all degree together
			result := utils.AddNew(deg3, deg1)
			utils.AddPlain(&result, &deg0, &result)
			c <- &result

		}(input[i], s.U.CopyUtilsWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels{
		output[i] = <-outputChannels[i]
	}

	return output

}

func (s Sigmoid) Backward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	outputChannels := make([]chan *ckks.Ciphertext, len(input))
	output := make([]*ckks.Ciphertext, len(input))

	for i := range input {
		// 0.012x^2 + 0.197

		outputChannels[i] = make(chan *ckks.Ciphertext)

		go func(inputEach *ckks.Ciphertext, utils utility.Utils, c chan *ckks.Ciphertext){

			// Calculate degree three
			xSquared := utils.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
			deg2Coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize(0.012, inputLength))
			deg2 := utils.MultiplyPlainNew(&xSquared, &deg2Coeff, true, false)

			// Encode deg0 as plaintext
			deg0 := *utils.Encoder.EncodeNTTNew(utils.Float64ToComplex128(utils.GenerateFilledArraySize(0.197, inputLength)), utils.Params.LogSlots())

			// Add all degree together
			result := utils.AddPlainNew(deg2, deg0)
			
			c <- &result

		}(input[i], s.U.CopyUtilsWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels{
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
