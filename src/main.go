package main

import (

	"github.com/perm-ai/GO-HEML-prototype/src/utility"
	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	"github.com/perm-ai/GO-HEML-prototype/src/importer"

)



func main() {

	lrData := importer.GetSeaData("./test-data/sea_data.json")
	utils := utility.NewUtils(true, true)
	linearRegression := ml.NewLinearRegression(utils)
	
	x := utils.Encrypt(lrData.Temp)
	y := utils.Encrypt(lrData.Sal)
	linearRegression.Train(&x, &y, 0.1, len(lrData.Temp), 5)

}
