package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type HousingData struct {
	Housing_median_age    	[]float64
	Median_income 			[]float64
	Median_house_value  	[]float64
}

func GetHousingData(filepath string) HousingData {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data HousingData
	json.Unmarshal(file, &data)

	return data
}
