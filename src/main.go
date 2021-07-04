package main

import (
	// "fmt"
	"math"

	"github.com/perm-ai/GO-HEML-prototype/src/importer"
	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

func main() {

	// lrData := importer.GetSeaData("./test-data/sea_data.json")
	lrData := importer.GetHousingData("./test-data/housing_data.json")
	utils := utility.NewUtils(math.Pow(2, 35), 0, true, true)
	linearRegression := ml.NewLinearRegression(utils)

	x := utils.Encrypt(lrData.Income)
	y := utils.Encrypt(lrData.Value)

	linearRegression.Train(&x, &y, 0.7, len(lrData.Income), 65)

}
