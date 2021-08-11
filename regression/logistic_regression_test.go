package regression

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/array"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

func NewDataPlain(x [][]float64, target []float64) DataPlain {
	fmt.Println("Data has coloumn :" + fmt.Sprint(len(x)))
	return DataPlain{x, target}
}
func NewEmptyData(column int, amount int) DataPlain {
	return DataPlain{array.GenFilledArraysofArrays(0, column, amount), make([]float64, amount)}

}

func EncryptData(data DataPlain, utils utility.Utils) Data {
	log := logger.NewLogger(true)
	EnData := make([]ckks.Ciphertext, len(data.x))
	for i, p := range data.x {
		log.Log("Encrypting Column" + fmt.Sprint(i))
		EnData[i] = utils.Encrypt(p)
	}
	log.Log("Encrypting target")
	enctar := utils.Encrypt(data.target)
	log.Log("Encryption complete")
	return Data{EnData, enctar, len(data.x[0])}

}
func generateData(dat int, testdata int, columnAmount int) (DataPlain, DataPlain) {

	d := NewEmptyData(columnAmount, dat)
	dT := NewEmptyData(columnAmount, testdata)
	rand.Seed(time.Now().UnixNano())
	w := make([]float64, columnAmount)
	for i := 0; i < columnAmount; i++ {
		w[i] = (rand.Float64() * 2) - 1
		if w[i] < 0.1 && w[i] > -0.1 {
			w[i] = w[i] * 3
		}
	}
	c := (rand.Float64() * 2) - 1.2
	// fmt.Println(w)
	// fmt.Println(c)
	ones := 0
	zeros := 0
	for i := 0; i < dat; i++ {
		result := 0.0
		for j := 0; j < columnAmount; j++ {
			d.x[j][i] = rand.Float64()
			result += w[j] * d.x[j][i]
		}
		result += c
		if result > 0 {
			d.target[i] = 1
			ones++

		} else {
			d.target[i] = 0
			zeros++
		}
	}
	for i := 0; i < testdata; i++ {
		result := 0.0
		for j := 0; j < columnAmount; j++ {
			dT.x[j][i] = rand.Float64()
			result += w[j] * dT.x[j][i]
		}
		result += c
		if result > 0 {
			dT.target[i] = 1
			ones++

		} else {
			dT.target[i] = 0
			zeros++
		}
	}

	if !(ones < 2*zeros && zeros < ones) && !(zeros < 2*ones && ones < zeros) {
		d, dT = generateData(dat, testdata, columnAmount)

	}
	fmt.Printf("1 : %d , 0 : %d\n", ones, zeros)
	return d, dT

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
	// data := importer.GetCSVNData(csvpath, a, false)

	log.Log("lr = 0.1,epoch = 100")
	data, dataTest := generateData(1000, 100, 2)
	log.Log("There are " + fmt.Sprint(len(data.x)) + "column")
	plaind := NewDataPlain(data.x, data.target)
	Endata := EncryptData(plaind, utils)
	log.Log("Initializing model")
	model := NewLogisticRegression(utils, 2)
	log.Log("Begin training")
	model.Train(Endata, 0.1, 5, true)
	log.Log("Training complete testing the model")
	plainT := NewDataPlain(dataTest.x, dataTest.target)
	model.LogTest(plainT)
}
