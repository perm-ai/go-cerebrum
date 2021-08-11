package activations

import (
	"fmt"

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

func (s Sigmoid) Forward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// y := 0.5 + 0.197x + 0.004x^3
	fmt.Println("Using sigmoid function")
	// Calculate degree three
	fmt.Println("Degree 3")
	xSquared := s.U.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	fmt.Println("Degree 3.2")
	deg3 := s.U.MultiplyConstNew(input.CopyNew(), 0.004, true, false)
	fmt.Println("Degree 3.3")
	s.U.Multiply(xSquared, deg3, &deg3, true, false)

	// Calculate degree one
	fmt.Println("Degree 1")
	deg1 := s.U.MultiplyConstNew(input.CopyNew(), 0.197, true, false)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext
	fmt.Print("Test")
	if _, ok := s.forwardDeg0[inputLength]; ok {
		deg0 = s.forwardDeg0[inputLength]
		fmt.Print(" A \n")
	} else {
		deg0 = *s.U.Encoder.EncodeNTTNew(s.U.Float64ToComplex128(s.U.GenerateFilledArraySize(0.5, inputLength)), s.U.Params.LogSlots())
		fmt.Print(" B \n")
	}

	// Add all degree together
	fmt.Println("Getting result")
	result := s.U.AddNew(deg3, deg1)
	s.U.AddPlain(&result, &deg0, &result)
	fmt.Println("result level : " + fmt.Sprint(result.Level()))
	return result

}

func (s Sigmoid) Backward(input ckks.Ciphertext, inputLength int) ckks.Ciphertext {

	// 0.012x^2 + 0.197

	// Calculate degree three
	xSquared := s.U.MultiplyNew(*input.CopyNew(), *input.CopyNew(), true, false)
	deg2 := s.U.MultiplyConstNew(&xSquared, 0.012, true, false)

	// Encode deg0 as plaintext
	var deg0 ckks.Plaintext

	if _, ok := s.backwardDeg0[inputLength]; ok {
		deg0 = s.backwardDeg0[inputLength]
	} else {
		deg0 = *s.U.Encoder.EncodeNTTNew(s.U.Float64ToComplex128(s.U.GenerateFilledArraySize(0.197, inputLength)), s.U.Params.LogSlots())
	}

	// Add all degree together
	result := s.U.AddPlainNew(deg2, deg0)

	return result

}
