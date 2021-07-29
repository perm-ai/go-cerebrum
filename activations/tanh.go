package activations

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

//=================================================
//					TANH
//=================================================

type Tanh struct {
	U        		utility.Utils
	backwardDeg0	map[int]ckks.Plaintext
}

func (t Tanh) Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// y = (-0.00752x^3) + (0.37x)

	// Calculate degree three
	xSquared := t.U.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg3 := t.U.MultiplyConstNew(input.CopyNew(), -0.00752, true, false)
	t.U.Multiply(xSquared, deg3, &deg3, true, false)

	// Calculate degree one
	deg1 := t.U.MultiplyConstNew(input.CopyNew(), 0.37, true, false)

	// Add all degree together
	result := t.U.AddNew(deg3, deg1)

	return result

}

func (t Tanh) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// (-0.02256x^2) + 0.37

	// Calculate degree three
	xSquared := t.U.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg2 := t.U.MultiplyConstNew(&xSquared, -0.02256, true, false)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := t.backwardDeg0[inputLength]; ok {
		deg0 = t.backwardDeg0[inputLength]
	} else {
		deg0 = *t.U.Encoder.EncodeNTTNew(t.U.Float64ToComplex128(t.U.GenerateFilledArraySize(0.37, inputLength)), t.U.Params.LogSlots())
	}

	// Add all degree together
	result := t.U.AddPlainNew(deg2, deg0)

	return result

}