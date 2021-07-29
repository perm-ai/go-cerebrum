package cmd

import (
	"fmt"
	"math"
	"os"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

func Decrypt(keyPath string, data string) {

	log := logger.NewLogger(true)

	if keyPath == "" {
		panic("No key path provided")
	}

	log.Log("Loading Keys")
	keys := key.LoadKeys(keyPath, 0, true, true, false, false, false)

	log.Log("Generating Utils")
	utils := utility.NewDecryptionUtils(keys, math.Pow(2, 35), true)

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
