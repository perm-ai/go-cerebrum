package utility

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
)

type Utils struct {
	hasSecretKey     bool
	bootstrapEnabled bool

	BootstrappingParams bootstrapping.Parameters
	Params              ckks.Parameters
	KeyChain            key.KeyChain

	Bootstrapper *bootstrapping.Bootstrapper
	Encoder      ckks.Encoder
	Evaluator    ckks.Evaluator
	Encryptor    rlwe.Encryptor
	Decryptor    rlwe.Decryptor

	Filters []rlwe.Plaintext
	Scale   float64
	log     logger.Logger
}

func NewUtils(keyChain key.KeyChain, scale float64, filtersAmount int, logEnabled bool) Utils {


	if keyChain.RelinKey == nil || keyChain.GaloisKey == nil {
		panic("Missing keys must have both relinearlize keys and galois keys in keychain to generate new utils")
	}

	bootstrapEnabled := keyChain.BtspGalKey != nil
	log := logger.NewLogger(logEnabled)

	paramSet := bootstrapping.DefaultParametersSparse[keyChain.ParamsIndex]
	ckksParams := paramSet.SchemeParams

	bootstrappingParams := paramSet.BootstrappingParams
	params, _ := ckks.NewParametersFromLiteral(ckksParams)

	log.Log("Util Initialization: Generating encoder, evaluator, encryptor, decryptor")
	encoder := ckks.NewEncoder(params)
	evaluator := ckks.NewEvaluator(params, rlwe.EvaluationKey{Rlk: keyChain.RelinKey})
	encryptor := ckks.NewEncryptor(params, keyChain.PublicKey)

	var decryptor rlwe.Decryptor
	decryptor = nil

	if keyChain.SecretKey != nil {
		decryptor = ckks.NewDecryptor(params, keyChain.SecretKey)
	}

	filters := make([]rlwe.Plaintext, filtersAmount)

	for i := range filters {
		filter := make([]complex128, filtersAmount)
		filter[i] = complex(1, 0)
		filters[i] = *encoder.EncodeNew(filter, params.MaxLevel(), params.DefaultScale(), params.LogSlots())
	}

	var bootstrapper *bootstrapping.Bootstrapper
	bootstrapper = nil

	if bootstrapEnabled {

		swkDtS, swkStD := bootstrappingParams.GenEncapsulationSwitchingKeys(params, keyChain.SecretKey)
		evalKeys := rlwe.EvaluationKey{Rlk: keyChain.RelinKey, Rtks: keyChain.BtspGalKey};
		bootstrappingKey := bootstrapping.EvaluationKeys{evalKeys, swkDtS, swkStD}

		var err error
		log.Log("Util Initialization: Generating bootstrapper")
		bootstrapper, err = bootstrapping.NewBootstrapper(params, bootstrappingParams, bootstrappingKey)

		if err != nil {
			panic("BOOTSTRAPPER GENERATION ERROR")
		}
	}

	return Utils{
		true,
		bootstrapEnabled,
		bootstrappingParams,
		params,
		keyChain,
		bootstrapper,
		encoder,
		evaluator,
		encryptor,
		decryptor,
		filters,
		scale,
		log,
	}

}

func NewDecryptionUtils(keyChain key.KeyChain, scale float64, logEnabled bool) Utils {
	log := logger.NewLogger(logEnabled)

	bootstrappingParams := bootstrapping.DefaultParameters[keyChain.ParamsIndex]
	params, _ := ckks.NewParametersFromLiteral(bootstrapping.DefaultCKKSParameters[keyChain.ParamsIndex])
	encoder := ckks.NewEncoder(params)
	encryptor := ckks.NewFastEncryptor(params, keyChain.PublicKey)
	decryptor := ckks.NewDecryptor(params, keyChain.SecretKey)

	return Utils{
		true,
		false,
		bootstrappingParams,
		params,
		keyChain,
		nil,
		encoder,
		nil,
		encryptor,
		decryptor,
		nil,
		scale,
		log,
	}

}

func NewEncryptionUtils(keyChain key.KeyChain, scale float64, logEnabled bool) Utils{

	log := logger.NewLogger(logEnabled)

	bootstrappingParams := bootstrapping.DefaultParameters[keyChain.ParamsIndex]
	params, _ := ckks.NewParametersFromLiteral(bootstrapping.DefaultCKKSParameters[keyChain.ParamsIndex])
	encoder := ckks.NewEncoder(params)

	var encryptor ckks.Encryptor
	if keyChain.PublicKey != nil {
		encryptor = ckks.NewFastEncryptor(params, keyChain.PublicKey)
	} else if keyChain.SecretKey != nil{
		encryptor = ckks.NewEncryptor(params, keyChain.SecretKey)
	} else {
		panic("Secret or Public key must be provided")
	}
	

	return Utils{
		true,
		false,
		bootstrappingParams,
		params,
		keyChain,
		nil,
		encoder,
		nil,
		encryptor,
		nil,
		nil,
		scale,
		log,
	}

}

func (u Utils) GenerateRandomFloatArray(length int, lowerBound float64, upperBound float64) []float64 {

	randomArr := make([]float64, u.Params.Slots())
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < length; i++ {

		randomArr[i] = (rand.Float64() * (upperBound - lowerBound)) + lowerBound

	}

	return randomArr

}

func (u Utils) GenerateRandomArray(lowerBound float64, upperBound float64, length int) []float64 {

	if lowerBound >= upperBound {
		panic("Lower bound must be higher than upper bound")
	}

	randomArr := make([]float64, u.Params.Slots())

	for i := 0; i < length; i++ {
		randomArr[i] = (rand.Float64() * (upperBound - lowerBound)) + lowerBound
	}

	return randomArr

}

// Encode into complex value
func (u Utils) Encode(value []float64) ckks.Plaintext {

	// Encode value
	// plaintext := ckks.NewPlaintext(&u.Params, u.Params.MaxLevel(), u.Params.Scale())
	plaintext := u.Encoder.EncodeNew(u.Float64ToComplex128(value), u.Params.LogSlots())

	return *plaintext

}

// Encode into complex value with non-default scale
func (u Utils) EncodeToScale(value []float64, scale float64) ckks.Plaintext {

	// Encode value
	plaintext := ckks.NewPlaintext(u.Params, u.Params.MaxLevel(), scale)
	u.Encoder.Encode(plaintext, u.Float64ToComplex128(value), u.Params.LogSlots())

	return *plaintext

}

// Encode into float coefficient
func (u Utils) EncodeCoeffs(value []float64) ckks.Plaintext {

	// Encode value
	plaintext := ckks.NewPlaintext(u.Params, u.Params.MaxLevel(), u.Params.Scale())
	u.Encoder.EncodeCoeffs(value, plaintext)

	return *plaintext

}

// Encode float array into NTT Plaintext
func (u Utils) EncodePlaintextFromArray(arr []float64) *ckks.Plaintext {
	return u.Encoder.EncodeNTTNew(u.Float64ToComplex128(arr), u.Params.LogSlots())
}

// Encode float array into NTT Plaintext
func (u Utils) EncodePlaintextFromArrayScale(arr []float64, scale float64) *ckks.Plaintext {
	pt := ckks.NewPlaintext(u.Params, u.Params.MaxLevel(), scale)
	u.Encoder.EncodeNTT(pt, u.Float64ToComplex128(arr), u.Params.LogSlots())
	return pt
}

// Decode complex plaintext and take real part returning float array
func (u Utils) Decode(value *ckks.Plaintext) []float64 {

	return u.Complex128ToFloat64(u.Encoder.Decode(value, u.Params.LogSlots()))

}

// Encode into float coefficient with non default scale
func (u Utils) EncodeCoeffsToScale(value []float64, scale float64) ckks.Plaintext {

	// Encode value
	plaintext := ckks.NewPlaintext(u.Params, u.Params.MaxLevel(), scale)
	u.Encoder.EncodeCoeffs(value, plaintext)

	return *plaintext

}

func (u Utils) Encrypt(value []float64) ckks.Ciphertext {

	// Encode value
	plaintext := u.EncodeToScale(value, u.Scale)

	// Encrypt value
	ciphertext := u.Encryptor.EncryptNew(&plaintext)

	return *ciphertext

}

func (u Utils) EncryptToPointer(value []float64) *ckks.Ciphertext {

	// Encode value
	plaintext := u.EncodeToScale(value, u.Scale)

	// Encrypt value
	ciphertext := u.Encryptor.EncryptNew(&plaintext)

	return ciphertext

}

func (u Utils) EncryptToScale(value []float64, scale float64) ckks.Ciphertext {

	// Encode value
	plaintext := u.EncodeToScale(value, scale)

	// Encrypt value
	ciphertext := u.Encryptor.EncryptNew(&plaintext)

	return *ciphertext

}

func (u Utils) EncryptToLevel(value []float64, level int) *ckks.Ciphertext {

	// Encode value
	plaintext := u.EncodeToScale(value, u.Scale)

	// Encrypt value
	ciphertext := ckks.NewCiphertext(u.Params, 1, level, u.Scale)
	u.Encryptor.Encrypt(&plaintext, ciphertext)

	return ciphertext

}

func (u Utils) EncryptToLevelScale(value []float64, level int, scale float64) *ckks.Ciphertext {

	// Encode value
	plaintext := u.EncodeToScale(value, scale)

	// Encrypt value
	ciphertext := ckks.NewCiphertext(u.Params, 1, level, scale)
	u.Encryptor.Encrypt(&plaintext, ciphertext)

	return ciphertext

}

func (u Utils) Decrypt(ciphertext *ckks.Ciphertext) []float64 {

	if u.Decryptor == nil {
		panic("Unable to decrypt due to lack of decryptor")
	}

	decrypted := u.Decryptor.DecryptNew(ciphertext)

	decoded := u.Decode(decrypted)

	return decoded

}

func ValidateResult(evalData []float64, expected []float64, isDot bool, decimalPrecision float64, log logger.Logger) bool {

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