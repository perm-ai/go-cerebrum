package main

import (
	"fmt"

	"github.com/perm-ai/GO-HEML-prototype/src/importer"
	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

func main() {

	lrData := importer.GetSeaData("./test-data/sea_data.json")
	utils := utility.NewUtils(true, true)
	linearRegression := ml.NewLinearRegression(utils)

	x := utils.Encrypt(lrData.Temp[0:32768])
	y := utils.Encrypt(lrData.Sal[0:32768])
	linearRegression.Train(&x, &y, 0.1, len(lrData.Temp), 5)
	m := utils.Decrypt(&linearRegression.M)
	b := utils.Decrypt(&linearRegression.B)
	fmt.Printf("Result M: %f B: %f", m, b)

}
