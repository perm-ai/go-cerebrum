package ml

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type LogisticRegressionPlain struct {
	b0 float64 //intercept
	b1 float64 //data-point 1
	b2 float64 //data-point 2

}

func NewLogisticRegressionPlain() LogisticRegressionPlain {
	return LogisticRegressionPlain{0.5, 0.5, 0.5}
}

func Predict(model LogisticRegressionPlain, x []float64, y []float64, j int) float64 {

	// Predict whether it is class 0 or 1

	// yhat = b0 + b1*x + b2*y
	// return sigmoid(yhat)
	yhat := model.b0 + model.b1*(x[j]) + model.b2*(y[j])

	return SigmoidApprox(yhat)

}

func Coefficients_Sgd(model LogisticRegressionPlain, x []float64, y []float64, target []float64, l float64, epoch int) LogisticRegressionPlain {

	// Perform stochastic gradient descent according to equation:
	// b = b + learning_rate * (y - yhat) * yhat * (1 - yhat) * x

	// Where b is the coefficient or weight being optimized, learning_rate is a learning
	// rate that you must configure (e.g. 0.01), (y – yhat) is the prediction error for the
	// model on the training data attributed to the weight, yhat is the prediction made by the
	// coefficients and x is the input value

	//fmt.Printf("l: %f Epoch: %o \n", l, epoch)
	for i := 0; i < epoch; i++ {
		for j := 0; j < len(x); j++ {
			yhat := Predict(model, x, y, j)
			// fmt.Printf("yhat: %f \n", yhat)
			error := target[j] - yhat
			// fmt.Printf("error: %f \n", error)
			model.b0 = model.b0 + (l * error * (1 - yhat) * yhat)
			model.b1 = model.b1 + (l * error * yhat * (1 - yhat) * x[j])
			model.b2 = model.b2 + (l * error * (1 - yhat) * yhat * y[j])
		}
	}
	//fmt.Printf("Trained -> b0: %f b1: %f, b2: %f \n", model.b0, model.b1, model.b2)
	return model
}

func SigmoidNew(input float64) float64 {
	// In real case, evaluate sigmoid by taylor estimation
	// 0.5 + 0.197x + 0.004x^3

	return 1.0 / (1.0 + math.Exp(-1*input))

}
func SigmoidApprox(x float64) float64 {

	y := 0.5 + 0.197*x + 0.004*math.Pow(x, 3.0)
	return y

}

func Train(model LogisticRegressionPlain, x []float64, y []float64, target []float64, l float64, epoch int) float64 {

	// Partition a test set and a training set
	// choose number of test data

	rand.Seed(time.Now().UnixNano())
	// NumberOfTestData := 20
	// fmt.Printf("Amount of test data : %o \n", NumberOfTestData)
	xtest := x
	ytest := y
	targettest := target
	// for i := 0; i < NumberOfTestData; i++ {
	// 	OrderRemoved := (int)(math.Floor((rand.Float64() * (float64)(len(x)))))
	// 	// fmt.Printf("Remove : %o \n", OrderRemoved)
	// 	xtest[i] = x[OrderRemoved]
	// 	ytest[i] = y[OrderRemoved]
	// 	targettest[i] = target[OrderRemoved]
	// 	remove(x, OrderRemoved)
	// 	remove(y, OrderRemoved)
	// 	remove(target, OrderRemoved)
	// 	NumberOfTestData--
	// }

	fmt.Println("Starting training process")

	model = Coefficients_Sgd(model, x, y, target, l, epoch) // train datasets

	acc := Test(model, xtest, ytest, targettest) * 100 //testing accuracy from testing data set
	fmt.Printf("Accuracy : %f \n", acc)
	return acc

}
func Test(model LogisticRegressionPlain, xtest []float64, ytest []float64, targettest []float64) float64 {

	// This function outputs accuracy of the model

	CorrectPrediction := 0
	for i := 0; i < len(xtest); i++ {
		PredictedTarget := Predict(model, xtest, ytest, i)
		// fmt.Printf("Data : %f Predicted = %f \n", targettest[i], math.Round(PredictedTarget))
		if math.Round(PredictedTarget) == targettest[i] {
			CorrectPrediction++
		}
	}
	return (float64)(CorrectPrediction) / (float64)(len(xtest))

}
func FindMinMax(input []float64) [2]float64 {

	// find minmax in format {min, max}

	max := math.Pow(2, -15)
	min := math.Pow(2, 15)
	for _, value := range input {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}
	minmax := [2]float64{min, max}
	return minmax
}

func Normalize_Data(input []float64) {

	// normalize dataset into (0,1)

	MinMax := FindMinMax(input)
	for i := 0; i < len(input); i++ {
		input[i] = (input[i] - MinMax[0]) / (MinMax[1] - MinMax[0])
	}
}

// func remove(s []float64, index int) []float64 {

// 	// remove and shift left

// 	return append(s[:index], s[index+1:]...)
// }
