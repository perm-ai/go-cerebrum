package activations

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type Activation interface {
	Forward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext
	Backward(input []*rlwe.Ciphertext, inputLength int) []*rlwe.Ciphertext
	GetForwardLevelConsumption() int
	GetBackwardLevelConsumption() int
	GetType() string
}