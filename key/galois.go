package key

import (
	"encoding/binary"
	"fmt"
	"os"

	"bufio"
	"unsafe"

	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/rlwe"
)

func Encode(swk *rlwe.SwitchingKey, pointer int, data []byte) (int, error) {

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

func GaloisDecode(swk *rlwe.SwitchingKey, data []byte) (pointer int, err error) {

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

		binary.BigEndian.PutUint32(data[pointer:pointer+4], uint32(galEL))
		pointer += 4

		if pointer, err = Encode(key, pointer, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func UnmarshalBinaryBatch(rtks *rlwe.RotationKeySet, keyFile *os.File) (err error) {

	rtks.Keys = make(map[uint64]*rlwe.SwitchingKey)

	fileInfo, e := keyFile.Stat()
	check(e)

	fileLen := fileInfo.Size()
	keyLen := 0
	pointer := 0
	r4 := bufio.NewReaderSize(keyFile, 1073741824)
	data, err := r4.Peek(1073741824)
	check(err)

	for len(data) > 0 {

		galEl := uint64(binary.BigEndian.Uint32(data))
		fmt.Println(galEl)
		data = data[4:]
		// cut data by 4?
		swk := new(rlwe.SwitchingKey)
		var inc int
		if inc, err = GaloisDecode(swk, data); err != nil {
			return err
		}

		if keyLen == 0 {
			keyLen = 4 + inc
		}

		data = data[inc:]
		rtks.Keys[galEl] = swk
		pointer += 4 + inc

		if len(data) < keyLen {
			if (int(fileLen) - pointer) < 1073741824 {

				data = ReadFromDesinatedPointer(pointer, (int(fileLen) - pointer), keyFile)
				continue
			}

			data = ReadFromDesinatedPointer(pointer, 1073741824, keyFile)
		}
	}

	return nil
}

func ReadFromDesinatedPointer(pointer int, size int, keyFile *os.File) []byte {

	// continue reading from "pointer" till "pointer + size"

	_, err := keyFile.Seek(int64(pointer), 0)
	check(err)
	r4 := bufio.NewReaderSize(keyFile, size)
	data, err := r4.Peek(size)
	check(err)
	return data
}

func IntToByteArray(num int64) []byte {
	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[i] = byt
	}
	return arr
}
