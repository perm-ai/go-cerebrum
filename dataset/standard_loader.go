package dataset

import (
	"sync"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/utility"
)

type StandardLoader struct {
	u      utility.Utils
	X      []*rlwe.Ciphertext
	Y      []*rlwe.Ciphertext
	Length int
}

func NewStandardLoader(dataX map[string]*rlwe.Ciphertext, order []string, dataY []*rlwe.Ciphertext, utils utility.Utils, length int) StandardLoader {

	x := make([]*rlwe.Ciphertext, len(order))

	for i := range order {

		x[i] = dataX[order[i]]

	}

	return StandardLoader{X: x, Y: dataY, Length: length, u: utils}

}

func (s StandardLoader) GetLength() int {
	return s.Length
}

func (s StandardLoader) Load1D(start int, batchSize int) ([]*rlwe.Ciphertext, []*rlwe.Ciphertext) {

	filter := make([]float64, s.Length)
	batchedX := make([]*rlwe.Ciphertext, len(s.X))
	batchedY := make([]*rlwe.Ciphertext, len(s.Y))

	for i := range filter {
		if i >= start && i < start+batchSize {
			filter[i] = 1
		} else {
			filter[i] = 0
		}
	}

	filterPlain := s.u.EncodePlaintextFromArray(filter)

	var xWg sync.WaitGroup
	var yWg sync.WaitGroup

	for xi := range batchedX {

		xWg.Add(1)

		go func(x *rlwe.Ciphertext, index int, utils utility.Utils) {
			defer xWg.Done()
			batchedX[index] = utils.MultiplyPlainNew(x, filterPlain, true, false)
		}(s.X[xi].CopyNew(), xi, s.u.CopyWithClonedEval())

	}

	for yi := range batchedY {

		yWg.Add(1)

		go func(y *rlwe.Ciphertext, index int, utils utility.Utils) {
			defer yWg.Done()
			batchedY[index] = utils.MultiplyPlainNew(y, filterPlain, true, false)
		}(s.Y[yi].CopyNew(), yi, s.u.CopyWithClonedEval())

	}

	xWg.Wait()
	yWg.Wait()

	return batchedX, batchedY

}

func (s StandardLoader) Load2D(start int, batchSize int) ([][][]*rlwe.Ciphertext, []*rlwe.Ciphertext) {

	// TODO: implement later
	return [][][]*rlwe.Ciphertext{}, []*rlwe.Ciphertext{}

}
