# GO-HEML-prototype

This repository contains the implementation of the library [Lattigo](https://github.com/ldsec/lattigo) for machine learning training
currently the following machine learning algorithms are implemented

* Linear Regression

## Required dependency

This library require Go lang to run

## Installation

1. Clone or download this repository to ```/go/src/github.com/perm-ai/{repo name}```
1. Download required dependencies using ```go get ./...```

## Usage

The following is the example usage of this repository

```go

// in main.go
package main

import (
  "math/rand"
  "github.com/perm-ai/go-cerebrum/utility"
  "fmt"
  "time"
)

func main(){

  // Set seed for random number generator
  rand.Seed(time.Now().UnixNano())

  // Initialize Utils: tool for doing HE operations
  
  // Function takes in 2 arguments: bootstrappingEnabled (bool), logEnabled(bool)
  utils := utility.NewUtils(true, true)
  
  // Generate number array
  randoms := utils.GenerateFilledArray(rand.Float64() * 5)
  twos := utils.GenerateFilledArray(2.0)
  pis := utils.GenerateFilledArray(3.14159)
  
  // Encrypt float arrays
  randomCiphertext := utils.Encrypt(randoms)
  twosCiphertext := utils.Encrypt(twos)

  // Encode Pi
  encodedPi := utils.Encode(pis)
  
  // Add (ciphertext)
  result := utils.AddNew(randomCiphertext, twosCiphertext)
  
  // Multiply and save result to sum (ciphertext with plaintext)
  utils.MultiplyPlain(&result, &encodedPi, &result, true, true)
  
  // Decryption
  decrypted := utils.Decrypt(&result)
  
  // Print out answer
  fmt.Println(decrypted)

}

```

