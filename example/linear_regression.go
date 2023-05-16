package example

import (
	"fmt"
	"math"
	"os"

	"github.com/perm-ai/go-cerebrum/importer"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/regression"
	"github.com/perm-ai/go-cerebrum/utility"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func LinearRegression(keyPath string, csv string, x int, y int, lr float64, epoch int, dest string) {

	// Initialize logger
	log := logger.NewLogger(true)
	utils := utility.Utils{}

	// Initialize key
	if keyPath == "" {
		keysChain := key.GenerateKeys(0, true, true)
		keysChain.DumpKeys("keychain", true, true, false, false, false)
		utils = utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)
	} else {
		keypair := key.LoadKeys(keyPath, 0, true, true, false, false)
		keys := key.GenerateKeysFromKeyPair(0, keypair.SecretKey, keypair.PublicKey, true, true)
		utils = utility.NewUtils(keys, math.Pow(2, 35), 0, true)
	}

	if csv == "" {
		panic("No csv filepath provided")
	}

	// Import csv data
	log.Log("Importing data from csv")
	data := importer.GetCSV(csv, x, y)

	// Encrypt feature 1
	log.Log("Encrypting X")
	feature1 := utils.Encrypt(data.FirstData)
	encX := []*rlwe.Ciphertext{&feature1}
	encXbin, _ := encX[0].MarshalBinary()

	// Log data
	log.Log(fmt.Sprintf("Encrypted X [%f %f . . . %f %f] => [%b %b . . . %b %b]",
		data.FirstData[0], data.FirstData[1], data.FirstData[len(data.FirstData)-2], data.FirstData[len(data.FirstData)-1],
		encXbin[0], encXbin[1], encXbin[len(encXbin)-2], encXbin[len(encXbin)-2]))

	// Encrypt label
	log.Log("Encrypting Y")
	encY := utils.Encrypt(data.SecondData)
	encYbin, _ := encY.MarshalBinary()
	log.Log(fmt.Sprintf("Encrypted Y [%f %f . . . %f %f] => [%b %b . . . %b %b]",
		data.SecondData[0], data.SecondData[1], data.SecondData[len(data.FirstData)-2], data.SecondData[len(data.FirstData)-1],
		encYbin[0], encYbin[1], encYbin[len(encYbin)-2], encYbin[len(encYbin)-2]))

	// Initialize linear regression model
	log.Log("Initializing model")
	model := regression.NewLinearRegression(utils, 1)

	// Begin training the model
	log.Log("Begin training")
	model.Train(encX, &encY, lr, len(data.FirstData), epoch)

	log.Log("Training complete, saving gradient")
	os.Mkdir(dest, 0777)

	// Print the trained weight
	fmt.Println(utils.Decrypt(model.Weight[0])[0])
	fmt.Println(utils.Decrypt(model.Bias)[0])

	// Save training results
	mBytes, mByteErr := model.Weight[0].MarshalBinary()
	check(mByteErr)

	mFile, mFileErr := os.Create(dest + "/m")
	check(mFileErr)

	_, mWriteErr := mFile.Write(mBytes)
	check(mWriteErr)

	bBytes, bByteErr := model.Bias.MarshalBinary()
	check(bByteErr)

	bFile, bFileErr := os.Create(dest + "/b")
	check(bFileErr)

	_, bWriteErr := bFile.Write(bBytes)
	check(bWriteErr)

	log.Log("Training result saved")
}
