package ml

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

type Activation interface {
	Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext
	Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext
}

//=================================================
//					SIGMOID
//=================================================

type Sigmoid struct {
	utils        utility.Utils
	forwardDeg0  map[int]ckks.Plaintext
	backwardDeg0 map[int]ckks.Plaintext
}

func (s Sigmoid) Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// y := 0.5 + 0.197x + 0.004x^3

	// Calculate degree three
	xSquared := s.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg3 := s.utils.MultiplyConstNew(input.CopyNew(), 0.004, true, false)
	s.utils.Multiply(xSquared, deg3, &deg3, true, false)

	// Calculate degree one
	deg1 := s.utils.MultiplyConstNew(input.CopyNew(), 0.197, true, false)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := s.forwardDeg0[inputLength]; ok {
		deg0 = s.forwardDeg0[inputLength]
	} else {
		deg0 = *s.utils.Encoder.EncodeNTTNew(s.utils.Float64ToComplex128(s.utils.GenerateFilledArraySize(0.5, inputLength)), s.utils.Params.LogSlots())
	}

	// Add all degree together
	result := s.utils.AddNew(deg3, deg1)
	s.utils.AddPlain(&result, &deg0, &result)

	return result

}

func (s Sigmoid) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// 0.012x^2 + 0.197

	// Calculate degree three
	xSquared := s.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg2 := s.utils.MultiplyConstNew(&xSquared, 0.012, true, false)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := s.backwardDeg0[inputLength]; ok {
		deg0 = s.backwardDeg0[inputLength]
	} else {
		deg0 = *s.utils.Encoder.EncodeNTTNew(s.utils.Float64ToComplex128(s.utils.GenerateFilledArraySize(0.197, inputLength)), s.utils.Params.LogSlots())
	}

	// Add all degree together
	result := s.utils.AddPlainNew(deg2, deg0)

	return result

}

//=================================================
//					TANH
//=================================================

type Tanh struct {
	utils        utility.Utils
	backwardDeg0 map[int]ckks.Plaintext
}

func (t Tanh) Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// y = (-0.00752x^3) + (0.37x)

	// Calculate degree three
	xSquared := t.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg3 := t.utils.MultiplyConstNew(input.CopyNew(), -0.00752, true, false)
	t.utils.Multiply(xSquared, deg3, &deg3, true, false)

	// Calculate degree one
	deg1 := t.utils.MultiplyConstNew(input.CopyNew(), 0.37, true, false)

	// Add all degree together
	result := t.utils.AddNew(deg3, deg1)

	return result

}

func (t Tanh) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// (-0.02256x^2) + 0.37

	// Calculate degree three
	xSquared := t.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg2 := t.utils.MultiplyConstNew(&xSquared, -0.02256, true, false)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := t.backwardDeg0[inputLength]; ok {
		deg0 = t.backwardDeg0[inputLength]
	} else {
		deg0 = *t.utils.Encoder.EncodeNTTNew(t.utils.Float64ToComplex128(t.utils.GenerateFilledArraySize(0.37, inputLength)), t.utils.Params.LogSlots())
	}

	// Add all degree together
	result := t.utils.AddPlainNew(deg2, deg0)

	return result

}

//=================================================
//					SOFTMAX
//=================================================

type Softmax struct {
	utils utility.Utils
}

func (s Softmax) Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// TODO: Implement Homomorphic Encryption frienly version of sigmoid
	return input

}

func (s Softmax) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// TODO: Implement Homomorphic Encryption frienly version of sigmoid
	return input

}
