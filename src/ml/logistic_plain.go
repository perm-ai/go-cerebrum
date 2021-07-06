package ml

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type LogisticRegression struct {
	b0 float64 //intercept
	b1 float64 //data-point 1
	b2 float64 //data-point 2

}

func NewLogisticRegression() LogisticRegression {
	return LogisticRegression{0.5, 0.5, 0.5}
}

func Predict(model LogisticRegression, x []float64, y []float64, j int) float64 {

	// yhat = b0 + b1*x + b2*y
	// return sigmoid(yhat)
	var yhat float64
	yhat = model.b0 + model.b1*(x[j]) + model.b2*(y[j])

	return SigmoidNew(yhat)

}

func Coefficients_Sgd(model LogisticRegression, x []float64, y []float64, target []float64, l float64, epoch int) LogisticRegression {
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
	// In real case, evaluate sigmoid by taylor estimation
	// 0.5 + 0.197x + 0.004x^3
	y := 0.5 + 0.197*x + 0.004*math.Pow(x, 3.0)
	return y

}

func Train(model LogisticRegression, x []float64, y []float64, target []float64, l float64, epoch int) float64 {
	rand.Seed(time.Now().UnixNano())
	// fmt.Println(x)
	// fmt.Println(y)
	// fmt.Printf("Amount of data : %o \n", len(x))
	// fmt.Printf("Amount of data : %o \n", len(y))
	NumberOfTestData := 20
	fmt.Printf("Amount of test data : %o \n", NumberOfTestData)
	xtest := make([]float64, NumberOfTestData)
	ytest := make([]float64, NumberOfTestData)
	targettest := make([]float64, NumberOfTestData)
	for i := 0; i < NumberOfTestData; i++ {
		OrderRemoved := (int)(math.Floor((rand.Float64() * (float64)(len(x)))))
		// fmt.Printf("Remove : %o \n", OrderRemoved)
		xtest[i] = x[OrderRemoved]
		ytest[i] = y[OrderRemoved]
		targettest[i] = target[OrderRemoved]
		// remove(x, OrderRemoved)
		// remove(y, OrderRemoved)
		// remove(target, OrderRemoved)
		// NumberOfData--
	}
	fmt.Println("Training time")
	model = Coefficients_Sgd(model, x, y, target, l, epoch)
	acc := Test(model, x, y, target) * 100
	fmt.Printf("Accuracy : %f \n", acc)
	return acc

}
func Test(model LogisticRegression, xtest []float64, ytest []float64, targettest []float64) float64 {
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
	MinMax := FindMinMax(input)
	for i := 0; i < len(input); i++ {
		input[i] = (input[i] - MinMax[0]) / (MinMax[1] - MinMax[0])
	}
}

func remove(s []float64, index int) []float64 {
	return append(s[:index], s[index+1:]...)
}
