package dataset

import (
	"sync"

	"github.com/tuneinsight/lattigo/v4/rlwe"
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

func (m MnistLoader) Load1D(start int, batchSize int) ([]*rlwe.Ciphertext, []*rlwe.Ciphertext) {

	var xWg, yWg sync.WaitGroup

	x := make([]*rlwe.Ciphertext, 784)
	y := make([]*rlwe.Ciphertext, 10)

	for i := range x {

		xWg.Add(1)

		go func(index int, utils utility.Utils) {

			defer xWg.Done()
			batchX := make([]float64, batchSize)

			for dataIdx := 0; dataIdx < batchSize; dataIdx++ {
				batchX[dataIdx] = m.RawData[dataIdx+start].Image[index]
			}

			x[index] = utils.EncryptToLevel(batchX, 9)

		}(i, m.utils.CopyWithClonedEncryptor())

	}

	for i := range y {

		yWg.Add(1)

		go func(index int, utils utility.Utils) {

			defer yWg.Done()
			batchY := make([]float64, batchSize)

			for dataIdx := 0; dataIdx < batchSize; dataIdx++ {
				batchY[dataIdx] = m.RawData[dataIdx+start].Label[index]
			}

			y[index] = utils.EncryptToLevel(batchY, 9)

		}(i, m.utils.CopyWithClonedEncryptor())

	}

	yWg.Wait()
	xWg.Wait()

	return x, y

}

func (m MnistLoader) Load2D(start int, batchSize int) ([][][]*rlwe.Ciphertext, []*rlwe.Ciphertext) {
	return [][][]*rlwe.Ciphertext{}, []*rlwe.Ciphertext{}
}
