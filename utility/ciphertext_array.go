package utility

import "github.com/tuneinsight/lattigo/v4/rlwe"


func Clone3dCiphertext(ct [][][]*rlwe.Ciphertext) [][][]*rlwe.Ciphertext {

	newArray := make([][][]*rlwe.Ciphertext, len(ct))

	for r := range ct{
		newArray[r] = make([][]*rlwe.Ciphertext, len(ct[r]))
		for c := range ct[r]{
			newArray[r][c] = make([]*rlwe.Ciphertext, len(ct[r][c]))
			for d := range ct[r][c]{
				newArray[r][c][d] = ct[r][c][d].CopyNew()
			}
		}
	}

	return newArray

}

func Clone1dCiphertext(ct []*rlwe.Ciphertext) []*rlwe.Ciphertext{

	newArray := make([]*rlwe.Ciphertext, len(ct))

	for i := range ct{
		newArray[i] = ct[i].CopyNew()
	}

	return newArray

}