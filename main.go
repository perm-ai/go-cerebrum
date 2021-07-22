package main

import (
	"github.com/ldsec/lattigo/v2/rlwe"
	"os"
)

func main() {

	galoisKey, _ := os.ReadFile("keyChain/galois_keys")
	
	var galkey *rlwe.RotationKeySet
	galkey = &rlwe.RotationKeySet{}
	UnmarshalBinary(galKey, galoisKey)
	
}
