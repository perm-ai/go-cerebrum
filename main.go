package main

import (
	"fmt"
	"math"

	// "github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

// "flag"
// "fmt"
// "os"

// "github.com/perm-ai/go-cerebrum/cmd"

func main() {

	utils := utility.NewUtils(math.Pow(2, 35), 0, false, true)

	ct1 := utils.EncryptToScale(utils.GenerateFilledArray(0.204134), 154742504912642859199365120)
	utils.Evaluator.DropLevel(&ct1, ct1.Level() - 23)

	ct2 := utils.EncryptToScale(utils.GenerateFilledArray(0.119975), 618970019619046239405867008)
	utils.Evaluator.DropLevel(&ct2, ct2.Level() - 21)

	utils.Sub(ct1, ct2, &ct1)
	fmt.Println(utils.Decrypt(&ct1)[0])
	// arguments := os.Args

	// if len(arguments) < 2 {
	// 	fmt.Println("list or count subcommand is required")
	// 	os.Exit(1)
	// }

	// lrCommand := flag.NewFlagSet("linearRegression", flag.ExitOnError)
	// decryptCommand := flag.NewFlagSet("decrypt", flag.ExitOnError)

	// // Flag for lr command
	// keyChainPath := lrCommand.String("key", "", "Load saved keys from given path. If not provided we will generate new keys.")
	// csvPath := lrCommand.String("csv", "", "Path to csv file for encryption of data.")
	// xColumn := lrCommand.Int("x", 0, "Index of column for x value starting at 0")
	// yColumn := lrCommand.Int("y", 1, "Index of column for y value starting at 0")
	// learningRate := lrCommand.Float64("lr", 0.01, "Training learning rate")
	// epoch := lrCommand.Int("epoch", 0, "Training epoch")
	// destination := lrCommand.String("destination", "linear_regression_result", "Destination output save location")

	// // Flag for decrypt command
	// key := decryptCommand.String("key", "", "Load saved keys from given path. If not provided we will generate new keys.")
	// data := decryptCommand.String("data", "", "Path to binary file of data to be decrypted.")

	// switch arguments[1] {
	// case "linearRegression":

	// 	lrCommand.Parse(arguments[2:])
	// 	fmt.Printf("Starting program with config keyChainPath: %s, csvPath: %s, xColumn: %d, yColumn: %d, learningRate: %f, epoch: %d, destination: %s\n", *keyChainPath, *csvPath, *xColumn, *yColumn, *learningRate, *epoch, *destination)

	// 	cmd.LinearRegression(*keyChainPath, *csvPath, *xColumn, *yColumn, *learningRate, *epoch, *destination)

	// case "decrypt":
	// 	decryptCommand.Parse(arguments[2:])
	// 	cmd.Decrypt(*key, *data)
	// }
}
