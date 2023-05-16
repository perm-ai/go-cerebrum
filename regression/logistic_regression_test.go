package regression

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/array"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

type logmodel struct {
	w []float64
	b float64
}
type DataPlain struct {
	X      [][]float64
	Target []float64
}

func NewDataPlain(x [][]float64, target []float64) DataPlain {
	fmt.Println("Data has coloumn :" + fmt.Sprint(len(x)))
	return DataPlain{x, target}
}
func NewEmptyData(column int, amount int) DataPlain {
	return DataPlain{array.GenFilledArraysofArrays(0, column, amount), make([]float64, amount)}

}

func EncryptData(data DataPlain, utils utility.Utils) Data {
	log := logger.NewLogger(true)
	EnData := make([]*rlwe.Ciphertext, len(data.X))
	for i, p := range data.X {
		log.Log("Encrypting Column" + fmt.Sprint(i))
		EnData[i] = utils.EncryptToPointer(p)
	}
	log.Log("Encrypting target")
	enctar := utils.Encrypt(data.Target)
	log.Log("Encryption complete")
	return Data{EnData, &enctar, len(data.X[0])}

}
func generateData(dat int, testdata int, columnAmount int) (DataPlain, DataPlain, []int) {
	number := make([]int, 2)
	d := NewEmptyData(columnAmount, dat)
	dT := NewEmptyData(columnAmount, testdata)
	rand.Seed(time.Now().UnixNano())
	w := make([]float64, columnAmount)
	for i := 0; i < columnAmount; i++ {
		w[i] = rand.NormFloat64()
		if w[i] < 0.1 && w[i] > -0.1 {
			w[i] = w[i] * 3
		}
	}
	c := rand.NormFloat64()
	// fmt.Println(w)
	// fmt.Println(c)
	ones := 0
	zeros := 0
	for i := 0; i < dat; i++ {
		result := 0.0
		for j := 0; j < columnAmount; j++ {
			d.X[j][i] = rand.NormFloat64()
			result += w[j] * d.X[j][i]
		}
		result += c
		if result > 0 {
			d.Target[i] = 1
			ones++

		} else {
			d.Target[i] = 0
			zeros++
		}
	}
	for i := 0; i < testdata; i++ {
		result := 0.0
		for j := 0; j < columnAmount; j++ {
			dT.X[j][i] = rand.NormFloat64()
			result += w[j] * dT.X[j][i]
		}
		result += c
		if result > 0 {
			dT.Target[i] = 1
			ones++

		} else {
			dT.Target[i] = 0
			zeros++
		}
	}

	number[0] = ones
	number[1] = zeros
	return d, dT, number

}
func TestLogisticRegression(t *testing.T) {

	utils := utility.Utils{}
	log := logger.NewLogger(true)
	keysChain := key.GenerateKeys(0, true, true)
	utils = utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)

	// csvpath := getStringfromConsole("Input path for data train")
	// testpath := getStringfromConsole("Input path for data test")
	// lr, err := strconv.ParseFloat(getStringfromConsole("Input learning rate "), 64)
	// check(err)
	// epoch, err := strconv.Atoi(getStringfromConsole("Input epoch "))
	// check(err)
	// a := []int{0, 1, 2}
	// data := importer.GetCSVNData("./", a, false)
	lr := 0.2
	epoch := 10
	log.Log("lr = .1,epoch = 15")
	data, dataTest, number := generateData(5000, 100, 2)
	for math.Abs((float64)(number[0])-((float64)(number[0]+number[1])/2)) > 400 {
		data, dataTest, number = generateData(5000, 100, 2)
	}
	log.Log("There are " + fmt.Sprint(len(data.X)) + "column")
	log.Log("There are ones : " + fmt.Sprint(number[0]) + " zeros : " + fmt.Sprint(number[1]))
	plaind := NewDataPlain(data.X, data.Target)
	Endata := EncryptData(plaind, utils)
	log.Log("Initializing model")
	model := NewLogisticRegression(utils, 2)
	log.Log("Begin training")
	model.Train(Endata.x, Endata.target, Endata.datalength, lr, epoch)
	log.Log("Training complete testing the model")
	plainT := NewDataPlain(dataTest.X, dataTest.Target)
	modelplain := model.decryptmodel()
	Acc := modelplain.Test(plainT, 0.5)
	fmt.Printf("Accuracy : %.3f", Acc*100)
}
func (model LogisticRegression) decryptmodel() logmodel {
	wplain := make([]float64, len(model.Weight))
	bplain := model.utils.Decrypt(model.Bias)[0]
	for i := range wplain {
		wplain[i] = model.utils.Decrypt(model.Weight[i])[0]
	}
	fmt.Println("w : " + fmt.Sprint(wplain))
	fmt.Println("b : " + fmt.Sprint(bplain))
	return logmodel{wplain, bplain}
}
func (model logmodel) Test(datatest DataPlain, threshold float64) float64 {

	predictedTarget := model.predict(datatest)
	prediction := make([]int, len(datatest.Target))
	correct := 0.0
	for i, p := range predictedTarget {
		fmt.Print("Data : [ ")
		for _, p := range datatest.X {
			fmt.Printf("%.3f ", p[i])
		}
		fmt.Println("]")
		if p >= threshold {
			prediction[i] = 1
		} else {
			prediction[i] = 0
		}
		fmt.Printf("Calculated : %.3f [Predicted %d , Actual data %f]\n\n", predictedTarget[i], prediction[i], datatest.Target[i])
		if float64(prediction[i]) == datatest.Target[i] {
			correct++
		}
	}
	return correct / float64(len(datatest.Target))
}

func (model logmodel) predict(data DataPlain) []float64 {

	dotProd := make([]float64, len(data.Target))
	for i, p := range data.X {
		dotProd = AddArrays(MulConst(p, model.w[i]), dotProd)
	}
	predictTarget := AddConst(dotProd, model.b)
	return SigmoidPlain(predictTarget)

}

func SigmoidPlain(input []float64) []float64 {
	output := make([]float64, len(input))
	for i, p := range input {
		output[i] = Sigmoid(p)
	}
	return output
}
func Sigmoid(input float64) float64 {
	return 1 / (1 + math.Pow(math.E, -1*input))
}

func AddConst(input []float64, cons float64) []float64 {
	output := make([]float64, len(input))
	for i, inp := range input {
		output[i] = inp + cons
	}
	return output
}
func MulConst(input []float64, cons float64) []float64 {
	output := make([]float64, len(input))
	for i, inp := range input {
		output[i] = cons * inp
	}
	return output
}
func AddArrays(input1 []float64, input2 []float64) []float64 {
	if len(input1) != len(input2) {
		panic("AddArrays Error, the arrays' sizes are unequal")
	}
	output := make([]float64, len(input1))
	for i, inp := range input1 {
		output[i] = inp + input2[i]
	}
	return output
}
