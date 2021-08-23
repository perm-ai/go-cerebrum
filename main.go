package main

import (
	"fmt"
	"math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/importer"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/regression"
	"github.com/perm-ai/go-cerebrum/utility"
)

func main() {

	utils := utility.Utils{}
	keysChain := key.GenerateKeys(0, true, true)
	utils = utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)

	data := importer.GetHousingData("importer/test-data/housing_test.json")

	fmt.Print(len(data.Value))
	data1 := utils.Encrypt(data.Age)
	data2 := utils.Encrypt(data.Income)
	data3 := utils.Encrypt(data.Value)
	model := regression.NewLinearRegression(utils, 2)
	independentVar := []ckks.Ciphertext{data1, data2}
	model.Train(independentVar, &data3, 0.1, len(data.Income), 20)
	slope := make([]float64, 2)
	for i := 0; i < 2; i++ {
		slope[i] = utils.Decrypt(&model.M[i])[0]
	}
	bias := utils.Decrypt(&model.B)

	fmt.Printf("The weights are biases are %f and %f", slope, bias[0])
}
