package regression

import (
	"fmt"
	"math"
	"testing"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/importer"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/utility"
)

func TestLinearRegression(t *testing.T) {

	// Insert data path here
	data := importer.GetHousingData("/usr/local/go/src/github.com/perm-ai/go-cerebrum/importer/test-data/housing_test.json")

	fmt.Println(len(data.Housing_median_age))
	fmt.Println(len(data.Median_income))
	fmt.Println(len(data.Median_house_value))

	utils := utility.Utils{}
	keysChain := key.GenerateKeys(0, true, true)
	utils = utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)

	data1 := utils.EncryptToPointer(data.Housing_median_age)
	data2 := utils.EncryptToPointer(data.Median_income)
	data3 := utils.EncryptToPointer(data.Median_house_value)
	model := NewLinearRegression(utils, 2)

	// utils.Evaluator.DropLevel(data1, data1.Level()-9)
	// utils.Evaluator.DropLevel(data2, data2.Level()-9)
	// utils.Evaluator.DropLevel(data3, data3.Level()-9)

	independentVar := []*rlwe.Ciphertext{data1, data2}
	model.Train(independentVar, data3, 0.1, len(data.Median_income), 20)
	slope := make([]float64, 2)
	for i := 0; i < 2; i++ {
		slope[i] = utils.Decrypt(model.Weight[i])[0]
	}
	bias := utils.Decrypt(model.Bias)

	fmt.Printf("The weights are biases are %f and %f", slope, bias[0])

}
