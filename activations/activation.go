package activations

import (
	"github.com/ldsec/lattigo/v2/ckks"
)

type Activation interface {
	Forward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext
	Backward(input []*ckks.Ciphertext, inputLength int) []*ckks.Ciphertext
	GetForwardLevelConsumption() int
	GetBackwardLevelConsumption() int
}