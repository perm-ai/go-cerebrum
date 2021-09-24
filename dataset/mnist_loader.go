package dataset

import (
	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/importer"
	"github.com/perm-ai/go-cerebrum/utility"
)

type MnistLoader struct {
	utils   utility.Utils
	RawData []importer.MnistData
}

func NewMnistLoader(utils utility.Utils, filePath string) MnistLoader {

	return MnistLoader{utils, importer.GetMnistData(filePath)}

}

func (m MnistLoader) GetLength() int {
	return len(m.RawData)
}

func (m MnistLoader) Load1D(start int, batchSize int) ([]*ckks.Ciphertext, []*ckks.Ciphertext) {

	x := make([]*ckks.Ciphertext, 784)
	y := make([]*ckks.Ciphertext, 10)

	for i := range x {

		batchX := make([]float64, batchSize)

		for dataIdx := range m.RawData[start : start+batchSize] {
			batchX[dataIdx] = m.RawData[dataIdx].Image[i]
		}

		x[i] = m.utils.EncryptToPointer(batchX)

	}

	for i := range y {

		batchY := make([]float64, batchSize)

		for dataIdx := range m.RawData[start : start+batchSize] {
			batchY[dataIdx] = m.RawData[dataIdx].Label[i]
		}

		y[i] = m.utils.EncryptToPointer(batchY)

	}

	return x, y

}

func (m MnistLoader) Load2D(start int, batchSize int) ([][][]*ckks.Ciphertext, []*ckks.Ciphertext) {
	return [][][]*ckks.Ciphertext{}, []*ckks.Ciphertext{}
}
