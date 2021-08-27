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

	data := importer.GetHousingData("importer/test-data/housing_test.json")

	fmt.Println(len(data.Housing_median_age))
	fmt.Println(len(data.Median_income))
	fmt.Println(len(data.Median_house_value))

	utils := utility.Utils{}
	keysChain := key.GenerateKeys(0, false, true)
	utils = utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)

	data1 := utils.Encrypt(data.Housing_median_age)
	data2 := utils.Encrypt(data.Median_income)
	data3 := utils.Encrypt(data.Median_house_value)
	model := regression.NewLinearRegression(utils, 2)
	independentVar := []ckks.Ciphertext{data1, data2}
	model.Train(independentVar, &data3, 0.1, len(data.Median_income), 20)
	slope := make([]float64, 2)
	for i := 0; i < 2; i++ {
		slope[i] = utils.Decrypt(&model.Weight[i])[0]
	}
	bias := utils.Decrypt(&model.Bias)

	fmt.Printf("The weights are biases are %f and %f", slope, bias[0])
}
