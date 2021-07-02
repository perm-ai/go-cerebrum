package utility

import (
	// "math"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/GO-HEML-prototype/src/logger"
)

type Utils struct {
	BootstrappingParams ckks.BootstrappingParameters
	Params              ckks.Parameters
	secretKey           ckks.SecretKey
	PublicKey           ckks.PublicKey
	RelinKey            ckks.EvaluationKey
	BootstrapingKey     ckks.BootstrappingKey
	GaloisKey           ckks.RotationKeys

	Bootstrapper 		*ckks.Bootstrapper
	Encoder      		ckks.Encoder
	Evaluator    		ckks.Evaluator
	Encryptor    		ckks.Encryptor
	Decryptor    		ckks.Decryptor

	Filters		 		[]ckks.Plaintext
	log 		 		logger.Logger
}

func NewUtils(scale float64, filtersAmount int, bootstrapEnabled bool, logEnabled bool) Utils {

	log := logger.NewLogger(logEnabled)

	Params := ckks.DefaultBootstrapSchemeParams[0]
	bootstrappingParams := ckks.DefaultBootstrapParams[0]

	Params.SetScale(scale)

	log.Log("Util Initialization: Generating key generator")
	keyGenerator := ckks.NewKeyGenerator(Params)

	log.Log("Util Initialization: Generating keys")
	secretKey, publicKey := keyGenerator.GenKeyPairSparse(bootstrappingParams.H)
	relinKey := keyGenerator.GenRelinKey(secretKey)
	galoisKey := keyGenerator.GenRotationKeysPow2(secretKey)

	log.Log("Util Initialization: Generating encoder, evaluator, encryptor, decryptor")
	Encoder := ckks.NewEncoder(Params)
	Evaluator := ckks.NewEvaluator(Params)
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
		var bootstrappingKey *ckks.BootstrappingKey
		bootstrappingKey = keyGenerator.GenBootstrappingKey(Params.LogSlots(), bootstrappingParams, secretKey)

		var err error
		var bootstrapper *ckks.Bootstrapper

		log.Log("Util Initialization: Generating bootstrapper")
		bootstrapper, err = ckks.NewBootstrapper(Params, bootstrappingParams, bootstrappingKey)

		if err != nil {
			panic("BOOTSTRAPPER GENERATION ERROR")
		}

		return Utils{
			*bootstrappingParams,
			*Params,
			*secretKey,
			*publicKey,
			*relinKey,
			*bootstrappingKey,
			*galoisKey,
			bootstrapper,
			Encoder,
			Evaluator,
			Encryptor,
			Decryptor,
			filters,
			log,
		}
	} else {
		return Utils{
			*bootstrappingParams,
			*Params,
			*secretKey,
			*publicKey,
			*relinKey,
			ckks.BootstrappingKey{},
			*galoisKey,
			&ckks.Bootstrapper{},
			Encoder,
			Evaluator,
			Encryptor,
			Decryptor,
			filters,
			log,
		}
	}

}

func (u Utils) Float64ToComplex128(value []float64) []complex128 {

	cmplx := make([]complex128, len(value))
	for i := range value {
		cmplx[i] = complex(value[i], 0)
	}
	return cmplx

}

func (u Utils) Complex128ToFloat64(value []complex128) []float64 {

	flt := make([]float64, len(value))
	for i := range value {
		flt[i] = real(value[i])
	}
	return flt

}

func (u Utils) GenerateFilledArray(value float64) []float64 {

	arr := make([]float64, u.Params.Slots())
	for i := range arr {
		arr[i] = value
	}

	return arr

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
	plaintext := ckks.NewPlaintext(&u.Params, u.Params.MaxLevel(), scale)
	u.Encoder.Encode(plaintext, u.Float64ToComplex128(value), u.Params.LogSlots())

	return *plaintext

}

// Encode into float coefficient
func (u Utils) EncodeCoeffs(value []float64) ckks.Plaintext {

	// Encode value
	plaintext := ckks.NewPlaintext(&u.Params, u.Params.MaxLevel(), u.Params.Scale())
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
	plaintext := ckks.NewPlaintext(&u.Params, u.Params.MaxLevel(), scale)
	u.Encoder.EncodeCoeffs(value, plaintext)

	return *plaintext

}

func (u Utils) Encrypt(value []float64) ckks.Ciphertext {

	// Encode value
	plaintext := u.Encode(value)

	// Encrypt value
	ciphertext := u.Encryptor.EncryptFastNew(&plaintext)

	return *ciphertext

}

func (u Utils) Decrypt(ciphertext *ckks.Ciphertext) []float64 {

	decrypted := u.Decryptor.DecryptNew(ciphertext)

	decoded := u.Decode(decrypted)

	return decoded

}
