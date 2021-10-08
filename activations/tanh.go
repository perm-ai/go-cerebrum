package activations

import (
	"math"
	"sync"

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

	deg3Coeff := t.U.EncodePlaintextFromArray(t.U.GenerateFilledArraySize(-0.00752, inputLength))
	deg1Coeff := t.U.EncodePlaintextFromArray(t.U.GenerateFilledArraySize(0.37, inputLength))

	var wg sync.WaitGroup

	for i := range input {

		wg.Add(1)

		go func(index int, utils utility.Utils) {

			defer wg.Done()

			// Calculate degree three
			xSquared := utils.MultiplyNew(input[index], input[index], true, false)
			output[index] = utils.MultiplyPlainNew(input[index], deg3Coeff, true, false)
			utils.Multiply(xSquared, output[index], output[index], true, false)

			// Calculate degree one
			deg1 := utils.MultiplyPlainNew(input[index], deg1Coeff, true, false)

			// Add all degree together
			utils.Add(output[index], deg1, output[index])

		}(i, t.U.CopyWithClonedEval())

	}

	wg.Wait()

	return output

}

func (t Tanh) Backward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {

	// (-0.02256x^2) + 0.37

	output := make([]*ckks.Ciphertext, len(input))
	outputChannels := make([]chan *ckks.Ciphertext, len(input))

	deg2Coeff := t.U.EncodePlaintextFromArrayScale(t.U.GenerateFilledArray(-0.02256), math.Pow(2, 30))
	deg0 := t.U.Encoder.EncodeNTTNew(t.U.Float64ToComplex128(t.U.GenerateFilledArraySize(0.37, inputLength)), t.U.Params.LogSlots())

	for i := range input {

		outputChannels[i] = make(chan *ckks.Ciphertext)

		go func(index int, utils utility.Utils, c chan *ckks.Ciphertext) {

			// Calculate degree 2
			result := utils.MultiplyNew(input[index], input[index], true, false)
			utils.MultiplyPlain(result, deg2Coeff, result, true, false)

			// Add all degree together
			utils.AddPlain(result, deg0, result)
			c <- result

		}(i, t.U.CopyWithClonedEval(), outputChannels[i])

	}

	for i := range outputChannels {
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
