package example

import (
	"fmt"
	"math"
	"os"

	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/dataset"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/layers"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/losses"
	"github.com/perm-ai/go-cerebrum/models"
	"github.com/perm-ai/go-cerebrum/utility"
)


func MNISTNeuralNetwork(datasetPath string){

	// Training parameters
	BATCH_SIZE := 2500
	LEARNING_RATE := 0.5
	EPOCH := 1

	// Generate keys and util struct
	keysChain := key.GenerateKeys(0, true, true)
	utils := utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)

	// Initialize data loader
	fmt.Println("Loading Data")
	loader := dataset.NewMnistLoader(utils, datasetPath)

	// Initialize activation functions
	var tanh activations.Activation
	tanh = activations.NewTanh(utils)

	var smx activations.Activation
	smx = activations.NewSoftmax(utils)

	// Initialize layers
	fmt.Println("Dense 1 generating")
	dense1 := layers.NewDense(utils, 784, 20, &tanh, true, BATCH_SIZE, LEARNING_RATE, 2)

	fmt.Println("Dense 2 generating")
	dense2 := layers.NewDense(utils, dense1.GetOutputSize(), 10, &smx, true, BATCH_SIZE, LEARNING_RATE, 2)

	// Configure bootstrapping for each layers
	dense1.SetBootstrapOutput(true, "forward")

	dense2.SetBootstrapOutput(true, "forward")
	dense2.SetBootstrapOutput(true, "backward")
	dense2.SetBootstrapActivation(true, "forward")

	// Initialize the model
	model := models.NewModel(utils, []layers.Layer1D{
		&dense1, &dense2,
	}, []layers.Layer2D{}, losses.CrossEntropy{U: utils}, false)

	timer := logger.StartTimer("Neural Network Training")

	// Train the model
	model.Train1D(loader, LEARNING_RATE, BATCH_SIZE, EPOCH)

	// Print the time taken
	timer.LogTimeTakenSecond()

	// Save trained parameters
	os.Mkdir("test_model_1", 0777)

	model.ExportModel1D("test_model_1")

}	