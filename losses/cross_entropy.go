package losses

import (
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

type CrossEntropy struct {
	U utility.Utils
}

func (c CrossEntropy) Forward(pred []*rlwe.Ciphertext, y []*rlwe.Ciphertext, predLength int) []*rlwe.Ciphertext {
	return pred
}

func (c CrossEntropy) Backward(pred []*rlwe.Ciphertext, y []*rlwe.Ciphertext, predLength int) []*rlwe.Ciphertext {
	result := make([]*rlwe.Ciphertext, len(pred))

	for i := range pred{
		result[i] = c.U.SubNew(pred[i], y[i])
	}

	return result
}
