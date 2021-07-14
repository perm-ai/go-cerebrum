package ml

// import (
// 	"fmt"
// 	"math"
// 	"strconv"

// 	// "github.com/ldsec/lattigo/v2/ckks"
// 	"github.com/perm-ai/GO-HEML-prototype/src/logger"
// )

// type LogisticRegression2 struct {
// 	b0 float64 //intercept
// 	b1 float64 //data-point 1
// 	b2 float64 //data-point 2

// }
// type LogisticRegressionGradientPlain struct {
// 	Db0 float64
// 	Db1 float64
// 	Db2 float64
// }

// func NewLogisticRegression2() LogisticRegression2 {
// 	return LogisticRegression2{0.5, 0.5, 0.5}
// }

// func (lr LogisticRegression2) PredictCipherPlain(x []float64, y []float64) []float64 {
// 	// Predict whether it is class 0 or 1

// 	// yhat = b0 + b1*x + b2*y
// 	// return sigmoid(yhat)
// 	var yhat = make([]float64, len(x))
// 	for i := 0; i < len(x); i++ {
// 		yhat[i] = lr.b0 + lr.b1*x[i] + lr.b2*y[i]
// 	}

// 	// decrypt yhat first then evaluate sigmoid?

// 	return SigmoidArray(yhat)
// }

// func (lr LogisticRegression2) SgdPlain(x []float64, y []float64, target []float64, learningRate float64, size int) LogisticRegressionGradientPlain {
// 	// Calculate backward gradient using the following equation
// 	// dM = (-2/n) * sum(input * (label - prediction)) * learning_rate
// 	yhat := lr.PredictCipherPlain(x, y) // get yhat
// 	var err = make([]float64, len(x))
// 	var Db2 = make([]float64, len(x))
// 	var Db1 = make([]float64, len(x))
// 	var Db0 = make([]float64, len(x))
// 	err = subtArray(target, yhat)
// 	for i := 0; i < len(x); i++ {
// 		// err := lr.utils.SubNew(target, yhat) // find error of yhat and target
// 		Db2[i] = (-2 / float64(size)) * learningRate * sumArray(mulArray(y, err))
// 		Db1[i] = (-2 / float64(size)) * learningRate * sumArray(mulArray(x, err))
// 		Db0[i] = (-2 / float64(size)) * learningRate * sumArray(err)
// 	}
// 	// Db2 := lr.utils.MultiplyNew(y, *err.CopyNew().Ciphertext(), true, false)
// 	// lr.utils.SumElementsInPlace(&Db2)
// 	// lr.utils.MultiplyConst(&Db2, lr.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &Db2, true, false)

// 	// Db1 := lr.utils.MultiplyNew(x, *err.CopyNew().Ciphertext(), true, false)
// 	// lr.utils.SumElementsInPlace(&Db1)
// 	// lr.utils.MultiplyConst(&Db1, lr.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &Db1, true, false)

// 	// Db0 := lr.utils.SumElementsNew(err)
// 	// lr.utils.MultiplyConst(&Db0, lr.utils.GenerateFilledArraySize((-2/float64(size))*learningRate, size), &Db0, true, false)
// 	// }
// 	return LogisticRegressionGradientPlain{}
// }

// func (lr *LogisticRegression2) UpdateGradientPlain(gradient LogisticRegressionGradientPlain) {
// 	// lr.utils.Sub(lr.b0, gradient.Db0, &lr.b0)
// 	// lr.utils.Sub(lr.b1, gradient.Db1, &lr.b1)
// 	// lr.utils.Sub(lr.b2, gradient.Db2, &lr.b2)
// 	lr.b0 = lr.b0 - gradient.Db0
// 	lr.b1 = lr.b1 - gradient.Db1
// 	lr.b2 = lr.b2 - gradient.Db2
// }

// func (model *LogisticRegression2) TrainLRPlain(x []float64, y []float64, target []float64, learningRate float64, size int, epoch int) {

// 	log := logger.NewLogger(true)
// 	log.Log("Starting Linear Regression Training on encrypted data")

// 	for i := 0; i < epoch; i++ {

// 		log.Log("Performing SGD " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
// 		fwd := model.SgdPlain(x, y, target, learningRate, size)

// 		log.Log("Updating gradients " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
// 		model.UpdateGradientPlain(fwd)

// 		// b0 := model.utils.Decrypt(&model.b0)
// 		// b1 := model.utils.Decrypt(&model.b1)
// 		// b2 := model.utils.Decrypt(&model.b2)

// 		// fmt.Printf("The three coefficients are %f, %f, and %f", b0[0], b1, b2)
// 		fmt.Println("Finished Training")
// 	}
// }

// func (model LogisticRegression2) AccuracyTestPlain(x []float64, y []float64, target []float64, size int) float64 {

// 	// Evaluate b0 + b1*data1 + b2*data2 then into sigmoid then check with target

// 	// model.utils.Decrypt(&x)
// 	// model.utils.Decrypt(&y)
// 	// model.utils.Decrypt(&target)

// 	correct := 0
// 	incorrect := 0

// 	for i := 0; i < size; i++ {
// 		yhat := model.b0 + model.b1*x[i] + model.b2*y[i]
// 		guess := SigmoidNew(yhat)

// 		if guess > 0.5 {
// 			guess = 1
// 		} else {
// 			guess = 0
// 		}

// 		if guess == target[i] {
// 			correct++
// 		} else {
// 			incorrect++
// 		}
// 	}

// 	acc := float64(correct) / float64(size) * float64(100)

// 	return acc
// }

// func SigmoidArray(x []float64) []float64 {
// 	var sig = make([]float64, len(x))
// 	for i := 0; i < len(x); i++ {
// 		sig[i] = 1.0 / (1.0 + math.Exp(-1*x[i]))
// 	}
// 	return sig
// }

// func sumArray(x []float64) float64 {
// 	sum := 0.0
// 	for i := 0; i < len(x); i++ {
// 		sum += x[i]
// 	}
// 	return sum
// }
// func subtArray(x []float64, y []float64) []float64 {
// 	var Ans = make([]float64, len(x))
// 	for i := 0; i < len(x); i++ {
// 		Ans[i] = x[i] - y[i]
// 	}
// 	return Ans
// }
// func mulArray(x []float64, y []float64) []float64 {
// 	var Ans = make([]float64, len(x))
// 	for i := 0; i < len(x); i++ {
// 		Ans[i] = x[i] * y[i]
// 	}
// 	return Ans
// }
