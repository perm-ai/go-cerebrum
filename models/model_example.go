package models

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
	"github.com/perm-ai/go-cerebrum/utility"
)


func ModelCreationExample() {

	/*
	_________________________________________________________________
	Layer (type)                 Output Shape              Param #   
	=================================================================
	conv2d (Conv2D)              (None, 30, 30, 32)        896       
	_________________________________________________________________
	max_pooling2d (MaxPooling2D) (None, 15, 15, 32)        0         
	_________________________________________________________________
	conv2d_1 (Conv2D)            (None, 13, 13, 64)        18496     
	_________________________________________________________________
	max_pooling2d_1 (MaxPooling2 (None, 6, 6, 64)          0         
	_________________________________________________________________
	conv2d_2 (Conv2D)            (None, 4, 4, 64)          36928     
	_________________________________________________________________
	flatten (Flatten)            (None, 1024)              0         
	_________________________________________________________________
	dense (Dense)                (None, 64)                65600     
	_________________________________________________________________
	dense_1 (Dense)              (None, 10)                650       
	=================================================================
	*/

	utils := utility.Utils{}
	keysChain := key.GenerateKeys(0, false, true)
	utils = utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)
	lr := 0.1

	var tanh activations.Activation
	tanh = activations.NewTanh(utils)

	var relu activations.Activation
	relu = activations.Relu{utils}

	var smx activations.Activation
	smx = activations.NewSoftmax(utils)

	fmt.Println("Conv2D 1 generating")
	conv1 := layers.NewConv2D(utils, 32, []int{3,3}, []int{1,1}, false, &relu, true, []int{32,32,3}, 30000)
	pool1 := layers.NewPoolingLayer(utils, conv1.GetOutputSize(), []int{2,2}, []int{1,1})

	fmt.Println("Conv2D 2 generating")
	conv2 := layers.NewConv2D(utils, 64, []int{3,3}, []int{1,1}, false, &relu, true, pool1.GetOutputSize(), 30000)
	pool2 := layers.NewPoolingLayer(utils, conv2.GetOutputSize(), []int{2,2}, []int{1,1})

	fmt.Println("Conv2D 3 generating")
	conv3 := layers.NewConv2D(utils, 64, []int{3,3}, []int{1,1}, false, &relu, true, pool2.GetOutputSize(), 30000)

	flatten := layers.NewFlatten(conv3.GetOutputSize())

	fmt.Println("Dense 1 generating")
	dense1 := layers.NewDense(utils, flatten.GetOutputSize(), 64, &tanh, true, 30000, lr, 9)

	fmt.Println("Dense 2 generating")
	dense2 := layers.NewDense(utils, dense1.GetOutputSize(), 10, &smx, true, 30000, lr, 9)

	model := NewModel(utils, []layers.Layer1D{
		&dense1, &dense2,
	}, []layers.Layer2D{
		&conv1, &pool1, &conv2, &pool2, &conv3,
	}, losses.CrossEntropy{U: utils}, true)
	
	fmt.Println(model.ForwardLevel)
	fmt.Println(model.BackwardLevel)

}

func PaperModel(){
	BATCH_SIZE := 2500
	LEARNING_RATE := 0.5
	EPOCH := 1

	keysChain := key.GenerateKeys(0, true, true)
	utils := utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)

	fmt.Println("Loading Data")
	// loader := dataset.NewMnistLoaderSmallBatch(utils, "/usr/local/go/src/github.com/perm-ai/go-cerebrum/importer/test-data/mnist_handwritten_train.json", 1, BATCH_SIZE)
	loader := dataset.NewMnistLoader(utils, "/usr/local/go/src/github.com/perm-ai/go-cerebrum/importer/test-data/mnist_handwritten_train.json")

	var tanh activations.Activation
	tanh = activations.NewTanh(utils)

	var smx activations.Activation
	smx = activations.NewSoftmax(utils)

	fmt.Println("Dense 1 generating")
	dense1 := layers.NewDense(utils, 784, 20, &tanh, true, BATCH_SIZE, LEARNING_RATE, 2)

	fmt.Println("Dense 2 generating")
	dense2 := layers.NewDense(utils, dense1.GetOutputSize(), 10, &smx, true, BATCH_SIZE, LEARNING_RATE, 2)

	dense1.SetBootstrapOutput(true, "forward")

	dense2.SetBootstrapOutput(true, "forward")
	dense2.SetBootstrapOutput(true, "backward")
	dense2.SetBootstrapActivation(true, "forward")

	model := NewModel(utils, []layers.Layer1D{
		&dense1, &dense2,
	}, []layers.Layer2D{}, losses.CrossEntropy{U: utils}, false)

	timer := logger.StartTimer("Neural Network Training")

	model.Train1D(loader, LEARNING_RATE, BATCH_SIZE, EPOCH)

	timer.LogTimeTakenSecond()

	os.Mkdir("test_model", 0777)

	model.ExportModel1D("test_model")
}