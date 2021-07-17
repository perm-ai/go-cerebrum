package cmd

import (
	"fmt"
	"math"
	"os"

	"github.com/perm-ai/GO-HEML-prototype/src/importer"
	"github.com/perm-ai/GO-HEML-prototype/src/logger"
	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	"github.com/perm-ai/GO-HEML-prototype/src/utility"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func LinearRegression(key string, csv string, x int, y int, lr float64, epoch int, dest string) {

	log := logger.NewLogger(true)
	utils := utility.Utils{}

	if key == "" {
		utils = utility.NewUtils(math.Pow(2, 40), 0, true, true)
	} else {
		keys := utility.LoadKeyPair(key)
		utils = utility.NewUtilsFromKeyPair(keys, math.Pow(2, 40), 0, true, true)
	}

	if csv == ""{
		panic("No csv filepath provided")
	}

	log.Log("Importing data from csv")
	data := importer.GetCSV(csv, x, y)

	log.Log("Encrypting X")
	encX := utils.Encrypt(data.FirstData)
	encXbin, _ := encX.MarshalBinary()
	log.Log(fmt.Sprintf("Encrypted X [%f %f . . . %f %f] => [%b %b . . . %b %b]", 
	data.FirstData[0], data.FirstData[1], data.FirstData[len(data.FirstData) - 2], data.FirstData[len(data.FirstData) - 1], 
	encXbin[0], encXbin[1], encXbin[len(encXbin) - 2], encXbin[len(encXbin) - 2]))

	log.Log("Encrypting Y")
	encY := utils.Encrypt(data.SecondData)
	encYbin, _ := encY.MarshalBinary()
	log.Log(fmt.Sprintf("Encrypted Y [%f %f . . . %f %f] => [%b %b . . . %b %b]", 
	data.SecondData[0], data.SecondData[1], data.SecondData[len(data.FirstData) - 2], data.SecondData[len(data.FirstData) - 1], 
	encYbin[0], encYbin[1], encYbin[len(encYbin) - 2], encYbin[len(encYbin) - 2]))

	log.Log("Initializing model")
	model := ml.NewLinearRegression(utils)

	log.Log("Begin training")
	model.Train(&encX, &encY, lr, len(data.FirstData), epoch)

	log.Log("Training complete, saving gradient")
	os.Mkdir(dest, 0777)

	mBytes, mByteErr := model.M.MarshalBinary()
	check(mByteErr)

	mFile, mFileErr := os.Create(dest+"/m")
	check(mFileErr)

	_, mWriteErr := mFile.Write(mBytes)
	check(mWriteErr)

	bBytes, bByteErr := model.B.MarshalBinary()
	check(bByteErr)

	bFile, bFileErr := os.Create(dest+"/b")
	check(bFileErr)

	_, bWriteErr := bFile.Write(bBytes)
	check(bWriteErr)

	log.Log("Training result saved")

}