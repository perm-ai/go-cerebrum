package regression

import (
	// "fmt"

	"fmt"
	"strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/array"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

type LogisticRegression struct {
	utils  utility.Utils
	weight []ckks.Ciphertext
	bias   ckks.Ciphertext
}

type LogisticRegressionGradient struct {
	dw []ckks.Ciphertext
	db ckks.Ciphertext
}

type Data struct {
	x          []ckks.Ciphertext
	target     ckks.Ciphertext
	datalength int
}

type DataPlain struct {
	x      [][]float64
	target []float64
}

func NewLogisticRegression(u utility.Utils, column int) LogisticRegression {

	value := u.GenerateFilledArray(0.5)
	w := make([]ckks.Ciphertext, column)
	for i := 0; i < column; i++ {
		w[i] = u.Encrypt(value)
	}
	b := u.Encrypt(value)

	return LogisticRegression{u, w, b}

}

func (model LogisticRegression) Forward(data Data) ckks.Ciphertext {

	//prediction(yhat) = sigmoid(w1*x1+w2*x2+...+b)
	result := model.utils.Encrypt(model.utils.GenerateFilledArray(0.0))
	sigmoid := activations.Sigmoid{U: model.utils}
	//w[i]*x[i]
	for i := range data.x {
		weight := model.utils.MultiplyNew(model.weight[i], data.x[i], true, false)
		model.utils.Add(weight, result, &result)
		fmt.Println("Computing w" + fmt.Sprint(i+1))

	}
	model.utils.Add(model.bias, result, &result)
	model.utils.MultiplyConst(&result, 0.1, &result, true, false)
	fmt.Println("Forward complete, computing sigmoid")
	if result.Level() < 5 {
		fmt.Println("bootstrapping result")
		model.utils.BootstrapInPlace(&result)
	}
	fmt.Println("result level : " + fmt.Sprint(result.Level()))
	return sigmoid.Forward(result, data.datalength)

}

func (model LogisticRegression) Backward(data Data, predict ckks.Ciphertext, lr float64) LogisticRegressionGradient {

	//error = prediction - actual data
	//gradientw = (2/n)(sum(error*datax))
	//gradientb = (2/n)(sum(error))
	dw := make([]ckks.Ciphertext, len(model.weight))
	err := model.utils.SubNew(predict, data.target)
	for i := range model.weight {
		fmt.Println("Computing w" + fmt.Sprint(i+1))
		dw[i] = model.utils.MultiplyNew(data.x[i], *err.CopyNew(), true, false)
		model.utils.SumElementsInPlace(&dw[i])
		model.utils.MultiplyConstArray(&dw[i], model.utils.GenerateFilledArraySize((-2/float64(data.datalength))*lr, data.datalength), &dw[i], true, false)
	}

	db := model.utils.SumElementsNew(err)
	model.utils.MultiplyConstArray(&db, model.utils.GenerateFilledArraySize((-2/float64(data.datalength))*lr, data.datalength), &db, true, false)

	return LogisticRegressionGradient{dw, db}

}

func (model *LogisticRegression) UpdateGradient(grad LogisticRegressionGradient) {

	for i := range grad.dw {
		model.utils.Sub(model.weight[i], grad.dw[i], &model.weight[i])
	}
	model.utils.Sub(model.bias, grad.db, &model.bias)

}

func (model *LogisticRegression) Train(data Data, learningRate float64, epoch int, test bool) {
	log := logger.NewLogger(test)
	log.Log("Starting Logistic Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Forward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Forward(data)
		log.Log("result level : " + fmt.Sprint(fwd.Level()))
		log.Log("result :" + fmt.Sprint(model.utils.Decrypt(fwd.CopyNew())[0:10]))
		log.Log("Backward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		grad := model.Backward(data, fwd, learningRate)
		log.Log("Gradient db level : " + fmt.Sprint(grad.db.Level()))
		log.Log("Gradient dw level : " + fmt.Sprint(grad.dw[0].Level()))

		log.Log("Updating gradient " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch) + "\n")
		model.UpdateGradient(grad)
		log.Log("weight 1 : " + fmt.Sprint(model.utils.Decrypt(model.weight[0].CopyNew())[0]))
		log.Log("weight 2 : " + fmt.Sprint(model.utils.Decrypt(model.weight[1].CopyNew())[0]))
		log.Log("bias : " + fmt.Sprint(model.utils.Decrypt(model.bias.CopyNew())[0]))
		log.Log("weight 1 level : " + fmt.Sprint(model.weight[0].Level()))
		log.Log("weight 2 level : " + fmt.Sprint(model.weight[1].Level()))
		log.Log("bias level : " + fmt.Sprint(model.bias.Level()))
		if model.weight[0].Level() < 7 {
			for i := range model.weight {
				model.utils.BootstrapInPlace(&model.weight[i])
			}
			model.utils.BootstrapInPlace(&model.bias)
		}
		log.Log("Bootstrap complete")
		log.Log("weight 1 : " + fmt.Sprint(model.utils.Decrypt(model.weight[0].CopyNew())[0]))
		log.Log("weight 2 : " + fmt.Sprint(model.utils.Decrypt(model.weight[1].CopyNew())[0]))
		log.Log("bias : " + fmt.Sprint(model.utils.Decrypt(model.bias.CopyNew())[0]))
		log.Log("weight 1 level : " + fmt.Sprint(model.weight[0].Level()))
		log.Log("weight 2 level : " + fmt.Sprint(model.weight[1].Level()))
		log.Log("bias level : " + fmt.Sprint(model.bias.Level()))
	}
}
func (model LogisticRegression) LogTest(data DataPlain) {
	//test the model and output accuracy
	fmt.Printf("Testing accuracy")
	wplain := make([]float64, len(model.weight))
	bplain := model.utils.Decrypt(&model.bias)[0]
	for i := range wplain {
		wplain[i] = model.utils.Decrypt(&model.weight[i])[0]
	}
	fmt.Println("w : " + fmt.Sprint(wplain))
	fmt.Println("b : " + fmt.Sprint(bplain))
	//get prediction
	correct := 0
	predictTarget := make([]float64, len(data.x))
	for j, p := range data.x {
		predictTarget = array.AddArraysNew(array.MulConstantArrayNew(wplain[j], p), predictTarget)
	}
	array.AddConstant(bplain, predictTarget, predictTarget)
	guess := array.SigmoidArray(predictTarget)
	//Check if correct
	var trueguess int
	for i, p := range guess {
		if p > 0.5 {
			trueguess = 1
		} else {
			trueguess = 0
		}
		fmt.Printf("(%f)Predicted : %d, Expected : %f", guess, trueguess, data.target[i])
		if p == data.target[i] {
			correct++
		}
	}

	acc := float64(correct) / float64(len(guess)) * 100.0
	fmt.Printf("Accuracy : %f", acc)

}
