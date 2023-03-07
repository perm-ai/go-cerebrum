package activations

import (
	"math"
	"sync"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					TANH
//=================================================

type Tanh struct {
	U            utility.Utils
	backwardDeg0 map[int]rlwe.Plaintext
}

func NewTanh(utils utility.Utils) Tanh {

	return Tanh{utils, make(map[int]rlwe.Plaintext)}

}

func (t Tanh) Forward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	// y = (-0.00752x^3) + (0.37x)

	output := make([]*rlwe.Ciphertext, len(input))

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

func (t Tanh) Backward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext {

	// (-0.02256x^2) + 0.37

	output := make([]*rlwe.Ciphertext, len(input))
	outputChannels := make([]chan *rlwe.Ciphertext, len(input))

	deg2Coeff := t.U.EncodePlaintextFromArrayScale(t.U.GenerateFilledArraySize(-0.02256, inputLength), math.Pow(2, 30))
	deg0 := t.U.Encoder.EncodeNew(t.U.Float64ToComplex128(t.U.GenerateFilledArraySize(0.37, inputLength)), t.U.Params.MaxLevel(), t.U.Params.NewScale(t.U.Scale), t.U.Params.LogSlots())

	for i := range input {

		outputChannels[i] = make(chan *rlwe.Ciphertext)

		go func(index int, utils utility.Utils, c chan *rlwe.Ciphertext) {

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
