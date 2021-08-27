package regression

import (
	"fmt"
	"strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

type LinearRegression struct {
	utils  utility.Utils
	Weight []ckks.Ciphertext
	Bias   ckks.Ciphertext
}

type LinearRegressionGradient struct {
	DM []ckks.Ciphertext
	DB ckks.Ciphertext
}

// need to pass in number of independent features
func NewLinearRegression(u utility.Utils, numOfFeatures int) LinearRegression {

	zeros := u.GenerateFilledArray(0.0)
	m := make([]ckks.Ciphertext, numOfFeatures)
	for i := 0; i < numOfFeatures; i++ {
		m[i] = u.Encrypt(zeros)
	}
	b := u.Encrypt(zeros)

	return LinearRegression{u, m, b}

}

func (l LinearRegression) Forward(input []ckks.Ciphertext) ckks.Ciphertext {

	// prepare result ciphertesxt

	result := l.utils.Encrypt(l.utils.GenerateFilledArray(0.0))

	// W*X for each feature, add sum in result

	for i := range input {
		dot := l.utils.MultiplyNew(input[i], l.Weight[i], true, false)
		l.utils.Add(result, dot, &result)
	}

	l.utils.Add(result, l.Bias, &result)

	return result

}

func (l LinearRegression) Backward(input []ckks.Ciphertext, output ckks.Ciphertext, y *ckks.Ciphertext, size int, learningRate float64) LinearRegressionGradient {

	// Calculate backward gradient using the following equation
	// dM = (-2/n) * sum(input * (label - prediction)) * learning_rate
	// dB = (-2/n) * sum(label - prediction) * learning_rate

	err := l.utils.SubNew(*y, output)

	dM := make([]ckks.Ciphertext, len(input))
	multiplier := l.utils.EncodePlaintextFromArray(l.utils.GenerateFilledArraySize((-2.0/float64(size))*learningRate, size))

	for i := range input {
		dM[i] = l.utils.MultiplyNew(input[i], *err.CopyNew(), true, false)
		l.utils.SumElementsInPlace(&dM[i])
		l.utils.MultiplyPlain(&dM[i], &multiplier, &dM[i], true, false)
	}

	dB := l.utils.SumElementsNew(err)
	l.utils.MultiplyPlain(&dB, &multiplier, &dB, true, false)

	return LinearRegressionGradient{dM, dB}

}

func (l *LinearRegression) UpdateGradient(gradient LinearRegressionGradient) {

	for i := range gradient.DM {
		l.utils.Sub(l.Weight[i], gradient.DM[i], &l.Weight[i])
	}

	l.utils.Sub(l.Bias, gradient.DB, &l.Bias)

}

// pack data in array of ciphertexts
func (model *LinearRegression) Train(x []ckks.Ciphertext, y *ckks.Ciphertext, learningRate float64, size int, epoch int) {

	log := logger.NewLogger(true)

	log.Log("Starting Linear Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Forward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Forward(x)

		log.Log("Backward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		grad := model.Backward(x, fwd, y, size, learningRate)

		log.Log("Updating gradient " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch) + "\n")
		model.UpdateGradient(grad)

		if model.Weight[0].Level() < 4 || model.Bias.Level() < 4 {
			fmt.Println("Bootstrapping gradient")
			if model.Bias.Level() != 1 {
				model.utils.Evaluator.DropLevel(&model.Bias, model.Bias.Level()-1)
			}
			for i := range x {
				model.utils.BootstrapInPlace(&model.Weight[i])
			}
			model.utils.BootstrapInPlace(&model.Bias)
		}

	}

}
