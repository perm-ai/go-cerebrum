package svm

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type BreastCancer struct {
	Texture_mean   []float64
	Concavity_mean []float64
	Symmetry_mean  []float64
	Diagnosis      []float64
}

func GetBreastCancerData(filepath string) BreastCancer {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data BreastCancer
	json.Unmarshal(file, &data)

	return data
}
