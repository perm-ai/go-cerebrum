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

var utils = NewUtils(math.Pow(2,35), 100, false, true)
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

	// Normal ct (same scale, same level)
	t1 := TestCase{u.Encrypt(random1), u.Encrypt(random2), random1, random2, randomAdd, randomSub, randomMul, randomDot}

	// Ct with different scale, same level
	t2d1encd := u.EncodeToScale(random1, math.Pow(2, 30))
	t2d2encd := u.EncodeToScale(random2, math.Pow(2, 60))
	t2d1enct := u.Encryptor.EncryptFastNew(&t2d1encd)
	t2d2enct := u.Encryptor.EncryptFastNew(&t2d2encd)
	t2 := TestCase{*t2d1enct, *t2d2enct, random1, random2, randomAdd, randomSub, randomMul, randomDot}

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

func EvalCorrectness(evalData []float64, expected []float64, isDot bool, decimalPrecision float64) bool {

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

		for i := range evalData {
			if math.Abs(evalData[i]-expected[0]) > precision {
				log.Log("Incorrect evaluation (Expected: " + fmt.Sprintf("%f", expected[i]) + " Got: " + fmt.Sprintf("%f", evalData[i]) + " Index: " + strconv.Itoa(i) + ")")
				return false
			}
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

func TestMultiplication(t *testing.T) {

	testCases := GenerateTestCases(utils)

	for i := range testCases {

		log.Log("Testing multiplication (" + strconv.Itoa(i+1) + "/4)")
		ct1 := testCases[i].data1
		ct2 := testCases[i].data2

		mulNew := utils.MultiplyNew(ct1, ct2, false, true)
		mulNewD := utils.Decrypt(&mulNew)

		if !EvalCorrectness(mulNewD, testCases[i].mulExpected, false, 1) {
			t.Error("Data wasn't correctly multiplied (MultiplyNew)")
		}

		newCiphertext1 := ckks.NewCiphertext(&utils.Params, 1, utils.Params.MaxLevel(), math.Pow(2, 40))
		utils.Multiply(ct1, ct2, newCiphertext1, false, true)
		mulD := utils.Decrypt(newCiphertext1)

		if !EvalCorrectness(mulD, testCases[i].mulExpected, false, 1) {
			t.Error("Data wasn't correctly multiplied (Multiply)")
		}

		mulNewRes := utils.MultiplyNew(ct1, ct2, true, true)
		mulNewResD := utils.Decrypt(&mulNewRes)

		if !EvalCorrectness(mulNewResD, testCases[i].mulExpected, false, 1) && mulNewRes.Scale() != ct1.Scale()*ct2.Scale() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescaleNew)")
		}

		newCiphertext2 := ckks.NewCiphertext(&utils.Params, 1, utils.Params.MaxLevel(), math.Pow(2, 40))
		utils.Multiply(ct1, ct2, newCiphertext2,  true, true)
		mulResD := utils.Decrypt(newCiphertext2)

		if !EvalCorrectness(mulResD, testCases[i].mulExpected, false, 1) && newCiphertext2.Scale() != ct1.Scale()*ct2.Scale() {
			t.Error("Data wasn't correctly multiplied (MultiplyRescale)")
		}

	}

}

func TestDotProduct(t *testing.T) {

	testCases := GenerateTestCases(utils)

	ct1 := testCases[0].data1
	ct2 := testCases[0].data2

	dotNew := utils.DotProductNew(ct1, ct2, true)
	dotNewD := utils.Decrypt(&dotNew)

	if !EvalCorrectness(dotNewD, testCases[0].dotExpected, true, -0.69) {
		t.Error("Dot product wasn't correctly calculated (DotProductNew)")
	}

	utils.DotProduct(ct1, ct2, &ct1, true)
	dotD := utils.Decrypt(&ct1)

	if !EvalCorrectness(dotD, testCases[0].dotExpected, true, -0.69) {
		t.Error("Dot product wasn't correctly calculated (DotProduct)")
	}

}

// func TestBootstrapping(t *testing.T) {

// 	pt := ckks.NewPlaintext(&utils.Params, 1, math.Pow(2, 40))
// 	utils.Encoder.Encode(pt, utils.Float64ToComplex128(utils.GenerateFilledArray(3.12)), utils.Params.LogSlots())
// 	ct := utils.Encryptor.EncryptFastNew(pt)
// 	preBootstrap := ct.Level()

// 	utils.BootstrapIfNecessary(ct)

// 	decrypted := utils.Decrypt(ct)

// 	// Test if bootstrap increase level and correctly decrypt
// 	if(ct.Level() <= preBootstrap || !EvalCorrectness(decrypted, utils.GenerateFilledArray(3.12), false, 1)){
// 		t.Error("Wasn't bootstrapped correctly")
// 	}

// 	encTwos := utils.Encrypt(utils.GenerateFilledArray(2))
// 	utils.Multiply(encTwos, *ct, ct, true, true)

// 	decrypted = utils.Decrypt(ct)

// 	if(!EvalCorrectness(decrypted, utils.GenerateFilledArray(3.12 * 2), false, 1)){
// 		t.Error("Wasn't evaluated correctly after bootstrap")
// 	}

// }

func TestTranspose(t *testing.T){

	// Test case:										Expected result:
	// [[01, 02, 03, 04, 05, 06, 07, 08, 09, 10],		[[01, 11, 21],
	//  [11, 12, 13, 14, 15, 16, 17, 18, 19, 20],		 [02, 12, 22], ...
	//  [21, 22, 23, 24, 25, 26, 27, 28, 29, 30]]		 [10, 20, 30]]

	testCase := make([]ckks.Ciphertext, 3)
	
	// Generate test case
	for i := range testCase{

		data := make([]float64, utils.Params.Slots());
		for j := 1; j <= 10; j++ {
			data[j-1] = float64((10 * i) + j); 
		}

		testCase[i] = utils.Encrypt(data);

	}

	// Generate expected array to check correctness of the function
	expected := make([][]float64, 10)

	for i := range expected{
		row := make([]float64, utils.Params.Slots())
		for j := 0; j < 3; j++ {
			row[j] = float64(i + 1 + (10 * j))
		}
		expected[i] = row
	}

	// Compute transposed array
	transposedCt := utils.Transpose(testCase, 10)

	for i := range transposedCt{
		decryptedResult := utils.Decrypt(&transposedCt[i])

		if !EvalCorrectness(decryptedResult, expected[i], false, 1){

			t.Error("Data was incorrectly transposed")

		}
	}

}

func TestOuterProduct(t *testing.T){

	// Test the correctness of outer product evaluation betweem two ciphertexts
	// Test case:			Expected:
	// A = E(3, 4)			[ E(6, 9, 15, 18),
	// B = E(2, 3, 5, 6)	  E(8, 12, 20, 24)]

	testCaseA := utils.Encrypt([]float64{3, 4})
	testCaseB := utils.Encrypt([]float64{2, 3, 5, 6})

	outerProduct := utils.Outer(&testCaseA, &testCaseB, 2, 4)
	
	for i := range outerProduct {

		decryptedProduct := utils.Decrypt(&outerProduct[i])
		expectedResult := make([]float64, utils.Params.Slots())
		
		if i == 0 {
			expectedResult[0] = 6
			expectedResult[1] = 9
			expectedResult[2] = 15
			expectedResult[3] = 18
		} else {
			expectedResult[0] = 8
			expectedResult[1] = 12
			expectedResult[2] = 20
			expectedResult[3] = 24
		}

		if !EvalCorrectness(decryptedProduct, expectedResult, false, 1){

			t.Error("Outer was incorrectly calculated")

		}

	}

}