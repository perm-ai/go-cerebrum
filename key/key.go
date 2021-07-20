package key

import (
	"math"
	"os"
	"sort"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/rlwe"
	"github.com/perm-ai/go-cerebrum/logger"
)


type KeyChain struct {

	ParamsIndex int
	SecretKey	*rlwe.SecretKey
	PublicKey	*rlwe.PublicKey
	RelinKey	*rlwe.RelinearizationKey
	GaloisKey	*rlwe.RotationKeySet
	BtspGalKey	*rlwe.RotationKeySet

}

func getPow2K(logSlots int) []int {

	ks := []int{}

	for i := 0; i <= logSlots; i++{
		positive := int(math.Pow(2, float64(i)))
		ks = append(ks, positive)
		ks = append(ks, (-1 * positive))
	}

	sort.Ints(ks[:])

	return ks

}

func GenerateKeys(paramsIndex int, btspEnabled bool, logEnabled bool) KeyChain {

	log := logger.NewLogger(logEnabled)

	bootstrappingParams := ckks.DefaultBootstrapParams[paramsIndex]
	Params, _ := bootstrappingParams.Params()

	log.Log("Util Initialization: Generating key generator")
	keyGenerator := ckks.NewKeyGenerator(Params)

	log.Log("Util Initialization: Generating private / public key pair")
	secretKey, publicKey := keyGenerator.GenKeyPairSparse(bootstrappingParams.H)

	return GenerateKeysFromKeyPair(paramsIndex, secretKey, publicKey, btspEnabled, logEnabled)

}

func GenerateKeysFromKeyPair(paramsIndex int, sk *rlwe.SecretKey, pk *rlwe.PublicKey, btspEnabled bool, logEnabled bool) KeyChain {

	log := logger.NewLogger(logEnabled)

	bootstrappingParams := ckks.DefaultBootstrapParams[paramsIndex]
	Params, _ := bootstrappingParams.Params()

	log.Log("Util Initialization: Generating key generator")
	keyGenerator := ckks.NewKeyGenerator(Params)

	publicKey := pk

	if publicKey == nil {
		publicKey = keyGenerator.GenPublicKey(sk)
	}

	log.Log("Util Initialization: Generating relin key")
	relinKey := keyGenerator.GenRelinearizationKey(sk, 2)

	log.Log("Util Initialization: Generating galois keys")
	galoisKey := keyGenerator.GenRotationKeysForRotations(getPow2K(Params.LogSlots()), true, sk)

	var btpRotationKeys *rlwe.RotationKeySet

	if btspEnabled {
		rotations := bootstrappingParams.RotationsForBootstrapping(Params.LogSlots())
		btpRotationKeys = keyGenerator.GenRotationKeysForRotations(rotations, true, sk)
	}

	return KeyChain{paramsIndex, sk, publicKey, relinKey, galoisKey, btpRotationKeys,}

}

func GenerateKeyPair(paramsIndex int) KeyChain {

	bootstrappingParams := ckks.DefaultBootstrapParams[paramsIndex]
	Params, _ := bootstrappingParams.Params()

	keyGenerator := ckks.NewKeyGenerator(Params)

	secretKey, publicKey := keyGenerator.GenKeyPairSparse(bootstrappingParams.H)

	return KeyChain{SecretKey: secretKey, PublicKey: publicKey}

}

func LoadKeys(dirName string, paramsIndex int, sk bool, pk bool, rlk bool, galk bool, btpGalK bool) KeyChain {

	toLoad := [5]bool{sk, pk, rlk, galk, btpGalK}
	fileNames := [5]string{"secret_key", "public_key", "relin_keys", "galois_keys", "bootstrap_galois_keys"}

	for i := range toLoad {
		if toLoad[i] && !fileExist(dirName + "/" + fileNames[i]){
			panic("File '" + dirName + "/" + fileNames[i] + "' does not exist")
		}
	}
	
	var skey *rlwe.SecretKey
	var pkey *rlwe.PublicKey
	var rlkey *rlwe.RelinearizationKey
	var galkey *rlwe.RotationKeySet
	var btpRotKey *rlwe.RotationKeySet

	for i := range toLoad {

		if toLoad[i] {

			var byteArr []byte
			var err error

			byteArr, err = os.ReadFile(dirName + "/" + fileNames[i])
			check(err)

			switch i{
			case 0:
				skey = &rlwe.SecretKey{}
				err = skey.UnmarshalBinary(byteArr)
			case 1:
				pkey = &rlwe.PublicKey{}
				err = pkey.UnmarshalBinary(byteArr)
			case 2:
				rlkey = &rlwe.RelinearizationKey{}
				err = rlkey.UnmarshalBinary(byteArr)
			case 3:
				galkey = &rlwe.RotationKeySet{}
				err = galkey.UnmarshalBinary(byteArr)
			case 4:
				btpRotKey = &rlwe.RotationKeySet{}
				err = btpRotKey.UnmarshalBinary(byteArr)
			}
			check(err)

		}
		
	}

	return KeyChain{paramsIndex, skey, pkey, rlkey, galkey, btpRotKey}

}

func (k KeyChain) DumpKeys(dirName string, sk bool, pk bool, rlk bool, galk bool, btpGalK bool){

	toSave := [5]bool{sk, pk, rlk, galk, btpGalK}

	if !fileExist(dirName){
		os.Mkdir(dirName, 0777)
	}

	if sk && k.SecretKey == nil{
		panic("Keychain doesn't have secret keys")
	}

	if pk && k.PublicKey == nil{
		panic("Keychain doesn't have public keys")
	}

	if rlk && k.RelinKey == nil{
		panic("Keychain doesn't have relinearlize keys")
	}

	if galk && k.GaloisKey == nil{
		panic("Keychain doesn't have galois keys")
	}

	if btpGalK && k.BtspGalKey == nil{
		panic("Keychain doesn't have bootstrapping galois keys")
	}

	for i := range toSave{
		var byteArr []byte
		var byteErr error
		var name string

		switch i{
		case 0:
			name = "secret_key"
			byteArr, byteErr = k.SecretKey.MarshalBinary()
		case 1:
			name = "public_key"
			byteArr, byteErr = k.PublicKey.MarshalBinary()
		case 2:
			name = "relin_keys"
			byteArr, byteErr = k.RelinKey.MarshalBinary()
		case 3:
			name = "galois_keys"
			byteArr, byteErr = k.GaloisKey.MarshalBinary()
		case 4:
			name = "bootstrap_galois_keys"
			byteArr, byteErr = k.BtspGalKey.MarshalBinary()
		}

		check(byteErr)

		f, e := os.Create(dirName + "/" + name)
		check(e)

		_, e = f.Write(byteArr)
		check(e)

	}

}

func fileExist(dirName string) bool {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		return false
	}
	return true
}

func check(err error){
	if err != nil {
		panic(err)
	}
}