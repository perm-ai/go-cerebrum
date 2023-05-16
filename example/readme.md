# Example usage of this library

## Linear Regression

The code below show how to train a linear regression model on the [this]("https://www.kaggle.com/datasets/camnugent/california-housing-prices") data.

```go
// main.go

package main

import (
	"github.com/perm-ai/go-cerebrum/example"
)

func main() {

    fileName := "./housing.csv";
    saveLocation := "./result";

    featureColumn := 7;
    labelColumn := 8;
    lr := 0.8;
    epoch := 30

    example.LinearRegression("", fileName, featureColumn, labelColumn, lr, epoch, saveLocation);

}
```

---

## Train a Neural Network on MNIST dataset

Download the mnist dataset using the command provided in the `download_mnist.sh` file. Use the `MNISTNeuralNetwork` function provided in this module and provide the path to the training data file.