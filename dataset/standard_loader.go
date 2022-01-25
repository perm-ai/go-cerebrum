package dataset

import "github.com/ldsec/lattigo/v2/ckks"

type StandardLoader struct {
	X []*ckks.Ciphertext
	Y []*ckks.Ciphertext
}

func NewStandardLoader(dataX map[string]*ckks.Ciphertext, order []string, dataY []*ckks.Ciphertext) StandardLoader {

	x := make([]*ckks.Ciphertext, len(order))

	for i := range order {

		x[i] = dataX[order[i]]

	}

	return StandardLoader{X: x, Y: dataY}

}

func (s StandardLoader) GetLength() int {
	return len(s.X)
}

func (s StandardLoader) Load1D(start int, batchSize int) ([]*ckks.Ciphertext, []*ckks.Ciphertext) {

	return s.X[start : start+batchSize], s.Y[start : start+batchSize]

}

func (s StandardLoader) Load2D(start int, batchSize int) ([][][]*ckks.Ciphertext, []*ckks.Ciphertext) {

	// TODO: implement later
	return [][][]*ckks.Ciphertext{}, []*ckks.Ciphertext{}

}
