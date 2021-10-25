package utility

import "github.com/ldsec/lattigo/v2/ckks"


func Clone3dCiphertext(ct [][][]*ckks.Ciphertext) [][][]*ckks.Ciphertext {

	newArray := make([][][]*ckks.Ciphertext, len(ct))

	for r := range ct{
		newArray[r] = make([][]*ckks.Ciphertext, len(ct[r]))
		for c := range ct[r]{
			newArray[r][c] = make([]*ckks.Ciphertext, len(ct[r][c]))
			for d := range ct[r][c]{
				newArray[r][c][d] = ct[r][c][d].CopyNew()
			}
		}
	}

	return newArray

}

func Clone1dCiphertext(ct []*ckks.Ciphertext) []*ckks.Ciphertext{

	newArray := make([]*ckks.Ciphertext, len(ct))

	for i := range ct{
		newArray[i] = ct[i].CopyNew()
	}

	return newArray

}