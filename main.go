package main

import (
	"fmt"
	"math"

	"github.com/perm-ai/go-cerebrum/activations"x 
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/layers"
	"github.com/perm-ai/go-cerebrum/losses"
	"github.com/perm-ai/go-cerebrum/models"
	"github.com/perm-ai/go-cerebrum/utility"
)

func main() {
	keyPair := key.LoadKeys("/usr/local/go/src/github.com/perm-ai/go-cerebrum/keychain", 0, true, true, false, false)
	keychain := key.GenerateKeysFromKeyPair(0, keyPair.SecretKey, keyPair.PublicKey, true, true)

	utils := utility.NewUtils(keychain, math.Pow(2, 35), 0, true)

	batchSize := 40

	var relu activations.Activation
	relu = activations.Relu{}

	fmt.Println("Dense 1 generating")

	// func layers.NewDense(inputUnit int, outputUnit int, activation *util.Activation, useBias bool, batchSize int) layers.Dense
	// func NewDense(utils utility.Utils, inputUnit int, outputUnit int, activation *activations.Activation, useBias bool, batchSize int, lr float64, weightLevel int) Dense
	dense1 := layers.NewDense(utils, 9, 50, &relu, true, batchSize, 0.3, 9)

	fmt.Println("Dense 2 generating")
	dense2 := layers.NewDense(utils, dense1.GetOutputSize(), 1, nil, true, batchSize, 0.3, dense1.GetWeightLevel())

	model := models.NewModel(utils, []layers.Layer1D{&dense1, &dense2}, []layers.Layer2D{}, losses.MSE{}, true)

}