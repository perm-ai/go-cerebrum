package main

import (

	// "math"

	"github.com/perm-ai/GO-HEML-prototype/src/importer"
	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	// "github.com/perm-ai/GO-HEML-prototype/src/utility"
)

func main() {

	// lrData := importer.GetSeaData("./test-data/sea_data.json")
	// lrData := importer.GetHousingData("./test-data/housing_data.json")
	// utils := utility.NewUtils(math.Pow(2, 35), 0, true, true)
	// linearRegression := ml.NewLinearRegression(utils)

	// x := utils.Encrypt(lrData.Income)
	// y := utils.Encrypt(lrData.Value)

	// linearRegression.Train(&x, &y, 0.7, len(lrData.Income), 65)
	lrData := importer.GetHeartData("./test-data/heart.json")
	x := lrData.Age
	y := lrData.Sex
	target := lrData.Target

	logisticRegression := ml.NewLogisticRegression()

	//fmt.Println(ml.Predict(logisticRegression, x, y))

	ml.Coefficients_Sgd(x, y, target, logisticRegression, 0.5, 20)
}
