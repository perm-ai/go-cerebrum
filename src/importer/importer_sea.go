package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type SeaData struct {
	Temp []float64
	Sal  []float64
}

func GetSeaData(filepath string) SeaData {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data SeaData
	json.Unmarshal(file, &data)

	return data
}