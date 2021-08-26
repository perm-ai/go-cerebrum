package dataset

import "github.com/ldsec/lattigo/v2/ckks"

type Loader interface {
	GetLength() int
	Load1D(start int, batchSize int) ([]*ckks.Ciphertext, []*ckks.Ciphertext)
	Load2D(start int, batchSize int) ([][][]*ckks.Ciphertext, []*ckks.Ciphertext)
}
