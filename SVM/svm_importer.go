package SVM

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type BreastCancer struct {
	texture_mean   []float64
	concavity_mean []float64
	symmetry_mean  []float64
}

func GetBreastCancerData(filepath string) BreastCancer {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data BreastCancer
	json.Unmarshal(file, &data)

	return data
}
