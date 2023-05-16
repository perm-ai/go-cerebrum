package dataset

import "github.com/tuneinsight/lattigo/v4/rlwe"

type Loader interface {
	GetLength() int
	Load1D(start int, batchSize int) ([]*rlwe.Ciphertext, []*rlwe.Ciphertext)
	Load2D(start int, batchSize int) ([][][]*rlwe.Ciphertext, []*rlwe.Ciphertext)
}
