package losses

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/utility"
)

type CrossEntropy struct {
	U utility.Utils
}

func (c CrossEntropy) Forward(pred []*ckks.Ciphertext, y []*ckks.Ciphertext, predLength int) []*ckks.Ciphertext {
	return pred
}

func (c CrossEntropy) Backward(pred []*ckks.Ciphertext, y []*ckks.Ciphertext, predLength int) []*ckks.Ciphertext {
	result := make([]*ckks.Ciphertext, len(pred))

	for i := range pred{
		result[i] = c.U.SubNew(pred[i], y[i])
	}

	return result
}
