package utility

import "github.com/ldsec/lattigo/v2/ckks"

type Utils struct {
	BootstrappingParams ckks.BootstrappingParameters
	Params              ckks.Parameters
	secretKey           ckks.SecretKey
	PublicKey           ckks.PublicKey
	RelinKey            ckks.EvaluationKey
	BootstrapingKey     ckks.BootstrappingKey
	GaloisKey           ckks.RotationKeys

	Bootstrapper *ckks.Bootstrapper
	Encoder      ckks.Encoder
	Evaluator    ckks.Evaluator
	Encryptor    ckks.Encryptor
	Decryptor    ckks.Decryptor
}

func NewUtils() Utils {

	Params := ckks.DefaultBootstrapSchemeParams[0]
	bootstrappingParams := ckks.DefaultBootstrapParams[0]

	Params.SetScale(40)

	keyGenerator := ckks.NewKeyGenerator(Params)

	secretKey, publicKey := keyGenerator.GenKeyPairSparse(bootstrappingParams.H)
	relinKey := keyGenerator.GenRelinKey(secretKey)
	galoisKey := keyGenerator.GenRotationKeysPow2(secretKey)

	Encoder := ckks.NewEncoder(Params)
	Evaluator := ckks.NewEvaluator(Params)
	Encryptor := ckks.NewEncryptorFromPk(Params, publicKey)
	Decryptor := ckks.NewDecryptor(Params, secretKey)

	bootstrappingKey := keyGenerator.GenBootstrappingKey(Params.LogSlots(), bootstrappingParams, secretKey)

	var err error
	var bootstrapper *ckks.Bootstrapper

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
	}

}

func (u Utils) Encode(value []float64) ckks.Plaintext {

	// Encode value
	plaintext := ckks.NewPlaintext(&u.Params, u.Params.MaxLevel(), u.Params.Scale())
	u.Encoder.EncodeCoeffs(value, plaintext)

	return *plaintext

}

func (u Utils) GenerateFilledArray(value float64) []float64 {

	arr := make([]float64, u.Params.LogSlots())
	for i := range arr {
		arr[i] = value
	}

	return arr

}

func (u Utils) EncodeToScale(value []float64, scale float64) ckks.Plaintext {

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

	decoded := u.Encoder.DecodeCoeffs(decrypted)

	return decoded

}
