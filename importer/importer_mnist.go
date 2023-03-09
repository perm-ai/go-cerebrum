package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/tuneinsight/lattigo/v4/rlwe"
)


type mnistData struct {
	Image []int
	Label int
}

type MnistData struct {
	Image []float64
	Label []float64
}

type EncryptedMnistData struct {
	Image rlwe.Ciphertext
	Label rlwe.Ciphertext
}

func getMnistData(filepath string) []mnistData {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data []mnistData
	json.Unmarshal([]byte(file), &data)

	return data
}

func GetMnistData(filepath string) []MnistData {

	datas := getMnistData(filepath)
	result := make([]MnistData, len(datas))

	for i := range datas {

		label := make([]float64, 10)
		label[datas[i].Label] = 1

		images := make([]float64, len(datas[i].Image))
		for image := range images{
			images[image] = float64(datas[i].Image[image]) / 255.0
		}

		result[i] = MnistData{images, label}

	}

	return result


}