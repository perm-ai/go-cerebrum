package utility

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/rlwe"
	"github.com/perm-ai/GO-HEML-prototype/src/logger"
)

type Utils struct {
	hasSecretKey     bool
	bootstrapEnabled bool

	BootstrappingParams ckks.BootstrappingParameters
	Params              ckks.Parameters
	secretKey           rlwe.SecretKey
	PublicKey           rlwe.PublicKey
	RelinKey            rlwe.RelinearizationKey
	GaloisKey           rlwe.RotationKeySet
	BtspGaloisKey       rlwe.RotationKeySet

	Bootstrapper *ckks.Bootstrapper
	Encoder      ckks.Encoder
	Evaluator    ckks.Evaluator
	Encryptor    ckks.Encryptor
	Decryptor    ckks.Decryptor

	Filters []ckks.Plaintext
	Scale   float64
	log     logger.Logger
}

func NewUtils(scale float64, filtersAmount int, bootstrapEnabled bool, logEnabled bool) Utils {

	log := logger.NewLogger(logEnabled)

	bootstrappingParams := ckks.DefaultBootstrapParams[0]
	Params, _ := bootstrappingParams.Params()

	log.Log("Util Initialization: Generating key generator")
	keyGenerator := ckks.NewKeyGenerator(Params)

	log.Log("Util Initialization: Generating private / public key pair")
	secretKey, publicKey := keyGenerator.GenKeyPairSparse(bootstrappingParams.H)

	log.Log("Util Initialization: Generating relin key")
	relinKey := keyGenerator.GenRelinearizationKey(secretKey)

	log.Log("Util Initialization: Generating galois keys")
	galoisKey := keyGenerator.GenRotationKeysForRotations(getSumElementsKs(Params.LogSlots()), true, secretKey)

	log.Log("Util Initialization: Generating encoder, evaluator, encryptor, decryptor")
	Encoder := ckks.NewEncoder(Params)
	Evaluator := ckks.NewEvaluator(Params, rlwe.EvaluationKey{Rlk: relinKey})
	Encryptor := ckks.NewEncryptorFromPk(Params, publicKey)
	Decryptor := ckks.NewDecryptor(Params, secretKey)

	filters := make([]ckks.Plaintext, filtersAmount)

	for i := range filters {
		filter := make([]complex128, filtersAmount)
		filter[i] = complex(1, 0)
		filters[i] = *Encoder.EncodeNTTAtLvlNew(Params.MaxLevel(), filter, Params.LogSlots())
	}

	if bootstrapEnabled {
		log.Log("Util Initialization: Generating bootstrapping key")

		rotations := bootstrappingParams.RotationsForBootstrapping(Params.LogSlots())
		rotationKeys := keyGenerator.GenRotationKeysForRotations(rotations, true, secretKey)

		bootstrappingKey := ckks.BootstrappingKey{Rlk: relinKey, Rtks: rotationKeys}

		var err error
		var bootstrapper *ckks.Bootstrapper

		log.Log("Util Initialization: Generating bootstrapper")
		bootstrapper, err = ckks.NewBootstrapper(Params, bootstrappingParams, bootstrappingKey)

		if err != nil {
			panic("BOOTSTRAPPER GENERATION ERROR")
		}

		return Utils{
			true,
			bootstrapEnabled,
			*bootstrappingParams,
			Params,
			*secretKey,
			*publicKey,
			*relinKey,
			*galoisKey,
			*rotationKeys,
			bootstrapper,
			Encoder,
			Evaluator,
			Encryptor,
			Decryptor,
			filters,
			scale,
			log,
		}
	} else {
		return Utils{
			true,
			bootstrapEnabled,
			*bootstrappingParams,
			Params,
			*secretKey,
			*publicKey,
			*relinKey,
			*galoisKey,
			rlwe.RotationKeySet{},
			&ckks.Bootstrapper{},
			Encoder,
			Evaluator,
			Encryptor,
			Decryptor,
			filters,
			scale,
			log,
		}
	}

}

func NewUtilsFromKeyChain(keyChain KeyChain, scale float64, filtersAmount int, logEnabled bool) Utils {

	log := logger.NewLogger(logEnabled)

	bootstrappingParams := ckks.DefaultBootstrapParams[0]
	Params, _ := bootstrappingParams.Params()

	log.Log("Util Initialization: Loading private / public key pair")
	secretKey := rlwe.NewSecretKey(Params.Parameters)

	if len(keyChain.secretKey) != 0 {
		secretKey.UnmarshalBinary(keyChain.secretKey)
	}

	publicKey := rlwe.NewPublicKey(Params.Parameters)
	publicKey.UnmarshalBinary(keyChain.PublicKey)

	log.Log("Util Initialization: Loading relin key")
	relinKey := ckks.NewRelinearizationKey(Params)
	relinKey.UnmarshalBinary(keyChain.RelinKey)

	log.Log("Util Initialization: Loading galois keys")

	ks := getSumElementsKs(Params.LogSlots())
	galEl := make([]uint64, len(ks)+1)

	for i := range galEl {
		if i == 0 {
			galEl[i] = Params.GaloisElementForRowRotation()
		} else {
			galEl[i] = Params.GaloisElementForColumnRotationBy(i)
		}
	}

	galoisKey := ckks.NewRotationKeySet(Params, galEl)
	galoisKey.UnmarshalBinary(keyChain.GaloisKey)

	log.Log("Util Initialization: Generating encoder, evaluator, encryptor, decryptor")
	Encoder := ckks.NewEncoder(Params)
	Evaluator := ckks.NewEvaluator(Params, rlwe.EvaluationKey{Rlk: relinKey})
	Encryptor := ckks.NewEncryptorFromPk(Params, publicKey)
	Decryptor := ckks.NewDecryptor(Params, secretKey)

	filters := make([]ckks.Plaintext, filtersAmount)

	for i := range filters {
		filter := make([]complex128, filtersAmount)
		filter[i] = complex(1, 0)
		filters[i] = *Encoder.EncodeNTTAtLvlNew(Params.MaxLevel(), filter, Params.LogSlots())
	}

	if keyChain.bootstrapEnabled {

		log.Log("Util Initialization: Generating bootstrapping key")

		rotations := bootstrappingParams.RotationsForBootstrapping(Params.LogSlots())
		btpGalEl := make([]uint64, len(rotations) + 1)

		for i := range btpGalEl{
			if i == 0{
				btpGalEl[i] = Params.GaloisElementForRowRotation()
			} else {
				btpGalEl[i] = Params.GaloisElementForColumnRotationBy(rotations[i-1])
			}
		}

		rotationKeys := ckks.NewRotationKeySet(Params, btpGalEl)
		rotationKeys.UnmarshalBinary(keyChain.BootstrapGaloisKey)

		bootstrappingKey := ckks.BootstrappingKey{Rlk: relinKey, Rtks: rotationKeys}

		var err error
		var bootstrapper *ckks.Bootstrapper

		log.Log("Util Initialization: Generating bootstrapper")
		bootstrapper, err = ckks.NewBootstrapper(Params, bootstrappingParams, bootstrappingKey)

		if err != nil {
			panic("BOOTSTRAPPER GENERATION ERROR")
		}

		return Utils{
			keyChain.hasSecretKey,
			keyChain.bootstrapEnabled,
			*bootstrappingParams,
			Params,
			*secretKey,
			*publicKey,
			*relinKey,
			*galoisKey,
			*rotationKeys,
			bootstrapper,
			Encoder,
			Evaluator,
			Encryptor,
			Decryptor,
			filters,
			scale,
			log,
		}
	} else {
		return Utils{
			keyChain.hasSecretKey,
			keyChain.bootstrapEnabled,
			*bootstrappingParams,
			Params,
			*secretKey,
			*publicKey,
			*relinKey,
			*galoisKey,
			rlwe.RotationKeySet{},
			&ckks.Bootstrapper{},
			Encoder,
			Evaluator,
			Encryptor,
			Decryptor,
			filters,
			scale,
			log,
		}
	}
}

func NewUtilsFromKeyPair(keyChain KeyChain, scale float64, filtersAmount int, bootstrapEnabled bool, logEnabled bool) Utils{

	if !keyChain.hasSecretKey || len(keyChain.secretKey) == 0 || len(keyChain.PublicKey) == 0{
		panic("No secret key binary provided")
	}

	log := logger.NewLogger(logEnabled)

	bootstrappingParams := ckks.DefaultBootstrapParams[0]
	Params, _ := bootstrappingParams.Params()

	log.Log("Util Initialization: Generating key generator")
	keyGenerator := ckks.NewKeyGenerator(Params)

	log.Log("Util Initialization: Loading private / public key pair")
	secretKey := rlwe.NewSecretKey(Params.Parameters)
	secretKey.UnmarshalBinary(keyChain.secretKey)

	publicKey := rlwe.NewPublicKey(Params.Parameters)
	publicKey.UnmarshalBinary(keyChain.PublicKey)

	log.Log("Util Initialization: Generating relin key")
	relinKey := keyGenerator.GenRelinearizationKey(secretKey)

	log.Log("Util Initialization: Generating galois keys")
	galoisKey := keyGenerator.GenRotationKeysForRotations(getSumElementsKs(Params.LogSlots()), true, secretKey)

	log.Log("Util Initialization: Generating encoder, evaluator, encryptor, decryptor")
	Encoder := ckks.NewEncoder(Params)
	Evaluator := ckks.NewEvaluator(Params, rlwe.EvaluationKey{Rlk: relinKey})
	Encryptor := ckks.NewEncryptorFromPk(Params, publicKey)
	Decryptor := ckks.NewDecryptor(Params, secretKey)

	filters := make([]ckks.Plaintext, filtersAmount)

	for i := range filters {
		filter := make([]complex128, filtersAmount)
		filter[i] = complex(1, 0)
		filters[i] = *Encoder.EncodeNTTAtLvlNew(Params.MaxLevel(), filter, Params.LogSlots())
	}

	btpRotKey := &rlwe.RotationKeySet{}
	bootstrapper := &ckks.Bootstrapper{}

	if bootstrapEnabled {
		log.Log("Util Initialization: Generating bootstrapping key")

		rotations := bootstrappingParams.RotationsForBootstrapping(Params.LogSlots())
		btpRotKey = keyGenerator.GenRotationKeysForRotations(rotations, true, secretKey)

		bootstrappingKey := ckks.BootstrappingKey{Rlk: relinKey, Rtks: btpRotKey}

		var err error

		log.Log("Util Initialization: Generating bootstrapper")
		bootstrapper, err = ckks.NewBootstrapper(Params, bootstrappingParams, bootstrappingKey)

		if err != nil {
			panic("BOOTSTRAPPER GENERATION ERROR")
		}
	}

	return Utils{
		true,
		bootstrapEnabled,
		*bootstrappingParams,
		Params,
		*secretKey,
		*publicKey,
		*relinKey,
		*galoisKey,
		*btpRotKey,
		bootstrapper,
		Encoder,
		Evaluator,
		Encryptor,
		Decryptor,
		filters,
		scale,
		log,
	}

}

func check(err error) {
	if err != nil {
		panic(err)
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
	ciphertext := u.Encryptor.EncryptFastNew(&plaintext)

	return *ciphertext

}

func (u Utils) EncryptToScale(value []float64, scale float64) ckks.Ciphertext {

	// Encode value
	plaintext := u.EncodeToScale(value, scale)

	// Encrypt value
	ciphertext := u.Encryptor.EncryptFastNew(&plaintext)

	return *ciphertext

}

func (u Utils) Decrypt(ciphertext *ckks.Ciphertext) []float64 {

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
