package utility

import (
	"fmt"
	"testing"

	"github.com/ldsec/lattigo/v2/ckks"
)

func addArrays(a []float64, b []float64) []float64 {

	for i := range a {
		a[i] = a[i] + b[i]
	}

	return a

}

func multiplyArrays(a []float64, b []float64) []float64 {

	for i := range a {
		a[i] = a[i] * b[i]
	}

	return a

}

func TestBootstrappingConsistency(t *testing.T) {

	log.Log("Testing bootstrapping consistency")

	ciphertext := utils.Encrypt(utils.GenerateFilledArraySize(500, 100))

	randomMulArrays := make([][]float64, 300)
	randomAddArrays := make([][]float64, 300)
	randomEncodedMulArrays := make([]ckks.Plaintext, 300)
	randomEncodedAddArrays := make([]ckks.Plaintext, 300)

	var plainResult []float64

	bootstrapCount := 0
	mulCount := 0
	addCount := 0

	for i := range randomMulArrays {
		randomMulArrays[i] = utils.GenerateRandomArray(0, 2, 100)
		randomEncodedMulArrays[i] = *utils.Encoder.EncodeNTTAtLvlNew(utils.Params.MaxLevel(), utils.Float64ToComplex128(randomMulArrays[i]), utils.Params.LogSlots())

		randomAddArrays[i] = utils.GenerateRandomArray(-1, 1, 100)
		randomEncodedAddArrays[i] = *utils.Encoder.EncodeNTTAtLvlNew(utils.Params.MaxLevel(), utils.Float64ToComplex128(randomAddArrays[i]), utils.Params.LogSlots())

		if i == 0 {
			plainResult = utils.GenerateFilledArraySize(500, 100)
		}

		log.Log("Testing Iteration " + fmt.Sprintf("%d / %d", i, len(randomAddArrays)))

		mulCount++
		log.Log(fmt.Sprintf("Multiplying (%d)", mulCount))
		plainResult = multiplyArrays(plainResult, randomMulArrays[i])
		utils.MultiplyPlain(&ciphertext, &randomEncodedMulArrays[i], &ciphertext, true, false)
		if !EvalCorrectness(utils.Decrypt(&ciphertext)[0:100], plainResult, false, 1) {
			t.Error("Multiplication was incorrectly evaluated")
		}

		addCount++
		log.Log("Adding " + fmt.Sprintf("(%d)", mulCount))
		plainResult = addArrays(plainResult, randomAddArrays[i])
		utils.Evaluator.Add(&ciphertext, &randomEncodedAddArrays[i], &ciphertext)
		if !EvalCorrectness(utils.Decrypt(&ciphertext)[0:100], plainResult, false, 1) {
			t.Error("Addition was incorrectly evaluated")
		}

		if ciphertext.Level() == 1 {
			bootstrapCount++
			log.Log(fmt.Sprintf("Bootstrapping (%d) Scale: %f, Level: %d", bootstrapCount, ciphertext.Scale(), ciphertext.Level()))
			utils.BootstrapInPlace(&ciphertext)
			log.Log(fmt.Sprintf("Post Bootstrapping (%d) Scale: %f, Level: %d", bootstrapCount, ciphertext.Scale(), ciphertext.Level()))
			if !EvalCorrectness(utils.Decrypt(&ciphertext)[0:100], plainResult, false, 1) {
				t.Error("Bootstrap changed the value in the ciphertext")
			}
		}

		fmt.Println()

	}

}
