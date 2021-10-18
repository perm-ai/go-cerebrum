package dataset

import (
	"sync"

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

func NewMnistLoaderSmallBatch(utils utility.Utils, filePath string, batchAmount int, batchSize int) MnistLoader {

	return MnistLoader{utils, importer.GetMnistData(filePath)[0 : batchSize*batchAmount]}

}

func (m MnistLoader) GetLength() int {
	return len(m.RawData)
}

func (m MnistLoader) Load1D(start int, batchSize int) ([]*ckks.Ciphertext, []*ckks.Ciphertext) {

	var xWg, yWg sync.WaitGroup

	x := make([]*ckks.Ciphertext, 784)
	y := make([]*ckks.Ciphertext, 10)

	for i := range x {

		xWg.Add(1)

		go func(index int, utils utility.Utils) {

			defer xWg.Done()
			batchX := make([]float64, batchSize)

			for dataIdx := start; dataIdx < start+batchSize; dataIdx++ {
				batchX[dataIdx] = m.RawData[dataIdx].Image[index]
			}

			x[index] = utils.EncryptToLevel(batchX, 9)

		}(i, m.utils.CopyWithClonedEncryptor())

	}

	for i := range y {

		yWg.Add(1)

		go func(index int, utils utility.Utils) {

			defer yWg.Done()
			batchY := make([]float64, batchSize)

			for dataIdx := start; dataIdx < start+batchSize; dataIdx++ {
				batchY[dataIdx] = m.RawData[dataIdx].Label[index]
			}

			y[index] = utils.EncryptToLevel(batchY, 9)

		}(i, m.utils.CopyWithClonedEncryptor())

	}

	yWg.Wait()
	xWg.Wait()

	return x, y

}

func (m MnistLoader) Load2D(start int, batchSize int) ([][][]*ckks.Ciphertext, []*ckks.Ciphertext) {
	return [][][]*ckks.Ciphertext{}, []*ckks.Ciphertext{}
}
