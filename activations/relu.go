package activations

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type Relu struct {
	U utility.Utils
}

func (r Relu) Forward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	// implement relu approximation according to equation (-1/120)x^4 + (5/24)x^2 + (1/2)x + 0.3
	output := make([]*ckks.Ciphertext, len(input))
	outputChannels := make([]chan *ckks.Ciphertext, len(input))

	for i := range input {

		outputChannels[i] = make(chan *ckks.Ciphertext)

		go func(inputEach *ckks.Ciphertext, utils utility.Utils, c chan *ckks.Ciphertext) {

			// calculate degree 4
			xSquared := utils.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
			xForthed := utils.MultiplyNew(xSquared, xSquared, true, false)
			deg4coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize((-1.0 / 120.0), inputLength))
			deg4 := utils.MultiplyPlainNew(&xForthed, &deg4coeff, true, false)

			// calculate degree 2
			deg2coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize((5.0 / 24.0), inputLength))
			deg2 := utils.MultiplyPlainNew(&xSquared, &deg2coeff, true, false)

			// calculate degree 1
			deg1coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize((0.5), inputLength))
			deg1 := utils.MultiplyPlainNew(inputEach.CopyNew(), &deg1coeff, true, false)

			// calculate degree 0
			deg0coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize((0.3), inputLength))

			// put everything together
			result1 := utils.AddNew(deg4, deg2)
			result2 := utils.AddNew(deg1, result1)
			result3 := utils.AddPlainNew(result2, deg0coeff)
			c <- &result3

		}(input[i], r.U.CopyWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels {
		output[i] = <-outputChannels[i]
	}

	return output

}

func (r Relu) Backward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	// (-1/30)x^3 + (10/24)x + 0.5

	output := make([]*ckks.Ciphertext, len(input))
	outputChannels := make([]chan *ckks.Ciphertext, len(input))

	for i := range input {

		outputChannels[i] = make(chan *ckks.Ciphertext)

		go func(inputEach *ckks.Ciphertext, utils utility.Utils, c chan *ckks.Ciphertext) {

			//calculate deg3
			xSquared := utils.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
			deg3coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize((-4.0 / 120.0), inputLength))
			xAndDeg3coeff := utils.MultiplyPlainNew(inputEach.CopyNew(), &deg3coeff, true, false)
			deg3 := utils.MultiplyNew(xSquared, xAndDeg3coeff, true, false)

			//calculate deg1

			deg1coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize((10.0 / 24.0), inputLength))
			deg1 := utils.MultiplyPlainNew(inputEach.CopyNew(), &deg1coeff, true, false)

			//calculate deg0

			deg0coeff := utils.EncodePlaintextFromArray(utils.GenerateFilledArraySize((0.5), inputLength))

			//add everything together

			result1 := utils.AddNew(deg3, deg1)
			result2 := utils.AddPlainNew(result1, deg0coeff)
			c <- &result2

		}(input[i], r.U.CopyWithClonedEval(), outputChannels[i])

	}
	return output
}

func (r Relu) GetForwardLevelConsumption() int {
	return 3
}

func (r Relu) GetBackwardLevelConsumption() int {
	return 2
}

func (r Relu) GetType() string {
	return "relu"
}
