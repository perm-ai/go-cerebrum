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
	age []float64
	sex []float64
}

func getCSV(filepath string) csvData {
	csvFile, _ := os.Open(filepath)
	reader := csv.NewReader(bufio.NewReader(csvFile))

	var data csvData

	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		age, _ := strconv.ParseFloat(line[0], 64)
		data.age = append(data.age, age)
		sex, _ := strconv.ParseFloat(line[1], 64)
		data.sex = append(data.sex, sex)
	}

	return data
}
