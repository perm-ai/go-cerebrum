package ml

import (
	"fmt"
	"strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/logger"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

type LinearRegression struct {
	utils utility.Utils
	M     ckks.Ciphertext
	B     ckks.Ciphertext
}

type LinearRegressionGradient struct {
	DM ckks.Ciphertext
	DB ckks.Ciphertext
}

func NewLinearRegression(u utility.Utils) LinearRegression {

	zeros := u.GenerateFilledArray(0.0)
	m := u.Encrypt(zeros)
	b := u.Encrypt(zeros)

	return LinearRegression{u, m, b}

}

func (l LinearRegression) Forward(input *ckks.Ciphertext) ckks.Ciphertext {

	result := l.utils.MultiplyRescaleNew(input, &l.M)
	l.utils.Add(result, l.B, &result)

	return result

}

func (l LinearRegression) Backward(input *ckks.Ciphertext, output *ckks.Ciphertext, y *ckks.Ciphertext, size int, learningRate float64) LinearRegressionGradient {

	err := l.utils.Evaluator.SubNew(y, output)

	dM := l.utils.MultiplyRescaleNew(input, err)
	l.utils.SumElementsInPlace(&dM)
	l.utils.MultiplyConstRescale(&dM, l.utils.GenerateFilledArray((-2/float64(size)) * learningRate), &dM)

	dB := l.utils.SumElementsNew(*err)
	l.utils.MultiplyConstRescale(&dB, l.utils.GenerateFilledArray((-2/float64(size)) * learningRate), &dB)

	return LinearRegressionGradient{dM, dB}

}

func (l *LinearRegression) UpdateGradient(gradient LinearRegressionGradient) {

	l.utils.Sub(l.M, gradient.DM, &l.M)
	l.utils.Sub(l.B, gradient.DB, &l.B)
	l.utils.BootstrapIfNecessary(&l.M)
	l.utils.BootstrapIfNecessary(&l.B)

}

func (model *LinearRegression) Train(x *ckks.Ciphertext, y *ckks.Ciphertext, learningRate float64, size int, epoch int) {

	log := logger.NewLogger(true)

	log.Log("Starting Linear Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Forward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Forward(x)
		log.Log("Backward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		grad := model.Backward(x, &fwd, y, size, learningRate)
		log.Log("Updating gradient " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		model.UpdateGradient(grad)
		m := model.utils.Decrypt(&model.M)
		b := model.utils.Decrypt(&model.B)
		fmt.Printf("Result M: %f B: %f\n", m[0], b[0])

	}

}
