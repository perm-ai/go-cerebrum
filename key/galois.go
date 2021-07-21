package key

import (
	"encoding/binary"

	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/rlwe"
)

func encode(swk *rlwe.SwitchingKey, pointer int, data []byte) (int, error) {

	var err error
	var inc int

	data[pointer] = uint8(len(swk.Value))

	pointer++

	for j := 0; j < len(swk.Value); j++ {

		if inc, err = swk.Value[j][0].WriteTo(data[pointer : pointer+swk.Value[j][0].GetDataLen(true)]); err != nil {
			return pointer, err
		}

		pointer += inc

		if inc, err = swk.Value[j][1].WriteTo(data[pointer : pointer+swk.Value[j][1].GetDataLen(true)]); err != nil {
			return pointer, err
		}

		pointer += inc
	}

	return pointer, nil
}

func decode(swk *rlwe.SwitchingKey, data []byte) (pointer int, err error) {

	decomposition := int(data[0])

	pointer = 1

	swk.Value = make([][2]*ring.Poly, decomposition)

	var inc int

	for j := 0; j < decomposition; j++ {

		swk.Value[j][0] = new(ring.Poly)
		if inc, err = swk.Value[j][0].DecodePolyNew(data[pointer:]); err != nil {
			return pointer, err
		}
		pointer += inc

		swk.Value[j][1] = new(ring.Poly)
		if inc, err = swk.Value[j][1].DecodePolyNew(data[pointer:]); err != nil {
			return pointer, err
		}
		pointer += inc

	}

	return pointer, nil
}

func MarshalBinary(rtks *rlwe.RotationKeySet) (data []byte, err error) {

	data = make([]byte, rtks.GetDataLen(true))

	pointer := int(0)

	for galEL, key := range rtks.Keys {

		binary.BigEndian.PutUint32(data[pointer:pointer+2], uint32(galEL))
		pointer += 2

		if pointer, err = encode(key, pointer, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func UnmarshalBinary(rtks *rlwe.RotationKeySet, data []byte) (err error) {

	rtks.Keys = make(map[uint64]*rlwe.SwitchingKey)

	for len(data) > 0 {

		galEl := uint64(binary.BigEndian.Uint32(data))
		data = data[4:]

		swk := new(rlwe.SwitchingKey)
		var inc int
		if inc, err = decode(swk, data); err != nil {
			return err
		}
		data = data[inc:]
		rtks.Keys[galEl] = swk

	}

	return nil
}
