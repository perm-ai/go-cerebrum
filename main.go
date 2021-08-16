package main

import (
	"fmt"
	"math"

	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/utility"
)

func main() {

	keyChain := key.GenerateKeys(0, false, true)
	utils := utility.NewUtils(keyChain, math.Pow(2, 35), 0, true)

	ct := utils.Encrypt(utils.GenerateFilledArray(2))

	// for _, ring := range ct.Value{
	// 	fmt.Println(ring.Coeffs)
	// }

	fmt.Println(len(ct.Value[0].Coeffs))
	fmt.Println(len(ct.Value[1].Coeffs))
	fmt.Println(ct.Value[1].Coeffs[:10000])

	// rawData := svm.GetBreastCancerData("./svm/SVM_dataset.json")
	// data1 := rawData.Texture_mean
	// data2 := rawData.Concavity_mean
	// data3 := rawData.Symmetry_mean
	// target := rawData.Diagnosis
	// numOfData := len(rawData.Texture_mean)
	// fmt.Printf("numOfData = %d \n", numOfData)
	// data := make([][]float64, numOfData)
	// for i := 0; i < len(rawData.Texture_mean); i++ {
	// 	data[i] = make([]float64, 4)
	// 	data[i][0] = data1[i]
	// 	data[i][1] = data2[i]
	// 	data[i][2] = data3[i]
	// 	data[i][3] = 1
	// }
	// fmt.Printf("This dataset consists of %d data points and %d features \n", len(data), len(data[0]))
	// fmt.Printf("Target has %d data points\n", len(target))
	// model := svm.NewSVMModel()
	// weights := model.TrainSVM(data, target, 20, 0.001, 4, 10000)
	// fmt.Printf("The weights of the model are %f \n", weights)
}
