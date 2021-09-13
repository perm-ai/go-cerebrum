package activations

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					TANH
//=================================================

type Tanh struct {
	U            utility.Utils
	backwardDeg0 map[int]ckks.Plaintext
}

func NewTanh(utils utility.Utils) Tanh {

	return Tanh{utils, make(map[int]ckks.Plaintext)}

}

func (t Tanh) Forward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	// y = (-0.00752x^3) + (0.37x)

	output := make([]*ckks.Ciphertext, len(input))
	outputChannels := make([]chan *ckks.Ciphertext, len(input))

	for i := range input {

		outputChannels[i] = make(chan *ckks.Ciphertext)

		go func(inputEach *ckks.Ciphertext, utils utility.Utils, c chan *ckks.Ciphertext){

			// Calculate degree three
			xSquared := utils.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
			deg3 := utils.MultiplyConstNew(inputEach.CopyNew(), -0.00752, true, false)
			utils.Multiply(xSquared, deg3, &deg3, true, false)

			// Calculate degree one
			deg1 := utils.MultiplyConstNew(inputEach.CopyNew(), 0.37, true, false)

			// Add all degree together
			result := utils.AddNew(deg3, deg1)
			c <- &result

		}(input[i], t.U.CopyUtilsWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels{
		output[i] = <-outputChannels[i]
	}

	return output

}

func (t Tanh) Backward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	// (-0.02256x^2) + 0.37

	output := make([]*ckks.Ciphertext, len(input))
	outputChannels := make([]chan *ckks.Ciphertext, len(input))

	for i := range input {

		outputChannels[i] = make(chan *ckks.Ciphertext)

		go func(inputEach *ckks.Ciphertext, utils utility.Utils, c chan *ckks.Ciphertext){

			// Calculate degree three
			xSquared := utils.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
			deg2 := utils.MultiplyConstNew(&xSquared, -0.02256, true, false)

			// Encode deg0 as plaintext
			deg0 := *utils.Encoder.EncodeNTTNew(utils.Float64ToComplex128(utils.GenerateFilledArraySize(0.37, inputLength)), utils.Params.LogSlots())

			// Add all degree together
			result := utils.AddPlainNew(deg2, deg0)
			c <- &result

		}(input[i], t.U.CopyUtilsWithClonedEval(), outputChannels[i])
		
	}

	for i := range outputChannels{
		output[i] = <-outputChannels[i]
	}

	return output

}

func (t Tanh) GetForwardLevelConsumption() int {

	return 2

}

func (t Tanh) GetBackwardLevelConsumption() int {

	return 2

}

func (t Tanh) GetType() string {
	return "tanh"
}