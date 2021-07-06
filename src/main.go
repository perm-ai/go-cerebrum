package main

import (

	//"math"
	// "fmt"

	//"math"

	"github.com/perm-ai/GO-HEML-prototype/src/importer"
	// "fmt"

	"github.com/perm-ai/GO-HEML-prototype/src/ml"
	// "github.com/perm-ai/GO-HEML-prototype/src/utility"
)

func main() {


	// lrData := importer.GetTitanicData("./test-data/titanic1.json")
	// x := lrData.Age
	// y := lrData.Pclass
	// target := lrData.Target
	// ml.Normalize_Data(x)
	// ml.Normalize_Data(y)
	// logisticRegression := ml.NewLogisticRegression()
	// ml.Train(logisticRegression, x, y, target, 0.1, 20)


	// Acc := 0.0
	// for i := 0; i < 10; i++ {
	// 	x, y, target := utility.GenerateLinearData(300)
	// 	logisticRegression := ml.NewLogisticRegression()
	// 	Acc += ml.Train(logisticRegression, x, y, target, 0.1, 20)
	// }
	// fmt.Printf("Average Accuracy : %f", Acc/10)

	utils := utility.NewUtils(math.Pow(2, 35), 0, true, true)
	testArray := utils.GenerateRandomNormalArray(5)

	for i := 0; i < 5; i++ {
		fmt.Println(testArray[i])
	}

	encryptedTestArray := utils.Encrypt(testArray)
	// testCipher := utils.MultiplyConstNew(encryptedTestArray, utils.GenerateFilledArraySize(2, 10),true ,false)

	// testCipher := ml.Sigmoid(encryptedTestArray)

	testcipher := utils.AddNew(encryptedTestArray, encryptedTestArray)
	decryptedTestArray := utils.Decrypt(&testCipher)

	for i := 0; i < 5; i++ {
		fmt.Println(decryptedTestArray[i])

}
