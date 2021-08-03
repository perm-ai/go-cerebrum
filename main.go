package main

import (
	"fmt"

	"github.com/perm-ai/go-cerebrum/svm"
)

func main() {
	rawData := svm.GetBreastCancerData("./svm/SVM_dataset.json")
	data1 := rawData.Texture_mean
	data2 := rawData.Concavity_mean
	data3 := rawData.Symmetry_mean
	target := rawData.Diagnosis
	numOfData := len(rawData.Texture_mean)
	fmt.Printf("numOfData = %d \n", numOfData)
	data := make([][]float64, numOfData)
	for i := 0; i < len(rawData.Texture_mean); i++ {
		data[i] = make([]float64, 4)
		data[i][0] = data1[i]
		data[i][1] = data2[i]
		data[i][2] = data3[i]
		data[i][3] = 1

	}
	model := svm.NewSVMModel()
	weights := model.TrainSVM(data, target, 20, 0.01, 3, 1000)
	fmt.Printf("The weights of the model are %f \n", weights)
}
