package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/regression"
	"github.com/perm-ai/go-cerebrum/utility"
)

type EncryptedData struct {
	Name      string   `json:"name"`
	Encrypted []Column `json:"encryptedData"`
}

type Column struct {
	ColumnName string  `json:"columnName"`
	Type       string  `json:"type"`
	Length     int     `json:"length"`
	Data       string  `json:"data"`
	Label      []Label `json:"label"`
}

type Label struct {
	Category string `json:"category"`
	Index    int    `json:"index"`
	Data     string `json:"data"`
}

func main() {

	// key.LoadKeys("/Users/phu/Desktop/Perm/Banpu Coal Data", 0, true, true, true, true)

	jsonFile, err := os.Open("/Users/phu/Desktop/Perm/Banpu Coal Data/Coal_Train_encrypted.json")

	if err != nil {
		fmt.Println(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var trainingData EncryptedData

	json.Unmarshal(byteValue, &trainingData)

	dataBytes := make([][]byte, len(trainingData.Encrypted))

	for i := 0; i < len(trainingData.Encrypted); i++ {
		dataBytes[i] = []byte{}
		dataBytes[i], _ = base64.StdEncoding.DecodeString(trainingData.Encrypted[i].Data)
	}

	ct := make([]ckks.Ciphertext, len(dataBytes))
	for i := 0; i < len(dataBytes); i++ {
		ct[i] = ckks.Ciphertext{}
		ct[i].UnmarshalBinary(dataBytes[i])
	}

	x := make([]*ckks.Ciphertext, 2)
	// x[0] = ckks.Ciphertext{}
	x[0] = ct[1].CopyNew()
	// x[1] = ckks.Ciphertext{}
	x[1] = ct[2].CopyNew()

	y := ckks.Ciphertext{}

	y  =*ct[0].CopyNew()

	keyPair := key.LoadKeys("/Users/phu/Desktop/Perm/Banpu Coal Data", 0, true, true, false, false)
	keychain := key.GenerateKeysFromKeyPair(0, keyPair.SecretKey, keyPair.PublicKey, true, true)

	utils := utility.NewUtils(keychain, math.Pow(2, 40), 100, true)

	model := regression.NewLinearRegression(utils, 2)

	model.Train(x, &y, 0.01, trainingData.Encrypted[0].Length, 20)

	m := make([][]float64, len(x))
	
	for i := 0; i < len(x); i++ {
		m[i] = utils.Decrypt(x[i])
	}
	// b := []float64
	b := utils.Decrypt(&y)
	fmt.Printf("The weights is %f and %f", m[0][0], m[1][0])
	fmt.Printf("The bias is %f", b[0])
}
