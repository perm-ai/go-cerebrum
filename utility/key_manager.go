package utility

import (
	"fmt"
	"os"

	"github.com/perm-ai/go-cerebrum/logger"
)

type KeyChain struct {
	hasSecretKey       bool
	bootstrapEnabled   bool
	secretKey          []byte
	PublicKey          []byte
	RelinKey           []byte
	GaloisKey          []byte
	BootstrapGaloisKey []byte
}

func LoadKeyPair(directoryPath string) KeyChain {

	log := logger.NewLogger(true)

	log.Log("Loading secret")
	if _, err := os.Stat(directoryPath + "/secret_key"); err == nil {

		secretByte, err1 := os.ReadFile(directoryPath + "/secret_key")
		check(err1)

		log.Log("Decoding public")
		publicByte, err2 := os.ReadFile(directoryPath + "/public_key")
		check(err2)

		return KeyChain{
			hasSecretKey: true,
			secretKey:    secretByte,
			PublicKey:    publicByte,
		}

	} else {
		panic("No secret key found at '" + directoryPath + "/secret_key'")
	}

}

func LoadKey(directoryPath string) KeyChain {

	log := logger.NewLogger(true)
	hasSecret := false
	secretByte := []byte{}

	log.Log("Decoding secret")
	if _, err := os.Stat(directoryPath + "/secret_key"); err == nil {

		var err1 error
		secretByte, err1 = os.ReadFile(directoryPath + "/secret_key")
		check(err1)
		hasSecret = true

	}

	log.Log("Decoding public")
	publicByte, err2 := os.ReadFile(directoryPath + "/public_key")
	check(err2)

	log.Log("Decoding relin")
	relinByte, err3 := os.ReadFile(directoryPath + "/relin_key")
	check(err3)

	log.Log("Decoding galois")
	galoisByte, err4 := os.ReadFile(directoryPath + "/galois_key")
	check(err4)

	bootstrapEnabled := false
	bootstrappingGalois := []byte{}

	if _, err := os.Stat(directoryPath + "/bootstrap_galois_key"); err == nil {

		log.Log("Decoding btp glk")
		var err5 error
		bootstrappingGalois, err5 = os.ReadFile(directoryPath + "/bootstrap_galois_key")
		check(err5)
		bootstrapEnabled = true

	}

	return KeyChain{
		hasSecret,
		bootstrapEnabled,
		secretByte,
		publicByte,
		relinByte,
		galoisByte,
		bootstrappingGalois,
	}

}

func (u Utils) DumpKeyPair(directoryPath string) {

	e := os.Mkdir(directoryPath, 0777)
	check(e)

	var file *os.File
	var err error

	if u.hasSecretKey {

		// Dumping secret key into byte array
		u.log.Log("Dumping SK")
		secret, err1 := u.secretKey.MarshalBinary()
		check(err1)
		u.log.Log(fmt.Sprintf("SK size: %d bytes", len(secret)))

		// dumping byte array into file
		file, err = os.Create(directoryPath + "/secret_key")
		check(err)
		file.Write(secret)

		// free memory
		secret = nil

	} else {
		panic("No secret key available for export.")
	}

	// Dumping public key into byte array
	u.log.Log("Dumping PK")
	public, err2 := u.PublicKey.MarshalBinary()
	check(err2)
	u.log.Log(fmt.Sprintf("PK size: %d bytes", len(public)))

	// dumping byte array into file
	file, err = os.Create(directoryPath + "/public_key")
	check(err)
	file.Write(public)

	// free memory
	public = nil

}

func (u Utils) DumpKeys(directoryPath string) {

	e := os.Mkdir(directoryPath, 0777)
	check(e)

	var file *os.File
	var err error

	if u.hasSecretKey {

		// Dumping secret key into byte array
		u.log.Log("Dumping SK")
		secret, err1 := u.secretKey.MarshalBinary()
		check(err1)
		u.log.Log(fmt.Sprintf("SK size: %d bytes", len(secret)))

		// dumping byte array into file
		file, err = os.Create(directoryPath + "/secret_key")
		check(err)
		file.Write(secret)

		// free memory
		secret = nil

	}

	// Dumping public key into byte array
	u.log.Log("Dumping PK")
	public, err2 := u.PublicKey.MarshalBinary()
	check(err2)
	u.log.Log(fmt.Sprintf("PK size: %d bytes", len(public)))

	// dumping byte array into file
	file, err = os.Create(directoryPath + "/public_key")
	check(err)
	file.Write(public)

	// free memory
	public = nil

	// Dumping relinearlize key into byte array
	u.log.Log("Dumping RLK")
	relin, err3 := u.RelinKey.MarshalBinary()
	check(err3)
	u.log.Log(fmt.Sprintf("RelinK size: %d bytes", len(relin)))

	// dumping byte array into file
	file, err = os.Create(directoryPath + "/relin_key")
	check(err)
	file.Write(relin)

	// free memory
	relin = nil

	// Dumping galois key into byte array
	u.log.Log("Dumping GLK")
	galois, err4 := u.GaloisKey.MarshalBinary()
	check(err4)
	u.log.Log(fmt.Sprintf("GLK size: %d bytes", len(galois)))

	// dumping byte array into file
	file, err = os.Create(directoryPath + "/galois_key")
	check(err)
	file.Write(galois)

	// free memory
	galois = nil

	if u.bootstrapEnabled {

		u.log.Log("Dumping BTP_GLK")

		// Dumping bootstrapping galois key into byte array
		bootstrappingGalois, err5 := u.BtspGaloisKey.MarshalBinary()
		check(err5)
		u.log.Log(fmt.Sprintf("BTP GLK size: %d bytes", len(bootstrappingGalois)))

		// dumping byte array into file
		file, err = os.Create(directoryPath + "/bootstrap_galois_key")
		check(err)
		file.Write(bootstrappingGalois)

		// free memory
		bootstrappingGalois = nil

	}

}
