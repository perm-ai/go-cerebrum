package importer

import (
	"bufio"
	"encoding/csv"

	// "encoding/json"
	// "fmt"
	"io"
	"log"
	"os"
	"strconv"
)

type csvData struct {
	FirstData  []float64
	SecondData []float64
}

func GetCSV(filepath string, colNum1 int, colNum2 int) csvData {

	// path, column number1, column number2

	csvFile, _ := os.Open(filepath)
	reader := csv.NewReader(bufio.NewReader(csvFile))

	var data csvData

	reader.Read()

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}

		firstData, err := strconv.ParseFloat(line[colNum1], 64)
		if err != nil {
			continue
		}
		secondData, err := strconv.ParseFloat(line[colNum2], 64)
		if err != nil {
			continue
		}

		data.FirstData = append(data.FirstData, firstData)
		data.SecondData = append(data.SecondData, secondData)
	}

	NormalizeData(data.FirstData)
	NormalizeData(data.SecondData)

	return data
}
func GetCSVNData(filepath string, colNum []int, normalize bool) [][]float64 {

	// path, array of column number, normalize data boolean

	csvFile, _ := os.Open(filepath)
	reader := csv.NewReader(bufio.NewReader(csvFile))

	data := make([][]float64, len(colNum))

	reader.Read()

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		for i, Num := range colNum {
			Ndata, err := strconv.ParseFloat(line[Num], 64)
			if err != nil {
				continue
			}
			data[i] = append(data[i], Ndata)
		}
	}
	if normalize {
		for i := range colNum {
			NormalizeData(data[i])
		}
	}

	return data
}
