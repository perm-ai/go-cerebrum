package main

import (

	//"math"
	// "fmt"

	//"math"

	// "github.com/perm-ai/GO-HEML-prototype/src/importer"
	"fmt"

	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

func main() {

	// lrData := importer.GetSeaData("./test-data/sea_data.json")
	// lrData := importer.GetHousingData("./test-data/housing_data.json")
	// utils := utility.NewUtils(math.Pow(2, 35), 0, true, true)
	// linearRegression := ml.NewLinearRegression(utils)

	// x := utils.Encrypt(lrData.Income)
	// y := utils.Encrypt(lrData.Value)

	// linearRegression.Train(&x, &y, 0.7, len(lrData.Income), 65)
	// lrData := importer.GetHeartData("./test-data/heart_data_test2.json")
	// x := lrData.Age
	// y := lrData.Trestbps
	// target := lrData.Target
	// ml.Normalize_Data(x)
	// ml.Normalize_Data(y)
	// logisticRegression := ml.NewLogisticRegression()
	// ml.Train(logisticRegression, x, y, target, 0.1, 20)
	Acc := 0.0
	for i := 0; i < 10; i++ {
		x, y, target := utility.GenerateLinearData(300)
		logisticRegression := ml.NewLogisticRegression()
		Acc += ml.Train(logisticRegression, x, y, target, 0.1, 20)
	}
	fmt.Printf("Average Accuracy : %f", Acc/10)

}
