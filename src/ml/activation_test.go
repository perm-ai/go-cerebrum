package ml

import (
	"math"
	"testing"

	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)


func TestSigmoid(t *testing.T){

	// y = 0.5 + 0.197x + 0.004x^3
	// 0.012x^2 + 0.197

	inputArrray := utils.GenerateRandomNormalArray(100)
	forwardExpected := make([]float64, utils.Params.Slots())
	backwardExpected := make([]float64, utils.Params.Slots())

	for i := 0; i < 100; i++{
		forwardExpected[i] = 0.5 + (0.197 * inputArrray[i]) + (0.004 * math.Pow(inputArrray[i], 3))
		backwardExpected[i] = (0.012 * math.Pow(inputArrray[i], 2)) + 0.197
	}

	encryptedInput := utils.Encrypt(inputArrray)

	sigmoid := Sigmoid{utils: utils}

	fwdResult := sigmoid.Forward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&fwdResult), forwardExpected, false, 1, log){
		t.Error("Sigmoid forward wasn't evaluated properly")
	}

	backwardResult := sigmoid.Backward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&backwardResult), backwardExpected, false, 1, log){
		t.Error("Sigmoid backward wasn't evaluated properly")
	}

}

func TestTanh(t *testing.T){

	// y = (-0.00752x^3) + (0.37x)
	// (-0.02256x^2) + 0.37

	inputArrray := utils.GenerateRandomNormalArray(100)
	forwardExpected := make([]float64, utils.Params.Slots())
	backwardExpected := make([]float64, utils.Params.Slots())

	for i := 0; i < 100; i++{
		forwardExpected[i] = (0.37 * inputArrray[i]) + (-0.00752 * math.Pow(inputArrray[i], 3))
		backwardExpected[i] = (-0.02256 * math.Pow(inputArrray[i], 2)) + 0.37
	}

	encryptedInput := utils.Encrypt(inputArrray)

	tanh := Tanh{utils: utils}

	fwdResult := tanh.Forward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&fwdResult), forwardExpected, false, 1, log){
		t.Error("Tanh forward wasn't evaluated properly")
	}

	backwardResult := tanh.Backward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&backwardResult), backwardExpected, false, 1, log){
		t.Error("Tanh backward wasn't evaluated properly")
	}

}