package utility

import (
	// "math"
	// "math/rand"
	// "strconv"
	"testing"
	// "time"
	// "github.com/ldsec/lattigo/v2/ckks"
	// "github.com/perm-ai/GO-HEML-prototype/src/logger"
)

func TestSigmoid(t *testing.T) {

	// Evaluate Sigmoid according to the approximation
	// 0.5 + 0.197x + 0.004x^3

	testCases := GenerateTestCases(utils)
	x := testCases[0].data1

	output := utils.MultiplyNew(x, x, true, false)                                                          // output = x * x
	output = utils.AddNew(output, utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.004), true, false)) // output = output + (x * 0.004)
	output = utils.AddNew(output, utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.197), true, false)) // output = output + 0.197 * x
	utils.AddConst(&output, utils.GenerateFilledArray(0.5), &output)                                        // output = output + 0.5
	utils.Decrypt(&output)

}
