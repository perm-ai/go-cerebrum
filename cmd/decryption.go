package cmd

import (
	"fmt"
	"os"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

func Decrypt(key string, data string) {

	log := logger.NewLogger(true)

	if key == "" {
		panic("No key path provided")
	}

	log.Log("Loading Keys")
	keys := utility.LoadKeyPair(key)

	log.Log("Generating Utils")
	utils := utility.NewDecryptionUtils(keys, true)

	log.Log("Reading binary file")
	ctBin, readErr := os.ReadFile(data)
	check(readErr)

	log.Log("Loading binary")
	ct := ckks.Ciphertext{}
	unmarshalErr := ct.UnmarshalBinary(ctBin)
	check(unmarshalErr)

	log.Log("Decrypting")
	decrypted := utils.Decrypt(&ct)

	log.Log(fmt.Sprintf("Result: %f", decrypted[:100]))

}
