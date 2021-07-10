package ml

// "fmt"

// type LogisticRegressionPlainBatch struct {
// 	b0 []float64 //intercept
// 	b1 []float64 //data-point 1
// 	b2 []float64 //data-point 2

// }

// type LogisticRegressionPlainGradientBatch struct {
// 	Db0 []float64
// 	Db1 []float64
// 	Db2 []float64
// }

// func NewLogisticRegressionPlainBatch() LogisticRegressionPlain {
// 	return LogisticRegressionPlain{0.0, 0.0, 0.0}
// }

// func PredictBatch(model LogisticRegressionPlain, x []float64, y []float64, j int) float64 {

// 	// Predict whether it is class 0 or 1

// 	// yhat = b0 + b1*x + b2*y
// 	// return sigmoid(yhat)
// 	yhat := model.b0 + model.b1*x + model.b2*y

// 	return SigmoidNew(yhat)

// }
