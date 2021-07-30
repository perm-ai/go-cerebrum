package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/perm-ai/go-cerebrum/importer"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/regression"
	"github.com/perm-ai/go-cerebrum/utility"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	utils := utility.Utils{}
	log := logger.NewLogger(true)
	var keyPath string
	if keyPath == "" {
		keysChain := key.GenerateKeys(0, true, true)
		keysChain.DumpKeys("keychain", true, true, false, false, false)
		utils = utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)
	} else {
		keypair := key.LoadKeys(keyPath, 0, true, true, false, false, false)
		keys := key.GenerateKeysFromKeyPair(0, keypair.SecretKey, keypair.PublicKey, true, true)
		utils = utility.NewUtils(keys, math.Pow(2, 35), 0, true)
	}
	csvpath := getStringfromConsole("Input path for data train")
	testpath := getStringfromConsole("Input path for data test")
	lr, err := strconv.ParseFloat(getStringfromConsole("Input learning rate "), 64)
	check(err)
	epoch, err := strconv.Atoi(getStringfromConsole("Input epoch "))
	check(err)
	a := []int{0, 1, 2}
	data := importer.GetCSVNData(csvpath, a, false)
	plaind := regression.NewDataPlain(data[0], data[1], data[2])
	Endata := regression.EncryptData(plaind, utils)
	log.Log("Initializing model")
	model := regression.NewLogisticmodel(utils)
	log.Log("Begin training")
	model.Train(Endata, lr, epoch)
	log.Log("Training complete testing the model")
	dataTest := importer.GetCSVNData(testpath, a, false)
	plainT := regression.NewDataPlain(dataTest[0], dataTest[1], dataTest[2])
	model.LogTest(plainT)
}

func getStringfromConsole(text string) string {

	fmt.Println(text)
	fmt.Print("->")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if scanner.Err() != nil {
		panic("Error, there is no input.")
	}
	return (scanner.Text())

}
