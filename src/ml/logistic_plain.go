package ml

import (
	"fmt"
	"math"
)

type LogisticRegression struct {
	b0 float64 //intercept
	b1 float64 //data-point 1
	b2 float64 //data-point 2

}

func NewLogisticRegression() LogisticRegression {
	return LogisticRegression{1, 1, 1}
}

func Predict(model LogisticRegression, x []float64, y []float64, j int) float64 {

	// yhat = b0 + b1*x + b2*y
	// return sigmoid(yhat)

	yhat := model.b0 + model.b1*(x[j]) + model.b2*(y[j])

	return SigmoidNew(yhat)

}

func Coefficients_Sgd(x []float64, y []float64, target []float64, model LogisticRegression, l float64, epoch int) LogisticRegression {
	fmt.Printf("l: %f Epoch: %o \n", l, epoch)
	for i := 0; i < epoch; i++ {
		for j := 0; j < len(x); j++ {
			yhat := Predict(model, x, y, j)
			fmt.Printf("y hat = %f \n", yhat)
			error := target[j] - yhat
			model.b0 += l * error * yhat * (1 - yhat)
			model.b1 += l * error * yhat * (1 - yhat) * x[j]
			model.b2 += l * error * yhat * (1 - yhat) * y[j]
		}
	}
	fmt.Printf("Trained -> b0: %f b1: %f, b2: %f", model.b0, model.b1, model.b2)
	return model
}

// func Train(x []float64, y []float64, target []float64, model LogisticRegression, l float64, epoch int){

// }
func SigmoidNew(input float64) float64 {
	// In real case, evaluate sigmoid by taylor estimation
	// 0.5 + 0.197x + 0.004x^3

	return 1.0 / (1.0 + math.Exp(-input))

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
