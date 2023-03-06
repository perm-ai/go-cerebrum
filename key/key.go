package key

import (
	"io/ioutil"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/logger"
)

type KeyChain struct {
	ParamsIndex int
	SecretKey   *rlwe.SecretKey
	PublicKey   *rlwe.PublicKey
	RelinKey    *rlwe.RelinearizationKey
	GaloisKey   *rlwe.RotationKeySet
	BtspGalKey  *rlwe.RotationKeySet
}

func GetPow2K(logSlots int) []int {

	ks := []int{}

	for i := 0; i <= logSlots; i++ {
		positive := int(math.Pow(2, float64(i)))
		ks = append(ks, positive)
		ks = append(ks, (-1 * positive))
	}

	sort.Ints(ks[:])

	return ks

}

func GenerateKeys(paramsIndex int, btspEnabled bool, logEnabled bool) KeyChain {

	log := logger.NewLogger(logEnabled)

	paramSet := bootstrapping.DefaultParametersSparse[paramsIndex]
	ckksParams := paramSet.SchemeParams

	Params, _ := ckks.NewParametersFromLiteral(ckksParams)
	
	log.Log("Util Initialization: Generating key generator")
	keyGenerator := ckks.NewKeyGenerator(Params)

	log.Log("Util Initialization: Generating private / public key pair")
	secretKey, publicKey := keyGenerator.GenKeyPair()

	return GenerateKeysFromKeyPair(paramsIndex, secretKey, publicKey, btspEnabled, logEnabled)

}

func GenerateKeysFromKeyPair(paramsIndex int, sk *rlwe.SecretKey, pk *rlwe.PublicKey, btspEnabled bool, logEnabled bool) KeyChain {

	log := logger.NewLogger(logEnabled)

	paramSet := bootstrapping.DefaultParametersSparse[paramsIndex]
	bootstrappingParams := paramSet.BootstrappingParams
	Params, _ := ckks.NewParametersFromLiteral(paramSet.SchemeParams)

	log.Log("Util Initialization: Generating key generator")
	keyGenerator := ckks.NewKeyGenerator(Params)

	publicKey := pk

	if publicKey == nil {
		publicKey = keyGenerator.GenPublicKey(sk)
	}

	log.Log("Util Initialization: Generating relin key")
	relinKey := keyGenerator.GenRelinearizationKey(sk, 2)

	log.Log("Util Initialization: Generating galois keys")
	galoisKey := keyGenerator.GenRotationKeysForRotations(GetPow2K(Params.LogSlots()), true, sk)

	var btpRotationKeys *rlwe.RotationKeySet

	if btspEnabled {
		rotations := bootstrappingParams.RotationsForBootstrapping(Params)
		btpRotationKeys = keyGenerator.GenRotationKeysForRotations(rotations, true, sk)
	}

	return KeyChain{paramsIndex, sk, publicKey, relinKey, galoisKey, btpRotationKeys}

}

func GenerateRelinKey(paramsIndex int, sk *rlwe.SecretKey) *rlwe.RelinearizationKey {

	paramSet := bootstrapping.DefaultParametersSparse[paramsIndex]
	Params, _ := ckks.NewParametersFromLiteral(paramSet.SchemeParams)
	keyGenerator := ckks.NewKeyGenerator(Params)

	return keyGenerator.GenRelinearizationKey(sk, 2)

}

func GenerateRotationKeys(paramsIndex int, sk *rlwe.SecretKey, galEl []uint64, concurrent bool, callback func(galEl uint64, swk *rlwe.SwitchingKey) error) []error{

	paramSet := bootstrapping.DefaultParametersSparse[paramsIndex]
	Params, _ := ckks.NewParametersFromLiteral(paramSet.SchemeParams)
	keyGenerator := NewKeyGenerator(Params, ckks.NewKeyGenerator(Params))

	if concurrent{
		return keyGenerator.GenRotationKeysConcurrent(galEl, sk, callback)
	} else {
		return keyGenerator.GenRotationKeys(galEl, sk, callback)
	}

}

func GenerateKeyPair(paramsIndex int) KeyChain {

	paramSet := bootstrapping.DefaultParametersSparse[paramsIndex]
	Params, _ := ckks.NewParametersFromLiteral(paramSet.SchemeParams)

	keyGenerator := ckks.NewKeyGenerator(Params)

	secretKey, publicKey := keyGenerator.GenKeyPair()

	return KeyChain{SecretKey: secretKey, PublicKey: publicKey}

}

func LoadKeys(dirName string, paramsIndex int, sk bool, pk bool, rlk bool, rotk bool) KeyChain {

	toLoad := [5]bool{sk, pk, rlk, rotk}
	fileNames := [5]string{"secret_key", "public_key", "relin_key", "rotation_key_"}

	for i := range toLoad {
		if toLoad[i] && !fileExist(dirName+"/"+fileNames[i]) {
			panic("File '" + dirName + "/" + fileNames[i] + "' does not exist")
		}
	}

	var skey *rlwe.SecretKey
	var pkey *rlwe.PublicKey
	var rlkey *rlwe.RelinearizationKey
	var rotKeys *rlwe.RotationKeySet

	for i := range toLoad {

		if toLoad[i] && i < 3 {

			var byteArr []byte
			var err error

			byteArr, err = os.ReadFile(dirName + "/" + fileNames[i])
			check(err)

			switch i {
			case 0:
				skey = &rlwe.SecretKey{}
				err = skey.UnmarshalBinary(byteArr)
			case 1:
				pkey = &rlwe.PublicKey{}
				err = pkey.UnmarshalBinary(byteArr)
			case 2:
				rlkey = &rlwe.RelinearizationKey{}
				err = rlkey.UnmarshalBinary(byteArr)

			}
			check(err)

		} else if toLoad[i] && i == 3 {

			files, err := ioutil.ReadDir(dirName)
			check(err)

			galEls := []uint64{}
			rotKeysMap := make(map[uint64]*rlwe.SwitchingKey)

			for i := range files{

				if strings.Contains(files[i].Name(), fileNames[3]){

					keyNameSep := strings.Split(files[i].Name(), "_")
					galEl, err := strconv.Atoi(keyNameSep[len(keyNameSep) - 1])
					check(err)
					rotK, err := os.ReadFile(path.Join(dirName, files[i].Name()))
					check(err)

					swk := &rlwe.SwitchingKey{}
					swk.UnmarshalBinary(rotK)
					rotKeysMap[uint64(galEl)] = swk
					galEls = append(galEls, uint64(galEl))

				}
				
			}

			paramSet := bootstrapping.DefaultParametersSparse[paramsIndex]
			Params, _ := ckks.NewParametersFromLiteral(paramSet.SchemeParams)
			rotKeys := rlwe.NewRotationKeySet(Params.Parameters, galEls)
			
			rotKeys.Keys = rotKeysMap
			
			check(err)
		}

	}

	return KeyChain{paramsIndex, skey, pkey, rlkey, rotKeys, rotKeys}

}

func LoadKeyPairFromBytes(dirName string, paramsIndex int, sk *[]byte, pk *[]byte) KeyChain {

	toLoad := [5]bool{sk != nil, pk != nil}
	fileNames := [5]string{"secret_key", "public_key", "relin_key", "galois_keys", "bootstrap_galois_keys"}

	for i := range toLoad {
		if toLoad[i] && !fileExist(dirName+"/"+fileNames[i]) {
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

			var err error

			switch i {
			case 0:
				skey = &rlwe.SecretKey{}
				err = skey.UnmarshalBinary(*sk)
			case 1:
				pkey = &rlwe.PublicKey{}
				err = pkey.UnmarshalBinary(*pk)
			}
			check(err)

		}

	}

	return KeyChain{paramsIndex, skey, pkey, rlkey, galkey, btpRotKey}

}

func (k KeyChain) DumpKeys(dirName string, sk bool, pk bool, rlk bool, galk bool, btpGalK bool) {

	log := logger.NewLogger(true)
	toSave := [5]bool{sk, pk, rlk, galk, btpGalK}

	if !fileExist(dirName) {
		e := os.Mkdir(dirName, 0777)
		check(e)
	}

	if sk && k.SecretKey == nil {
		panic("Keychain doesn't have secret keys")
	}

	if pk && k.PublicKey == nil {
		panic("Keychain doesn't have public keys")
	}

	if rlk && k.RelinKey == nil {
		panic("Keychain doesn't have relinearlize keys")
	}

	if galk && k.GaloisKey == nil {
		panic("Keychain doesn't have galois keys")
	}

	if btpGalK && k.BtspGalKey == nil {
		panic("Keychain doesn't have bootstrapping galois keys")
	}

	for i := range toSave {
		var byteArr []byte
		var byteErr error
		var name string

		if toSave[i] {

			switch i {
			case 0:
				name = "secret_key"
				log.Log("Marshalling " + name)
				byteArr, byteErr = k.SecretKey.MarshalBinary()
			case 1:
				name = "public_key"
				log.Log("Marshalling " + name)
				byteArr, byteErr = k.PublicKey.MarshalBinary()
			case 2:
				name = "relin_key"
				log.Log("Marshalling " + name)
				byteArr, byteErr = k.RelinKey.MarshalBinary()
			case 3:
				name = "galois_keys"
				log.Log("Marshalling " + name)
				byteArr, byteErr = k.GaloisKey.MarshalBinary()
			case 4:
				name = "bootstrap_galois_keys"
				log.Log("Marshalling " + name)
				byteArr, byteErr = k.BtspGalKey.MarshalBinary()
			}

			check(byteErr)

			f, e := os.Create(dirName + "/" + name)
			check(e)

			log.Log("Saving " + name)
			_, e = f.Write(byteArr)
			check(e)

		}

	}

}

func fileExist(dirName string) bool {
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		return false
	}
	return true
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}