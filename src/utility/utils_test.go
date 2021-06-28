package utility

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/logger"
)

var utils = NewUtils(false, true)
var log = logger.NewLogger(true)

type TestCase struct {
	data1       ckks.Ciphertext
	data2       ckks.Ciphertext
	rawData1    []float64
	rawData2    []float64
	addExpected []float64
	subExpected []float64
	mulExpected []float64
	dotExpected []float64
}

func GenerateTestCases(u Utils) [4]TestCase {

	rand.Seed(time.Now().UnixNano())
	log.Log("Generating Test Cases")
	random1 := make([]float64, utils.Params.Slots())
	random2 := make([]float64, utils.Params.Slots())
	randomAdd := make([]float64, utils.Params.Slots())
	randomSub := make([]float64, utils.Params.Slots())
	randomMul := make([]float64, utils.Params.Slots())
	randomDot := make([]float64, utils.Params.Slots())

	for i := 0; i < int(utils.Params.Slots()); i++ {
		random1[i] = rand.Float64() * 100
		random2[i] = rand.Float64() * 100
		randomAdd[i] = random1[i] + random2[i]
		randomSub[i] = random1[i] - random2[i]
		randomMul[i] = random1[i] * random2[i]
		randomDot[0] += randomMul[i]
	}

	fmt.Println(random1[0], random2[0])

	// Normal ct (same scale, same level)
	t1 := TestCase{u.Encrypt(random1), u.Encrypt(random2), random1, random2, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different scale, same level
	t2d1encd := u.EncodeToScale(random1, math.Pow(2, 30))
	t2d2encd := u.EncodeToScale(random2, math.Pow(2, 60))
	t2d1enct := u.Encryptor.EncryptFastNew(&t2d1encd)
	t2d2enct := u.Encryptor.EncryptFastNew(&t2d2encd)
	t2 := TestCase{*t2d1enct, *t2d2enct, random1, random2, randomAdd, randomSub, randomMul, randomDot}
	fmt.Printf("t2d2encd: %f, t2d2enct: %f\n", u.Decode(&t2d2encd)[0], u.Decrypt(t2d2enct)[0])

	// Ct with different level, same scale
	t3d1enct := u.Encrypt(random1)
	t3d2enct := u.Encrypt(random2)
	u.Evaluator.DropLevel(&t3d2enct, 3)
	t3 := TestCase{t3d1enct, t3d2enct, random1, random2, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different scale, different level
	t4d1encd := u.EncodeToScale(random1, math.Pow(2, 30))
	t4d2encd := u.EncodeToScale(random2, math.Pow(2, 60))
	t4d1enct := u.Encryptor.EncryptFastNew(&t4d1encd)
	t4d2enct := u.Encryptor.EncryptFastNew(&t4d2encd)
	u.Evaluator.DropLevel(t4d2enct, 3)
	t4 := TestCase{*t4d1enct, *t4d2enct, random1, random2, randomAdd, randomSub, randomMul, randomDot}

	return [4]TestCase{t1, t2, t3, t4}

}

func EvalCorrectness(evalData []float64, expected []float64, isDot bool, decimalPrecision int) bool {

	precision := math.Pow(10, float64(-1*decimalPrecision))

	if !isDot {

		if len(expected) != len(evalData) {
			log.Log("Data has inequal length")
			return false
		}

		for i := range evalData {
			if math.Abs(evalData[i]-expected[i]) > precision {
				log.Log("Incorrect evaluation (Expected: " + fmt.Sprintf("%f", expected[i]) + " Got: " + fmt.Sprintf("%f", evalData[i]) + " Index: " + strconv.Itoa(i) + ")")
				return false
			}
		}

	} else {
		if math.Abs(evalData[0]-expected[0]) > precision {
			return false
		}
	}

	return true

}

func TestComplexToFloat(t *testing.T) {

	data := make([]complex128, utils.Params.Slots())
	data[0] = complex(324.4, 0)
	data[122] = complex(75916.3, 0)
	data[300] = complex(2334578347, 0)

	float := utils.Complex128ToFloat64(data)

	if !(float[0] == 324.4 && float[122] == 75916.3 && float[300] == 2334578347) {
		t.Error("Complex array wasn't correctly converted to float")
	}

}

func TestFloatToComplex(t *testing.T) {

	data := make([]float64, utils.Params.Slots())
	data[0] = 324.4
	data[122] = 75916.3
	data[300] = 2334578347

	float := utils.Float64ToComplex128(data)

	if !(float[0] == complex(324.4, 0) && float[122] == complex(75916.3, 0) && float[300] == complex(2334578347, 0)) {
		t.Error("Complex array wasn't correctly converted to float")
	}

}

func TestEncodingDecoding(t *testing.T) {

	data := utils.GenerateFilledArray(0.0)
	data[0] = 324.4
	data[122] = 75916.3
	data[300] = 2334556.3

	encoded := utils.Encode(data)
	decoded := utils.Decode(&encoded)

	fmt.Println(decoded[0])

	if !EvalCorrectness(decoded, data, false, 1) {
		t.Error("Data wasn't correctly encoded")
	}

}

func TestEncodingToScale(t *testing.T) {

	data := utils.GenerateFilledArray(0.0)
	data[0] = 324.4
	data[122] = 75916.3
	data[300] = 2334556.3

	encoded := utils.EncodeToScale(data, math.Pow(2.0, 20.0))
	decoded := utils.Decode(&encoded)

	if !EvalCorrectness(decoded, data, false, 1) {
		t.Error("Data wasn't correctly encoded to scale (2^80)")
	}

	data = utils.GenerateFilledArray(0.0)
	data[0] = 20
	data[122] = 30
	data[300] = 50

	encoded = utils.EncodeToScale(data, math.Pow(2.0, 60))
	decoded = utils.Decode(&encoded)

	if !EvalCorrectness(decoded, data, false, 1) {
		t.Error("Data wasn't correctly encoded to scale (2^60)")
	}

	data = utils.GenerateFilledArray(0.0)
	data[0] = 20
	data[122] = 30
	data[300] = 50

	encoded = utils.EncodeToScale(data, math.Pow(2.0, 80))
	decoded = utils.Decode(&encoded)

	if !EvalCorrectness(decoded, data, false, 1) {
		t.Error("Data wasn't correctly encoded to scale (2^80)")
	}

}

func TestEncryptionDecryption(t *testing.T) {

	data := utils.GenerateFilledArray(0.0)
	data[0] = 324.4
	data[5] = 2334556.3
	data[122] = 75916.3

	ct := utils.Encrypt(data)
	dt := utils.Decrypt(&ct)

	if !(math.Abs(dt[0]-324.4) < 0.1 && math.Abs(dt[122]-75916.3) < 0.1 && math.Abs(dt[5]-2334556.3) < 0.1) {
		t.Error("Data wasn't correctly Encrypted and Decrypted")
	}

}

func TestAddition(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing addition (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		fmt.Println(utils.Decrypt(&testCases[i].data1)[0], utils.Decrypt(&testCases[i].data2)[0], testCases[i].data1.Scale(), testCases[i].data2.Scale())
		fmt.Println(testCases[i].rawData1[0], testCases[i].rawData2[0])

		sum := utils.AddNew(ct1, ct2)
		addNewD := utils.Decrypt(&sum)

		if !EvalCorrectness(addNewD, testCases[i].addExpected, false, 1) {
			t.Error("Data wasn't correctly added (AddNew)")
		}

		utils.Add(ct1, ct2, &sum)
		addD := utils.Decrypt(&sum)

		if !EvalCorrectness(addD, testCases[i].addExpected, false, 1) {
			t.Error("Data wasn't correctly added (Add)")
		}

	}

}

func TestSubtraction(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing subtraction (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		subNew := utils.SubNew(ct1, ct2)
		subNewD := utils.Decrypt(&subNew)

		if !EvalCorrectness(subNewD, testCases[i].subExpected, false, 1) {
			t.Error("Data wasn't correctly subtracted (SubNew)")
		}

		utils.Sub(ct1, ct2, &ct1)
		subD := utils.Decrypt(&ct1)

		if !EvalCorrectness(subD, testCases[i].subExpected, false, 1) {
			t.Error("Data wasn't correctly subtracted (Sub)")
		}

	}

}

func TestEqualizeScale(t *testing.T) {

	testCases := GenerateTestCases(utils)

	encrypted1 := testCases[1].data1
	encrypted2 := testCases[1].data2

	utils.EqualizeScale(&encrypted1, &encrypted2)

	if encrypted1.Scale() != encrypted2.Scale() {
		t.Error("Ciphertext scale weren't properly equalized")
	}

	decrypted1 := utils.Decrypt(&encrypted1)
	decrypted2 := utils.Decrypt(&encrypted2)

	if !(EvalCorrectness(decrypted1, testCases[1].rawData1, false, 1) && !EvalCorrectness(decrypted2, testCases[1].rawData2, false, 1)) {
		t.Error("Ciphertext scale weren't properly equalized")
	}

}

func TestMultiplication(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing multiplication (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		mulNew := utils.MultiplyNew(ct1, ct2)
		mulNewD := utils.Decrypt(&mulNew)

		if !EvalCorrectness(mulNewD, testCases[i].mulExpected, false, 1) {
			t.Error("Data wasn't correctly multiplied (MultiplyNew)")
		}

		newCiphertext1 := ckks.NewCiphertext(&utils.Params, 1, utils.Params.MaxLevel(), math.Pow(2, 40))
		utils.Multiply(ct1, ct2, newCiphertext1)
		mulD := utils.Decrypt(newCiphertext1)

		if !EvalCorrectness(mulD, testCases[i].mulExpected, false, 1) {
			t.Error("Data wasn't correctly multiplied (Multiply)")
		}

		mulNewRes := utils.MultiplyRescaleNew(&ct1, &ct2)
		mulNewResD := utils.Decrypt(&mulNewRes)

		if !EvalCorrectness(mulNewResD, testCases[i].mulExpected, false, 1) && mulNewRes.Scale() != ct1.Scale()*ct2.Scale() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescaleNew)")
		}

		newCiphertext2 := ckks.NewCiphertext(&utils.Params, 1, utils.Params.MaxLevel(), math.Pow(2, 40))
		utils.MultiplyRescale(ct1, ct2, newCiphertext2)
		mulResD := utils.Decrypt(newCiphertext2)

		if !EvalCorrectness(mulResD, testCases[i].mulExpected, false, 1) && newCiphertext2.Scale() != ct1.Scale()*ct2.Scale() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescale)")
		}

	}

}

func TestDotProduct(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing dot product (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		dotNew := utils.DotProductNew(&ct1, &ct2)
		dotNewD := utils.Decrypt(&dotNew)

		if !EvalCorrectness(dotNewD, testCases[i].dotExpected, true, 1) {
			t.Error("Dot product wasn't correctly calculated (DotProductNew)")
		}

		utils.DotProduct(ct1, ct2, &ct1)
		dotD := utils.Decrypt(&ct1)

		if !EvalCorrectness(dotD, testCases[i].dotExpected, true, 2) {
			t.Error("Dot product wasn't correctly calculated (DotProduct)")
		}

	}

}