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
	forwardDeg0  map[int]ckks.Plaintext
	backwardDeg0 map[int]ckks.Plaintext
}

func (s Sigmoid) Forward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {
	output := make([]*ckks.Ciphertext, len(input))
	for i, inputEach := range input {

		// y := 0.5 + 0.197x - 0.004x^3
		// Calculate degree three
		xSquared := s.U.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
		deg3Coeff := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(-0.004, inputLength))
		deg3 := s.U.MultiplyPlainNew(inputEach.CopyNew(), &deg3Coeff, true, false)
		s.U.Multiply(xSquared, deg3, &deg3, true, false)

		// Calculate degree one
		deg1Coeff := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(0.197, inputLength))
		deg1 := s.U.MultiplyPlainNew(inputEach.CopyNew(), &deg1Coeff, true, false)

		// Encode deg0 as plaintext
		var deg0 ckks.Plaintext
		if _, ok := s.forwardDeg0[inputLength]; ok {
			deg0 = s.forwardDeg0[inputLength]
		} else {
			deg0 = *s.U.Encoder.EncodeNTTNew(s.U.Float64ToComplex128(s.U.GenerateFilledArraySize(0.5, inputLength)), s.U.Params.LogSlots())
		}

		// Add all degree together
		result := s.U.AddNew(deg3, deg1)
		s.U.AddPlain(&result, &deg0, &result)
		output[i] = &result
	}
	return output

}

func (s Sigmoid) Backward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext {
	output := make([]*ckks.Ciphertext, len(input))
	for i, inputEach := range input {
		// 0.012x^2 + 0.197

		// Calculate degree three
		xSquared := s.U.MultiplyNew(*inputEach.CopyNew(), *inputEach.CopyNew(), true, false)
		deg2Coeff := s.U.EncodePlaintextFromArray(s.U.GenerateFilledArraySize(0.012, inputLength))
		deg2 := s.U.MultiplyPlainNew(&xSquared, &deg2Coeff, true, false)

		// Encode deg0 as plaintext
		var deg0 ckks.Plaintext

		if _, ok := s.backwardDeg0[inputLength]; ok {
			deg0 = s.backwardDeg0[inputLength]
		} else {
			deg0 = *s.U.Encoder.EncodeNTTNew(s.U.Float64ToComplex128(s.U.GenerateFilledArraySize(0.197, inputLength)), s.U.Params.LogSlots())
		}

		// Add all degree together
		result := s.U.AddPlainNew(deg2, deg0)
		output[i] = &result
	}
	return output

}

func (s Sigmoid) GetForwardLevelConsumption() int {
	return 2
}

func (s Sigmoid) GetBackwardLevelConsumption() int {
	return 2
}
