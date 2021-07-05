package ml

import (
	"math"
)

type LogisticRegression struct {

	b0 float64 //intercept
	b1 float64 //data-point 1
	b2 float64 //data-point 2

}

func NewLogisticRegression() LogisticRegression{
	return LogisticRegression{0, 0, 0}
}

func Predict(model LogisticRegression, x []float64, y []float64) []float64 {

	// yhat = b0 + b1*x + b2*y
	// return sigmoid(yhat)
	ans := make([]float64, len(x), 32768)

	for row := 0; row < len(x); row++ {


		yhat := model.b0 + model.b1*(x[row]) + model.b2*(y[row])

		if SigmoidNew(yhat) > 0.5 {
    		ans[row] = 1
		} else {
    		ans[row] = 0
		}

	}

	return ans

}

func SigmoidNew(input float64) float64 {
	// In real case, evaluate sigmoid by taylor estimation
	// 0.5 + 0.197x + 0.004x^3

	return 1.0/(1.0 + math.Exp(-input))

}
