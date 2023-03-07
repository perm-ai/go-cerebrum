package activations

import (
	"sync"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

type Relu struct {
	U utility.Utils
}

func (r Relu) Forward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	// implement relu approximation according to equation (-1/120)x^4 + (5/24)x^2 + (1/2)x + 0.3
	output := make([]*rlwe.Ciphertext, len(input))
	outputChannels := make([]chan *rlwe.Ciphertext, len(input))

	deg4coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((-1.0 / 120.0), inputLength))
	deg2coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((5.0 / 24.0), inputLength))
	deg1coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((0.5), inputLength))
	deg0coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((0.3), inputLength))

	for i := range input {

		outputChannels[i] = make(chan *rlwe.Ciphertext)

		go func(inputEach *rlwe.Ciphertext, utils utility.Utils, c chan *rlwe.Ciphertext) {

			// calculate degree 4
			xSquared := utils.MultiplyNew(inputEach.CopyNew(), inputEach.CopyNew(), true, false)
			xForthed := utils.MultiplyNew(xSquared, xSquared, true, false)
			deg4 := utils.MultiplyPlainNew(xForthed, deg4coeff, true, false)

			// calculate degree 2
			deg2 := utils.MultiplyPlainNew(xSquared, deg2coeff, true, false)

			// calculate degree 1
			deg1 := utils.MultiplyPlainNew(inputEach.CopyNew(), deg1coeff, true, false)

			// put everything together
			result1 := utils.AddNew(deg4, deg2)
			result2 := utils.AddNew(deg1, result1)
			result3 := utils.AddPlainNew(result2, deg0coeff)
			c <- result3

		}(input[i], r.U.CopyWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels {
		output[i] = <-outputChannels[i]
	}

	return output

}

func (r Relu) Backward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	// (-1/30)x^3 + (10/24)x + 0.5

	output := make([]*rlwe.Ciphertext, len(input))

	deg3coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((-4.0 / 120.0), inputLength))
	deg1coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((10.0 / 24.0), inputLength))
	deg0coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((0.5), inputLength))

	var wg sync.WaitGroup

	for i := range input {

		wg.Add(1)

		go func(inputEach *rlwe.Ciphertext, utils utility.Utils, index int) {

			defer wg.Done()

			//calculate deg3
			xSquared := utils.MultiplyNew(inputEach.CopyNew(), inputEach.CopyNew(), true, false)
			xAndDeg3coeff := utils.MultiplyPlainNew(inputEach.CopyNew(), deg3coeff, true, false)
			deg3 := utils.MultiplyNew(xSquared, xAndDeg3coeff, true, false)

			//calculate deg1
			deg1 := utils.MultiplyPlainNew(inputEach.CopyNew(), deg1coeff, true, false)

			//add everything together

			result1 := utils.AddNew(deg3, deg1)
			output[index] = utils.AddPlainNew(result1, deg0coeff)

		}(input[i], r.U.CopyWithClonedEval(), i)

	}

	wg.Wait()

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
