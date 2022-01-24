package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type CoalData struct {
	TM_AR       []float64
	TS_AR       []float64
	M_AD        []float64
	ASH_AD      []float64
	ASH_AR      []float64
	Sulfate_SO3 []float64
	Silica_SiO2 []float64
	Calcium_CaO []float64
	Iron_Fe2O3  []float64
	Price       []float64
}

func GetCoalData(filepath string) CoalData {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data CoalData
	json.Unmarshal(file, &data)

	return data
}
