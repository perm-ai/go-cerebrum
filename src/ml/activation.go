package ml

import "github.com/ldsec/lattigo/v2/ckks"


type Activation interface {

	forward(input ckks.Ciphertext) 	ckks.Ciphertext
	backward(input ckks.Ciphertext)	ckks.Ciphertext

}