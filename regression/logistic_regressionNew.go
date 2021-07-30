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

type Logisticmodel struct {
	utils utility.Utils
	w1    ckks.Ciphertext
	w2    ckks.Ciphertext
	b     ckks.Ciphertext
}

type LogisticGradient struct {
	dw1 ckks.Ciphertext
	dw2 ckks.Ciphertext
	db  ckks.Ciphertext
}

type Data struct {
	x1         ckks.Ciphertext
	x2         ckks.Ciphertext
	target     ckks.Ciphertext
	datalength int
}

type DataPlain struct {
	x1     []float64
	x2     []float64
	target []float64
}

func NewDataPlain(x1 []float64, x2 []float64, target []float64) DataPlain {
	return DataPlain{x1, x2, target}
}

func EncryptData(data DataPlain, utils utility.Utils) Data {
	log := logger.NewLogger(true)
	log.Log("Encrypting X1")
	encX1 := utils.Encrypt(data.x1)
	log.Log("Encrypting X2")
	encX2 := utils.Encrypt(data.x2)
	log.Log("Encrypting target")
	enctar := utils.Encrypt(data.target)
	log.Log("Encryption complete")
	return Data{encX1, encX2, enctar, len(data.x1)}

}
func NewLogisticmodel(u utility.Utils) Logisticmodel {

	value := u.GenerateFilledArray(0.5)
	w1 := u.Encrypt(value)
	w2 := u.Encrypt(value)
	b := u.Encrypt(value)

	return Logisticmodel{u, w1, w2, b}

}

func (model Logisticmodel) Forward(data Data) ckks.Ciphertext {

	//prediction(yhat) = sigmoid(w1*x1+w2*x2+b)

	sigmoid := activations.Sigmoid{U: model.utils}
	weight1 := model.utils.MultiplyNew(model.w1, data.x1, true, false)
	weight2 := model.utils.MultiplyNew(model.w2, data.x2, true, false)
	result := model.utils.AddNew(model.utils.AddNew(weight1, weight2), model.b)

	return sigmoid.Forward(result, data.datalength)

}

func (model Logisticmodel) Backward(data Data, predict ckks.Ciphertext, lr float64) LogisticGradient {

	//error = prediction - actual data
	//gradient = (2/n)(sum(error*datax))

	err := model.utils.SubNew(predict, data.target)
	dw1 := model.utils.MultiplyNew(data.x1, *err.CopyNew(), true, false)
	model.utils.SumElementsInPlace(&dw1)
	model.utils.MultiplyConstArray(&dw1, model.utils.GenerateFilledArraySize((-2/float64(data.datalength))*lr, data.datalength), &dw1, true, false)

	dw2 := model.utils.MultiplyNew(data.x2, *err.CopyNew(), true, false)
	model.utils.SumElementsInPlace(&dw2)
	model.utils.MultiplyConstArray(&dw2, model.utils.GenerateFilledArraySize((-2/float64(data.datalength))*lr, data.datalength), &dw2, true, false)

	db := model.utils.SumElementsNew(err)
	model.utils.MultiplyConstArray(&db, model.utils.GenerateFilledArraySize((-2/float64(data.datalength))*lr, data.datalength), &db, true, false)

	return LogisticGradient{dw1, dw2, db}

}

func (model *Logisticmodel) UpdateGradient(grad LogisticGradient) {

	model.utils.Sub(model.w1, grad.dw1, &model.w1)
	model.utils.Sub(model.w2, grad.dw2, &model.w2)
	model.utils.Sub(model.b, grad.db, &model.b)

}

func (model *Logisticmodel) Train(data Data, learningRate float64, epoch int) {
	log := logger.NewLogger(true)
	log.Log("Starting Logistic Regression Training on encrypted data")

	for i := 0; i < epoch; i++ {

		log.Log("Forward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		fwd := model.Forward(data)

		log.Log("Backward propagating " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch))
		grad := model.Backward(data, fwd, learningRate)

		log.Log("Updating gradient " + strconv.Itoa(i+1) + "/" + strconv.Itoa(epoch) + "\n")
		model.UpdateGradient(grad)

		if model.w1.Level() < 7 || model.w2.Level() < 7 || model.b.Level() < 7 {
			fmt.Println("Bootstrapping gradient")
			if model.b.Level() != 1 {
				model.utils.Evaluator.DropLevel(&model.b, model.b.Level()-1)
			}
			model.utils.BootstrapInPlace(&model.b)
			model.utils.BootstrapInPlace(&model.w1)
			model.utils.BootstrapInPlace(&model.w2)

			fmt.Printf("NEW b scale and level is %f and %d \n", model.b.Scale, model.b.Level())
			fmt.Printf("NEW w1 scale and level is %f and %d \n", model.w1.Scale, model.w1.Level())
			fmt.Printf("NEW w2 scale and level is %f and %d \n", model.w2.Scale, model.w2.Level())

		}

	}
}
func (model Logisticmodel) LogTest(data DataPlain) {
	//test the model and output accuracy
	fmt.Printf("Testing accuracy")
	bplain := model.utils.Decrypt(&model.b)
	w1plain := model.utils.Decrypt(&model.w1)
	w2plain := model.utils.Decrypt(&model.w2)
	correct := 0

	for i := 0; i < len(data.x1); i++ {
		yhat := bplain[i] + w1plain[i]*data.x1[i] + w2plain[i]*data.x2[i]
		guess := SigmoidCheck(yhat)
		var trueguess int
		if guess > 0.5 {
			trueguess = 1
		}
		fmt.Println("(%f)Predicted : %d, Expected : %f", guess, trueguess, data.target[i])
		if guess == data.target[i] {
			correct++
		}
	}

	acc := float64(correct) / float64(len(data.x1)) * 100.0
	fmt.Println("Accuracy : %f", acc)

}
