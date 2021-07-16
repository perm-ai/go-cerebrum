package importer

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type TitanicData struct {

	// Not sure if we should be using Age & Sex as float because they are all unsighned ints
	// (not sure if it'll mess something up in the process of using ckks scheme)
	// Basic Logistic Regression is basically evualating linear regression then inputting the data
	// into Sigmoid which will turn the output into a probability (0,1) in which outputs >0.5 will
	// result in class A and <0.5 in class B (classification between A and B)

	Age      []float64
	Pclass   []float64
	Target   []float64
}

func GetTitanicData(filepath string) TitanicData {
	jsonFile, _ := os.Open(filepath)
	defer jsonFile.Close()
	file, _ := ioutil.ReadAll(jsonFile)

	var data TitanicData
	json.Unmarshal(file, &data)

	return data
}
