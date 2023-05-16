package activations

import (
	"fmt"
	"math"
	"testing"

	"github.com/tuneinsight/lattigo/v4/rlwe" 
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

var log = logger.NewLogger(true)
var keyChain = key.GenerateKeys(0, false, true)
var utils = utility.NewUtils(keyChain, math.Pow(2, 40), 100, true)

func TestSigmoid(t *testing.T) {

	// y = 0.5 + 0.197x + 0.004x^3
	// 0.012x^2 + 0.197

	inputArrray := utils.GenerateRandomNormalArray(100,1)
	forwardExpected := make([]float64, utils.Params.Slots())
	backwardExpected := make([]float64, utils.Params.Slots())

	for i := 0; i < 100; i++ {
		forwardExpected[i] = 0.5 + (0.197 * inputArrray[i]) - (0.004 * math.Pow(inputArrray[i], 3))
		backwardExpected[i] = (0.012 * math.Pow(inputArrray[i], 2)) + 0.197
	}

	encryptedInput := utils.Encrypt(inputArrray)

	sigmoid := Sigmoid{U: utils}

	fwdResult := sigmoid.Forward([]*rlwe.Ciphertext{&encryptedInput}, 100)

	if !utility.ValidateResult(utils.Decrypt(fwdResult[0]), forwardExpected, false, 1, log) {
		t.Error("Sigmoid forward wasn't evaluated properly")
	}

	fmt.Println("Starting backward")

	backwardResult := sigmoid.Backward([]*rlwe.Ciphertext{&encryptedInput}, 100)

	if !utility.ValidateResult(utils.Decrypt(backwardResult[0]), backwardExpected, false, 1, log) {
		t.Error("Sigmoid backward wasn't evaluated properly")
	}

}

func TestTanh(t *testing.T) {

	// y = (-0.00752x^3) + (0.37x)
	// (-0.02256x^2) + 0.37

	testCases := make([]*rlwe.Ciphertext, 10)
	forwardExpected := make([][]float64, 10)
	backwardExpected := make([][]float64, 10)

	for testCase := 0; testCase < 10; testCase++{
		inputArrray := utils.GenerateRandomNormalArray(100,1)
		forwardExpected[testCase] = make([]float64, utils.Params.Slots())
		backwardExpected[testCase] = make([]float64, utils.Params.Slots())

		for i := 0; i < 100; i++ {
			forwardExpected[testCase][i] = (0.37 * inputArrray[i]) + (-0.00752 * math.Pow(inputArrray[i], 3))
			backwardExpected[testCase][i] = (-0.02256 * math.Pow(inputArrray[i], 2)) + 0.37
		}

		testCases[testCase] = utils.EncryptToPointer(inputArrray)
	}

	tanh := Tanh{U: utils}

	fwdResult := tanh.Forward(testCases, 100)
	backwardResult := tanh.Backward(testCases, 100)

	for testCase := range testCases {
		if !utility.ValidateResult(utils.Decrypt(fwdResult[testCase]), forwardExpected[testCase], false, 1, log) {
			t.Error(fmt.Sprintf("Tanh forward wasn't evaluated properly (testcase %d)", testCase))
		}
	
		if !utility.ValidateResult(utils.Decrypt(backwardResult[testCase]), backwardExpected[testCase], false, 1, log) {
			t.Error(fmt.Sprintf("Tanh backward wasn't evaluated properly (testcase %d)", testCase))
		}
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
	fmt.Printf("Scale: %f\n", encInput.Scale.Float64())
	startingLevel := encInput.Level()

	// Softmax initiation
	softmax := NewSoftmax(utils)

	// Calculate softmax forward
	result := softmax.Forward([]*rlwe.Ciphertext{&encInput}, 10)[0]
	decResult := utils.Decrypt(result)

	fmt.Printf("Used: %d levels (%d - %d)\n", startingLevel-result.Level(), startingLevel, result.Level())

	// Evaluate correctness with precision 0.5
	if !utility.ValidateResult(decResult[0:10], expected[0:10], false, -0.3, log) {
		t.Error("Softmax forward wasn't evaluated properly")
	}

}
