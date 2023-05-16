package losses

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

type Loss interface {
	Forward(pred []*rlwe.Ciphertext, y []*rlwe.Ciphertext, predLength int) []*rlwe.Ciphertext
	Backward(pred []*rlwe.Ciphertext, y []*rlwe.Ciphertext, predLength int) []*rlwe.Ciphertext
}
