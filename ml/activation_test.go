package ml

import (
	"fmt"
	"math"
	"testing"

	"github.com/perm-ai/go-cerebrum/utility"
)

func TestSigmoid(t *testing.T) {

	// y = 0.5 + 0.197x + 0.004x^3
	// 0.012x^2 + 0.197

	inputArrray := utils.GenerateRandomNormalArray(100)
	forwardExpected := make([]float64, utils.Params.Slots())
	backwardExpected := make([]float64, utils.Params.Slots())

	for i := 0; i < 100; i++ {
		forwardExpected[i] = 0.5 + (0.197 * inputArrray[i]) + (0.004 * math.Pow(inputArrray[i], 3))
		backwardExpected[i] = (0.012 * math.Pow(inputArrray[i], 2)) + 0.197
	}

	encryptedInput := utils.Encrypt(inputArrray)

	sigmoid := Sigmoid{utils: utils}

	fwdResult := sigmoid.Forward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&fwdResult), forwardExpected, false, 1, log) {
		t.Error("Sigmoid forward wasn't evaluated properly")
	}

	backwardResult := sigmoid.Backward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&backwardResult), backwardExpected, false, 1, log) {
		t.Error("Sigmoid backward wasn't evaluated properly")
	}

}

func TestTanh(t *testing.T) {

	// y = (-0.00752x^3) + (0.37x)
	// (-0.02256x^2) + 0.37

	inputArrray := utils.GenerateRandomNormalArray(100)
	forwardExpected := make([]float64, utils.Params.Slots())
	backwardExpected := make([]float64, utils.Params.Slots())

	for i := 0; i < 100; i++ {
		forwardExpected[i] = (0.37 * inputArrray[i]) + (-0.00752 * math.Pow(inputArrray[i], 3))
		backwardExpected[i] = (-0.02256 * math.Pow(inputArrray[i], 2)) + 0.37
	}

	encryptedInput := utils.Encrypt(inputArrray)

	tanh := Tanh{utils: utils}

	fwdResult := tanh.Forward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&fwdResult), forwardExpected, false, 1, log) {
		t.Error("Tanh forward wasn't evaluated properly")
	}

	backwardResult := tanh.Backward(encryptedInput, 100)

	if !utility.ValidateResult(utils.Decrypt(&backwardResult), backwardExpected, false, 1, log) {
		t.Error("Tanh backward wasn't evaluated properly")
	}

}

func TestSoftmax(t *testing.T) {

	// Generate test cases
	randomArr := []float64{2.0805165053225812954, 0.565725005884827212, 0.1697818800998136352, -0.9019945467371572339, -0.09284242516708391313, 0.74402791096417377295, 1.16814415104123592, 1.398561882988180513, 1.6169174606244177577, 0.5570947367915441425}
	expArr := make([]float64, len(randomArr))
	expected := make([]float64, len(randomArr))

	sum := float64(0)

	for i, n := range randomArr {
		expArr[i] = math.Exp(n)
	}

	for _, n := range expArr {
		sum += n
	}

	for i, n := range expArr {
		expected[i] = n / sum
	}

	// Encryption
	encInput := utils.Encrypt(randomArr)
	utils.Evaluator.DropLevel(&encInput, encInput.Level()-9)
	startingLevel := encInput.Level()

	// Softmax initiation
	softmax := NewSoftmax(utils)

	// Calculate softmax forward
	result := softmax.Forward(encInput, 10)

	fmt.Printf("Used: %d levels (%d - %d)\n", startingLevel-result.Level(), startingLevel, result.Level())

	// Evaluate correctness with precision 0.5
	if !utility.ValidateResult(utils.Decrypt(&result)[0:10], expected[0:10], false, -0.3, log) {
		t.Error("Softmax forward wasn't evaluated properly")
	}

}
