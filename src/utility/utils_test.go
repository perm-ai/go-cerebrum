package utility

import (
	"math"
	"math/rand"
	"testing"
	"strconv"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/logger"
)

var utils = NewUtils(true)
var testCases = generateTestCases(utils)
var log = logger.NewLogger(true)

type TestCase struct {
	data1       ckks.Ciphertext
	data2       ckks.Ciphertext
	addExpected []float64
	subExpected []float64
	mulExpected []float64
	dotExpected []float64
}

func generateTestCases(u Utils) [4]TestCase {

	log.Log("Generating Test Cases")
	random1 := make([]float64, utils.Params.Slots())
	random2 := make([]float64, utils.Params.Slots())
	randomAdd := make([]float64, utils.Params.Slots())
	randomSub := make([]float64, utils.Params.Slots())
	randomMul := make([]float64, utils.Params.Slots())
	randomDot := make([]float64, utils.Params.Slots())

	for i := range random1 {
		random1[i] = rand.Float64() * 5000
		random2[i] = rand.Float64() * 5000
		randomAdd[i] = random1[i] + random2[i]
		randomSub[i] = random1[i] - random2[i]
		randomMul[i] = random1[i] * random2[i]
		randomDot[0] += randomMul[i]
	}

	// Normal ct (same scale, same level)
	t1 := TestCase{u.Encrypt(random1), u.Encrypt(random2), randomAdd, randomSub, randomMul, randomDot}

	// Ct with different scale, same level
	t2d1encd := u.EncodeToScale(random1, math.Pow(2, 40))
	t2d2encd := u.EncodeToScale(random1, math.Pow(2, 80))
	t2d1enct := u.Encryptor.EncryptFastNew(&t2d1encd)
	t2d2enct := u.Encryptor.EncryptFastNew(&t2d2encd)
	t2 := TestCase{*t2d1enct, *t2d2enct, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different level, same scale
	t3d1enct := u.Encrypt(random1)
	t3d2enct := u.Encrypt(random2)
	u.Evaluator.DropLevel(&t3d2enct, 3)
	t3 := TestCase{t3d1enct, t3d2enct, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different scale, different level
	t4d1encd := u.EncodeToScale(random1, math.Pow(2, 40))
	t4d2encd := u.EncodeToScale(random1, math.Pow(2, 80))
	t4d1enct := u.Encryptor.EncryptFastNew(&t4d1encd)
	t4d2enct := u.Encryptor.EncryptFastNew(&t4d2encd)
	u.Evaluator.DropLevel(t4d2enct, 3)
	t4 := TestCase{*t4d1enct, *t4d2enct, randomAdd, randomSub, randomMul, randomDot}

	return [4]TestCase{t1, t2, t3, t4}

}

func evalCorrectness(evalData []float64, testCase TestCase, operation string, decimalPrecision int) bool {

	precision := math.Pow(10, float64(-1*decimalPrecision))

	if operation != "dot" {
		var expected []float64

		switch operation {
		case "add":
			expected = testCase.addExpected
		case "sub":
			expected = testCase.subExpected
		case "mul":
			expected = testCase.mulExpected
		default:
			panic("Operation provided is invalid (only 'add', 'sub', 'mul', 'dot' are allowed)")
		}

		if len(expected) != len(evalData) {
			return false
		}

		for i := range evalData {
			if math.Abs(evalData[i]-expected[i]) < precision {
				return false
			}
		}

	} else if operation == "dot" {
		if math.Abs(evalData[0]-testCase.dotExpected[0]) < precision {
			return false
		}
	} else {
		panic("Operation provided is invalid (only 'add', 'sub', 'mul', 'dot' are allowed)")
	}

	return true

}

func TestEncodingDecoding(t *testing.T) {

	data := make([]float64, utils.Params.Slots())
	data[0] = 324.4
	data[122] = 75916.3

	encoded := utils.Encode(data)
	decoded := utils.Decode(&encoded)

	if !(math.Abs(decoded[0]-324.4) < 0.1 && math.Abs(decoded[122]-75916.3) < 0.1) {
		t.Error("Data wasn't correctly encoded")
	}

}

func TestEncryptionDecryption(t *testing.T) {

	data := make([]float64, utils.Params.Slots())
	data[0] = 324.4
	data[5] = 976.8
	data[122] = 75916.3

	ct := utils.Encrypt(data)
	dt := utils.Decrypt(&ct)

	if !(math.Abs(dt[0]-324.4) < 0.1 && math.Abs(dt[122]-75916.3) < 0.1 && math.Abs(dt[5]-976.8) < 0.1) {
		t.Error("Data wasn't correctly Encrypted and Decrypted")
	}

}

func TestAddition(t *testing.T) {

	for i := range testCases {

		log.Log("Testing addition (" + strconv.Itoa(i) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		addNew := utils.AddNew(&ct1, &ct2)
		addNewD := utils.Decrypt(&addNew)

		if !evalCorrectness(addNewD, testCases[i], "add", 1) {
			t.Error("Data wasn't correctly added (AddNew)")
		}

		utils.Add(ct1, ct2, &ct1)
		addD := utils.Decrypt(&ct1)

		if !evalCorrectness(addD, testCases[i], "add", 1) {
			t.Error("Data wasn't correctly added (Add)")
		}

	}

}

func TestSubtraction(t *testing.T) {

	for i := range testCases {

		log.Log("Testing subtraction (" + strconv.Itoa(i) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		subNew := utils.SubNew(&ct1, &ct2)
		subNewD := utils.Decrypt(&subNew)

		if !evalCorrectness(subNewD, testCases[i], "sub", 1) {
			t.Error("Data wasn't correctly added (SubNew)")
		}

		utils.Sub(ct1, ct2, &ct1)
		subD := utils.Decrypt(&ct1)

		if !evalCorrectness(subD, testCases[i], "sub", 1) {
			t.Error("Data wasn't correctly added (Sub)")
		}

	}

}

func TestEqualizeScale(t *testing.T) {

	encrypted1 := testCases[1].data1
	encrypted2 := testCases[1].data2

	utils.EqualizeScale(&encrypted1, &encrypted2)

	if encrypted1.Scale() != encrypted2.Scale() {
		t.Error("Ciphertext scale weren't properly equalized")
	}

}

func TestMultiplication(t *testing.T) {

	for i := range testCases {

		log.Log("Testing multiplication (" + strconv.Itoa(i) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		mulNew := utils.MultiplyNew(&ct1, &ct2)
		mulNewD := utils.Decrypt(&mulNew)

		if !evalCorrectness(mulNewD, testCases[i], "mul", 1) {
			t.Error("Data wasn't correctly multiplied (MultiplyNew)")
		}

		newCiphertext1 := ckks.NewCiphertext(&utils.Params, 1, utils.Params.MaxLevel(), math.Pow(2, 40))
		utils.Multiply(ct1, ct2, newCiphertext1)
		mulD := utils.Decrypt(newCiphertext1)

		if !evalCorrectness(mulD, testCases[i], "mul", 1) {
			t.Error("Data wasn't correctly multiplied (Multiply)")
		}

		mulNewRes := utils.MultiplyRescaleNew(&ct1, &ct2)
		mulNewResD := utils.Decrypt(&mulNewRes)

		if !evalCorrectness(mulNewResD, testCases[i], "mul", 1) && mulNewRes.Scale() != ct1.Scale()*ct2.Scale() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescaleNew)")
		}

		newCiphertext2 := ckks.NewCiphertext(&utils.Params, 1, utils.Params.MaxLevel(), math.Pow(2, 40))
		utils.MultiplyRescale(ct1, ct2, newCiphertext2)
		mulResD := utils.Decrypt(newCiphertext2)

		if !evalCorrectness(mulResD, testCases[i], "mul", 1) && newCiphertext2.Scale() != ct1.Scale()*ct2.Scale() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescale)")
		}


	}

}

func TestDotProduct(t *testing.T) {

	for i := range testCases {

		log.Log("Testing dot product (" + strconv.Itoa(i) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		dotNew := utils.DotProductNew(&ct1, &ct2)
		dotNewD := utils.Decrypt(&dotNew)

		if !evalCorrectness(dotNewD, testCases[i], "dot", 1) {
			t.Error("Dot product wasn't correctly calculated (DotProductNew)")
		}

		utils.DotProduct(ct1, ct2, &ct1)
		dotD := utils.Decrypt(&ct1)

		if !evalCorrectness(dotD, testCases[i], "dot", 2) {
			t.Error("Dot product wasn't correctly calculated (DotProduct)")
		}

	}

}