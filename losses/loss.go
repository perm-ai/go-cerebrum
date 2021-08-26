package losses

import (
	"github.com/ldsec/lattigo/v2/ckks"
)

type Loss interface {
	Forward(pred []*ckks.Ciphertext, y []*ckks.Ciphertext, predLength int) []*ckks.Ciphertext
	Backward(pred []*ckks.Ciphertext, y []*ckks.Ciphertext, predLength int) []*ckks.Ciphertext
}
