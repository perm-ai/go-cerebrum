package main

import (

	//"math"
	//"fmt"

	"fmt"

	"github.com/perm-ai/GO-HEML-prototype/src/importer"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"

	// "fmt"

	"github.com/perm-ai/GO-HEML-prototype/src/ml"
)

func main() {

	// lrData := importer.GetTitanicData("./test-data/titanic1.json")
	// x := lrData.Age
	// y := lrData.Pclass
	// target := lrData.Target
	// ml.Normalize_Data(x)
	// ml.Normalize_Data(y)
	// logisticRegression := ml.NewLogisticRegression()
	// ml.Train(logisticRegression, x, y, target, 0.1, 20)

	// Acc := 0.0
	// for i := 0; i < 10; i++ {
	// 	x, y, target := utility.GenerateLinearData(300)
	// 	logisticRegression := ml.NewLogisticRegression()
	// 	Acc += ml.Train(logisticRegression, x, y, target, 0.1, 20)
	// }
	// fmt.Printf("Average Accuracy : %f", Acc/10)
	// datasetSize := 300
	// utils := utility.NewUtils(math.Pow(2, 35), 0, true, true)
	// data1, data2, data3 := utility.GenerateLinearData(datasetSize)
	// logisticRegression := ml.NewLogisticRegression(utils)

	datasetSize := 713
	utils := utility.NewUtils(35, 0, true, true)
	// data1, data2, data3 := utility.GenerateLinearData(datasetSize)
	// logisticRegression := ml.NewLogisticRegression(utils)

	// x := utils.Encrypt(data1)
	// y := utils.Encrypt(data2)
	// target := utils.Encrypt(data3)

	lrData := importer.GetTitanicData("./test-data/titanic1.json")
	data1 := lrData.Age
	data2 := lrData.Pclass
	data3 := lrData.Target
	ml.Normalize_Data(data1)
	ml.Normalize_Data(data2)

	// data1Plain := data1
	// data2Plain := data2
	// data3Plain := data3
	// LogisticRegression2 := ml.NewLogisticRegression2()
	// LogisticRegression2.TrainLRPlain(data1Plain, data2Plain, data2Plain, 0.01, 713, 20)

	// LogisticRegression2.AccuracyTestPlain(data1Plain, data2Plain, data3Plain, 713)

	// accuracyPlain := LogisticRegression2.AccuracyTestPlain(data1Plain, data2Plain, data3Plain, 713)
	// fmt.Printf("The accuracy of this plain logistic regression model is %f percent \n", accuracyPlain)

	x := utils.Encrypt(data1)
	y := utils.Encrypt(data2)
	target := utils.Encrypt(data3)

	logisticRegression := ml.NewLogisticRegression(utils)

	logisticRegression.TrainLR(x, y, target, 0.1, datasetSize, 5)

	accuracy := logisticRegression.AccuracyTest(data1, data2, data3, datasetSize)
	fmt.Printf("The accuracy of this logistic regression model is %f percent \n", accuracy)
}
