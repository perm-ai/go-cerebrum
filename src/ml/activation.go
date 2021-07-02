package ml

import "github.com/ldsec/lattigo/v2/ckks"


type Activation interface {

	forward() 	ckks.Ciphertext
	backward()	ckks.Ciphertext

}