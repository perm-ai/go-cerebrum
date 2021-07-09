package utility

import (
	"math"
	// "math/rand"
	// "strconv"
	"fmt"
	"testing"

	// "time"
	"github.com/ldsec/lattigo/v2/ckks"
	// "github.com/perm-ai/GO-HEML-prototype/src/logger"
)

func TestSigmoid(t *testing.T) {

	// Evaluate Sigmoid according to the approximation
	// 0.5 + 0.197x + 0.004x^3

	testCase := make([]ckks.Ciphertext, 2)
	data1 := make([]float64, utils.Params.Slots())
	for i := 0; i < int(utils.Params.Slots()); i++ {
		data1[i] = float64(1)
	}

	data2 := make([]float64, utils.Params.Slots())
	for i := 0; i < int(utils.Params.Slots()); i++ {
		data2[i] = float64(0.5)
	}

	testCase[0] = utils.Encrypt(data1)
	testCase[1] = utils.Encrypt(data2)
	x := testCase[0]
	y := testCase[1]

	output := utils.MultiplyNew(x, x, true, false)                                                                            // output = x * x
	output = utils.MultiplyNew(output, utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.004), true, false), true, false) // output = output * (x * 0.004)
	output = utils.AddNew(output, utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.197), true, false))                   // output = output + 0.197 * x
	output = utils.AddNew(output, y)                                                                                          // output = output + 0.5

	// output := utils.AddConstNew(&y, utils.GenerateFilledArray(0.5))

	ans := utils.Decrypt(&output)

	if EvalCorrectness(ans, utils.GenerateFilledArray(0.701), false, 2) {
		fmt.Println("The data was correctly evaluated")
	} else {
		fmt.Println("The data was not correctly evaluated")
	}

	// utils.Evaluator.EvaluatePoly()
}

func TestSigApprox(t *testing.T) {
	testCase := make([]ckks.Ciphertext, 1)
	data1 := make([]float64, utils.Params.Slots())
	for i := 0; i < int(utils.Params.Slots()); i++ {
		data1[i] = float64(1)
	}

	testCase[0] = utils.Encrypt(data1)

	x := testCase[0]

	output := utils.MultiplyNew(utils.MultiplyNew(x, utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.004), true, false), true, false), utils.MultiplyNew(x, x, true, false), true, false) // output = 0.004 * x^3
	// output = utils.MultiplyNew(output, utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.004), true, false), true, false) // output = output * (x * 0.004)
	output = utils.SubNew(utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.197), true, false), output) // output = output + 0.197 * x

	SigCont := utils.GenerateFilledArray(0.5)
	encoded := utils.EncodeToScale(SigCont, math.Pow(2.0, 20.0))
	utils.ReEncodeAsNTT(&encoded)

	output = utils.AddPlainNew(output, encoded) // output = output + 0.5

	ans := utils.Decrypt(&output)

	if EvalCorrectness(ans, utils.GenerateFilledArray(0.693), false, 10) {
		fmt.Println("The data was correctly evaluated")
	} else {
		fmt.Println("The data was not correctly evaluated")
	}

}
