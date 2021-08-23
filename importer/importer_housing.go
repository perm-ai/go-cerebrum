package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type HousingData struct {
	Age    []float64
	Income []float64
	Value  []float64
}

func GetHousingData(filepath string) HousingData {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data HousingData
	json.Unmarshal(file, &data)

	return data
}
