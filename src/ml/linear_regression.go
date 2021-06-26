package ml

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

type LinearRegression struct {

	utils 	utility.Utils
	M 		ckks.Ciphertext
	B		ckks.Ciphertext

}

type LinearRegressionGradient struct {

	DM 		ckks.Ciphertext
	DB		ckks.Ciphertext

}

func NewLinearRegression(u utility.Utils) LinearRegression {

	zeros := u.GenerateFilledArray(0.0)
	m := u.Encrypt(zeros)
	b := u.Encrypt(zeros)

	return LinearRegression{u, m, b}

}

func (l LinearRegression) Forward(input *ckks.Ciphertext) ckks.Ciphertext {

	result := l.utils.MultiplyNew(input, &l.M)
	l.utils.Add(result, l.B, &result)

	return result

}

func (l LinearRegression) Backward(input *ckks.Ciphertext, output *ckks.Ciphertext, y *ckks.Ciphertext, size int) LinearRegressionGradient{

	err := l.utils.Evaluator.SubNew(y, output)

	averager := l.utils.Encrypt(l.utils.GenerateFilledArray(-2 / float64(size)))

	dM := l.utils.MultiplyNew(input, err)
	l.utils.SumElementsInPlace(&dM)
	l.utils.Multiply(dM, averager, &dM)

	dB := l.utils.SumElementsNew(*err)
	l.utils.Multiply(averager, dB, &dB)

	return LinearRegressionGradient{dM, dB}

}

func (l *LinearRegression) UpdateGradient(gradient LinearRegressionGradient, learningRate float64){

	lrCt := l.utils.Encrypt(l.utils.GenerateFilledArray(learningRate))
	l.utils.Multiply(gradient.DM, lrCt, &gradient.DM)
	l.utils.Multiply(gradient.DB, lrCt, &gradient.DB)

	l.utils.Sub(l.M, gradient.DM, &l.M)
	l.utils.Sub(l.B, gradient.DB, &l.B)

}