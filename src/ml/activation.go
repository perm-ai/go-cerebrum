package ml

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)


type Activation interface {

	forward(input ckks.Ciphertext, inputLength int) 	ckks.Ciphertext
	backward(input ckks.Ciphertext, inputLength int)	ckks.Ciphertext

}

//=================================================
//					SIGMOID
//=================================================

type Sigmoid struct {
	utils			utility.Utils
	forwardDeg0		map[int]ckks.Plaintext
	backwardDeg0	map[int]ckks.Plaintext
}

func (s Sigmoid) forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// y := 0.5 + 0.197x + 0.004x^3
	
	// Calculate degree three
	xSquared := s.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg3 := s.utils.Evaluator.MultByConstNew(input.CopyNew(), 0.004)
	s.utils.Multiply(xSquared, *deg3, deg3, true, false)

	// Calculate degree one
	deg1 := s.utils.Evaluator.MultByConstNew(input.CopyNew(), 0.197)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := s.forwardDeg0[inputLength]; ok {
		deg0 = s.forwardDeg0[inputLength]
	} else {
		deg0 = *s.utils.Encoder.EncodeNTTNew(s.utils.Float64ToComplex128(s.utils.GenerateFilledArraySize(0.5, inputLength)), s.utils.Params.LogSlots())
	}

	// Add all degree together
	result := s.utils.AddNew(*deg3, *deg1)
	s.utils.AddPlain(&result, &deg0, &result)

	return result

}

func (s Sigmoid) backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// 0.012x^2 + 0.197
	
	// Calculate degree three
	xSquared := s.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg2 := s.utils.Evaluator.MultByConstNew(&xSquared, 0.012)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := s.backwardDeg0[inputLength]; ok {
		deg0 = s.backwardDeg0[inputLength]
	} else {
		deg0 = *s.utils.Encoder.EncodeNTTNew(s.utils.Float64ToComplex128(s.utils.GenerateFilledArraySize(0.197, inputLength)), s.utils.Params.LogSlots())
	}

	// Add all degree together
	result := s.utils.AddPlainNew(*deg2, deg0)

	return result

}

//=================================================
//					SIGMOID
//=================================================

type Tanh struct {
	utils			utility.Utils
	backwardDeg0	map[int]ckks.Plaintext
}

func (t Tanh) forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// y = (-0.00752x^3) + (0.37x)
	
	// Calculate degree three
	xSquared := t.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg3 := t.utils.Evaluator.MultByConstNew(input.CopyNew(), -0.00752)
	t.utils.Multiply(xSquared, *deg3, deg3, true, false)

	// Calculate degree one
	deg1 := t.utils.Evaluator.MultByConstNew(input.CopyNew(), 0.37)

	// Add all degree together
	result := t.utils.AddNew(*deg3, *deg1)

	return result

}

func (t Tanh) backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// (-0.02256x^2) + 0.37
	
	// Calculate degree three
	xSquared := t.utils.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg2 := t.utils.Evaluator.MultByConstNew(&xSquared, -0.02256)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := t.backwardDeg0[inputLength]; ok {
		deg0 = t.backwardDeg0[inputLength]
	} else {
		deg0 = *t.utils.Encoder.EncodeNTTNew(t.utils.Float64ToComplex128(t.utils.GenerateFilledArraySize(0.37, inputLength)), t.utils.Params.LogSlots())
	}

	// Add all degree together
	result := t.utils.AddPlainNew(*deg2, deg0)

	return result

}