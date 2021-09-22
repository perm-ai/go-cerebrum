package main

import (
	"fmt"
	"math"
	"os"

	"github.com/perm-ai/go-cerebrum/activations"
	"github.com/perm-ai/go-cerebrum/dataset"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/layers"
	"github.com/perm-ai/go-cerebrum/losses"
	"github.com/perm-ai/go-cerebrum/models"
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

	BATCH_SIZE := 600
	LEARNING_RATE := 0.1
	EPOCH := 1

	keysChain := key.GenerateKeys(0, false, true)
	utils := utility.NewUtils(keysChain, math.Pow(2, 35), 0, true)

	var tanh activations.Activation
	tanh = activations.NewTanh(utils)

	var smx activations.Activation
	smx = activations.NewSoftmax(utils)

	fmt.Println("Dense 1 generating")
	dense1 := layers.NewDense(utils, 784, 32, &tanh, true, BATCH_SIZE)

	fmt.Println("Dense 2 generating")
	dense2 := layers.NewDense(utils, dense1.GetOutputSize(), 10, &smx, true, BATCH_SIZE)

	model := models.NewModel(utils, []layers.Layer1D{
		&dense1, &dense2,
	}, []layers.Layer2D{}, losses.CrossEntropy{U: utils})

	fmt.Println("Loading Data")
	loader := dataset.NewMnistLoader(utils, "./importer/test-data/mnist_handwritten_train.json")
	
	model.Train1D(loader, LEARNING_RATE, BATCH_SIZE, EPOCH)

	os.Mkdir("test_model_weight", 0777)

	model.ExportModel1D("test_model_weight")

}
