package ml

import (
	// "fmt"
	"fmt"
	"math"
	"strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/logger"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

type LogisticRegression struct {
	utils utility.Utils
	b0    ckks.Ciphertext // intercept
	b1    ckks.Ciphertext // data1
	b2    ckks.Ciphertext // data2
}

type LogisticRegressionGradient struct {
	Db0 ckks.Ciphertext
	Db1 ckks.Ciphertext
	Db2 ckks.Ciphertext
}

func NewLogisticRegression(u utility.Utils) LogisticRegression {

	zeros := u.GenerateFilledArray(0.0)
	b0 := u.Encrypt(zeros)
	b1 := u.Encrypt(zeros)
	b2 := u.Encrypt(zeros)

	return LogisticRegression{u, b0, b1, b2}

}

func (lr LogisticRegression) Sigmoid(x ckks.Ciphertext) ckks.Ciphertext {

	// Evaluate Sigmoid according to the approximation
	// 0.5 + 0.197x + 0.004x^3

	output := lr.utils.MultiplyNew(lr.utils.MultiplyNew(x, lr.utils.MultiplyConstArrayNew(x, lr.utils.GenerateFilledArray(0.004), true, false), true, false), lr.utils.MultiplyNew(x, x, true, false), true, false) // output = x * x
	// output = utils.MultiplyNew(x, utils.MultiplyConstNew(x, utils.GenerateFilledArray(0.004), true, false), true, false) // output = output * (x * 0.004)
	output = lr.utils.SubNew(lr.utils.MultiplyConstArrayNew(x, lr.utils.GenerateFilledArray(0.197), true, false), output) // output = output + 0.197 * x

	SigCont := lr.utils.GenerateFilledArray(0.5)
	encoded := lr.utils.EncodeToScale(SigCont, math.Pow(2.0, 20.0))
	lr.utils.ReEncodeAsNTT(&encoded)

	output = lr.utils.AddPlainNew(output, encoded)

	return output

}

func SigmoidCheck(input float64) float64 {
	// In real case, evaluate sigmoid by taylor estimation
	// 0.5 + 0.197x + 0.004x^3

	return 1.0 / (1.0 + math.Exp(-1*input))

}

func (lr LogisticRegression) PredictCipher(x ckks.Ciphertext, y ckks.Ciphertext) ckks.Ciphertext {
	// Predict whether it is class 0 or 1

	// yhat = b0 + b1*x + b2*y
	// return sigmoid(yhat)
	yhat := lr.utils.AddNew(lr.utils.MultiplyNew(lr.b2, y, true, false), lr.utils.MultiplyNew(lr.b1, x, true, false))
	yhat = lr.utils.AddNew(yhat, lr.b0)

	// decrypt yhat first then evaluate sigmoid?

	return lr.Sigmoid(yhat)
}

func (lr LogisticRegression) Sgd(x ckks.Ciphertext, y ckks.Ciphertext, target ckks.Ciphertext, learningRate float64, size int) LogisticRegressionGradient {
	// Calculate backward gradient using the following equation
	// dM = (-2/n) * sum(input * (label - prediction)) * learning_rate
	yhat := lr.PredictCipher(x, y)       // get yhat
	err := lr.utils.SubNew(target, yhat) // find error of yhat and target

	Db2 := lr.utils.MultiplyNew(y, *err.CopyNew(), true, false)
	lr.utils.SumElementsInPlace(&Db2)
	lr.utils.MultiplyConstArray(&Db2, lr.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &Db2, true, false)

	Db1 := lr.utils.MultiplyNew(x, *err.CopyNew(), true, false)
	lr.utils.SumElementsInPlace(&Db1)
	lr.utils.MultiplyConstArray(&Db1, lr.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &Db1, true, false)

	Db0 := lr.utils.SumElementsNew(err)
	lr.utils.MultiplyConstArray(&Db0, lr.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &Db0, true, false)

	return LogisticRegressionGradient{Db0, Db1, Db2}
}

func (lr *LogisticRegression) UpdateGradient(gradient LogisticRegressionGradient) {
	lr.utils.Sub(lr.b0, gradient.Db0, &lr.b0)
	lr.utils.Sub(lr.b1, gradient.Db1, &lr.b1)
	lr.utils.Sub(lr.b2, gradient.Db2, &lr.b2)
}

func (model *LogisticRegression) TrainLR(x ckks.Ciphertext, y ckks.Ciphertext, target ckks.Ciphertext, learningRate float64, size int, epoch int) {

	log := logger.NewLogger(true)
	log.Log("Starting Logistic Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Performing SGD " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Sgd(x, y, target, learningRate, size)

		log.Log("Updating gradients " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		model.UpdateGradient(fwd)

		// b0 := model.utils.Decrypt(&model.b0)
		// b1 := model.utils.Decrypt(&model.b1)
		// b2 := model.utils.Decrypt(&model.b2)

		// fmt.Printf("The three coefficients are %f, %f, and %f", b0[0], b1, b2)
		fmt.Println("Finished Training")
	}
}

func (model LogisticRegression) AccuracyTest(x []float64, y []float64, target []float64, size int) float64 {

	// Evaluate b0 + b1*data1 + b2*data2 then into sigmoid then check with target

	// model.utils.Decrypt(&x)
	// model.utils.Decrypt(&y)
	// model.utils.Decrypt(&target)

	b0plain := model.utils.Decrypt(&model.b0)
	b1plain := model.utils.Decrypt(&model.b1)
	b2plain := model.utils.Decrypt(&model.b2)

	correct := 0
	incorrect := 0

	for i := 0; i < size; i++ {
		yhat := b0plain[i] + b1plain[i]*x[i] + b2plain[i]*y[i]
		guess := SigmoidCheck(yhat)

		if guess > 0.5 {
			guess = 1
		} else {
			guess = 0
		}

		if guess == target[i] {
			correct++
		} else {
			incorrect++
		}
	}

	acc := float64(correct) / float64(size) * float64(100)

	return acc
}
