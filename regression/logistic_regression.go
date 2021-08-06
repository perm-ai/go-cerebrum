package regression

import (
	// "fmt"

	"fmt"
	"strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/activations"
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

// func NewDataPlain(x1 []float64, x2 []float64, target []float64) DataPlain {
// 	return DataPlain{x1, x2, target}
// }

// func EncryptData(data DataPlain, utils utility.Utils) Data {
// 	log := logger.NewLogger(true)
// 	log.Log("Encrypting X1")
// 	encX1 := utils.Encrypt(data.x1)
// 	log.Log("Encrypting X2")
// 	encX2 := utils.Encrypt(data.x2)
// 	log.Log("Encrypting target")
// 	enctar := utils.Encrypt(data.target)
// 	log.Log("Encryption complete")
// 	return Data{encX1, encX2, enctar, len(data.x1)}

// }
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
	}
	model.utils.Add(model.bias, result, &result)
	model.utils.MultiplyConst(&result, 0.1, &result, true, false)
	return sigmoid.Forward(result, data.datalength)

}

func (model LogisticRegression) Backward(data Data, predict ckks.Ciphertext, lr float64) LogisticRegressionGradient {

	//error = prediction - actual data
	//gradientw = (2/n)(sum(error*datax))
	//gradientb = (2/n)(sum(error))
	dw := make([]ckks.Ciphertext, len(model.weight))
	err := model.utils.SubNew(predict, data.target)
	for i := range model.weight {
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

func (model *LogisticRegression) Train(data Data, learningRate float64, epoch int) {
	log := logger.NewLogger(true)
	log.Log("Starting Logistic Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Forward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Forward(data)

		log.Log("Backward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		grad := model.Backward(data, fwd, learningRate)

		log.Log("Updating gradient " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch) + "\n")
		model.UpdateGradient(grad)

		if model.weight[0].Level() < 7 {
			fmt.Println("Bootstrapping gradient")
			if model.bias.Level() != 1 {
				model.utils.Evaluator.DropLevel(&model.bias, model.bias.Level()-1)
			}
			for i := range model.weight {
				model.utils.BootstrapInPlace(&model.weight[i])
			}
			model.utils.BootstrapInPlace(&model.bias)

		}

	}
}
func (model LogisticRegression) LogTest(data DataPlain) {
	//test the model and output accuracy
	fmt.Printf("Testing accuracy")
	wplain := make([][]float64, len(model.weight))
	bplain := model.utils.Decrypt(&model.bias)
	for i := range wplain {
		wplain[i] = model.utils.Decrypt(&model.weight[i])
	}
	correct := 0

	for i := 0; i < len(data.x); i++ {
		yhat := bplain[i] + w1plain[i]*data.x1[i] + w2plain[i]*data.x2[i]
		guess := SigmoidCheck(yhat)
		var trueguess int
		if guess > 0.5 {
			trueguess = 1
		}
		fmt.Printf("(%f)Predicted : %d, Expected : %f", guess, trueguess, data.target[i])
		if guess == data.target[i] {
			correct++
		}
	}

	acc := float64(correct) / float64(len(data.x1)) * 100.0
	fmt.Printf("Accuracy : %f", acc)

}
