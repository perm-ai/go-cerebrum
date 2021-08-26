package activations

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type Relu struct {
	U utility.Utils
}

func (r Relu) Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// implement relu approximation according to equation (-1/120)x^4 + (5/24)x^2 + (1/2)x + 0.3

	// calculate degree 4
	xSquared := r.U.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	xForthed := r.U.MultiplyNew(xSquared, xSquared, true, false)
	deg4coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((-1.0 / 120.0), inputLength))
	deg4 := r.U.MultiplyPlainNew(&xForthed, &deg4coeff, true, false)

	// calculate degree 2
	deg2coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((5.0 / 24.0), inputLength))
	deg2 := r.U.MultiplyPlainNew(&xSquared, &deg2coeff, true, false)

	// calculate degree 1
	deg1coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((0.5), inputLength))
	deg1 := r.U.MultiplyPlainNew(input.CopyNew(), &deg1coeff, true, false)

	// calculate degree 0
	deg0coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((0.3), inputLength))

	// put everything together
	result1 := r.U.AddNew(deg4, deg2)
	result2 := r.U.AddNew(deg1, result1)
	result3 := r.U.AddPlainNew(result2, deg0coeff)

	return result3

}

func (r Relu) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {
	// (-1/30)x^3 + (10/24)x + 0.5

	//calculate deg3
	xSquared := r.U.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg3coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((-4.0 / 120.0), inputLength))
	xAndDeg3coeff := r.U.MultiplyPlainNew(input.CopyNew(), &deg3coeff, true, false)
	deg3 := r.U.MultiplyNew(xSquared, xAndDeg3coeff, true, false)

	//calculate deg1

	deg1coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((10.0 / 24.0), inputLength))
	deg1 := r.U.MultiplyPlainNew(input.CopyNew(), &deg1coeff, true, false)

	//calculate deg0

	deg0coeff := r.U.EncodePlaintextFromArray(r.U.GenerateFilledArraySize((0.5), inputLength))

	//add everything together

	result1 := r.U.AddNew(deg3, deg1)
	result2 := r.U.AddPlainNew(result1, deg0coeff)

	return result2
}

func (r Relu) GetForwardLevelConsumption() int {
	return 3
}

func (r Relu) GetBackwardLevelConsumption() int {
	return 2
}