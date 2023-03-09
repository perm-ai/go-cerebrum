package key

import (
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"

	"bufio"
	"unsafe"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/tuneinsight/lattigo/v4/rlwe"
)

func EncodeSwitchingKey(swk *rlwe.SwitchingKey, pointer int, data []byte) (int, error) {

	var err error
	var inc int

	data[pointer] = uint8(len(swk.Value))
	pointer++
	data[pointer] = uint8(len(swk.Value[0]))
	pointer++

	for i := range swk.Value {
		for _, el := range swk.Value[i] {

			if inc, err = el.Encode64(data[pointer:]); err != nil {
				return pointer, err
			}
			pointer += inc
		}
	}

	// for j := 0; j < len(swk.Value); j++ {

	// 	if inc, err = swk.Value[j][0].Decode64(data[pointer : pointer+swk.Value[j][0].MarshalBinarySize()]); err != nil {
	// 		return pointer, err
	// 	}

	// 	pointer += inc

	// 	if inc, err = swk.Value[j][1].Decode64(data[pointer : pointer+swk.Value[j][1].MarshalBinarySize()]); err != nil {
	// 		return pointer, err
	// 	}

	// 	pointer += inc
	// }

	return pointer, nil
}

func DecodeSwitchingKey(swk *rlwe.SwitchingKey, data []byte) (pointer int, err error) {

	decompRNS := int(data[0])
	decompBIT := int(data[1])

	pointer = 2

	swk.Value = make([][]rlwe.CiphertextQP, decompRNS)

	var inc int

	for i := range swk.Value {

		swk.Value[i] = make([]rlwe.CiphertextQP, decompBIT)

		for j := range swk.Value[i] {

			if inc, err = swk.Value[i][j].Decode64(data[pointer:]); err != nil {
				return
			}
			pointer += inc
		}
	}

	return
}

func MarshalBinary(rtks *rlwe.RotationKeySet) (data []byte, err error) {
	
	data = make([]byte, rtks.MarshalBinarySize())

	pointer := int(0)

	var inc int
	for galEL, key := range rtks.Keys {

		binary.BigEndian.PutUint64(data[pointer:], galEL)
		pointer += 8

		if inc, err = key.Encode(data[pointer:]); err != nil {
			return nil, err
		}

		pointer += inc
	}

	return data, nil
}

// func UnmarshalBinaryBatch(rtks *rlwe.RotationKeySet, keyFile *os.File) (err error) {

// 	rtks.Keys = make(map[uint64]*rlwe.SwitchingKey)

// 	fileInfo, e := keyFile.Stat()
// 	check(e)

// 	fileLen := fileInfo.Size()
// 	keyLen := 0
// 	pointer := 0
// 	r4 := bufio.NewReaderSize(keyFile, 1073741824)
// 	data, err := r4.Peek(1073741824)
// 	check(err)

// 	for len(data) > 0 {

// 		galEl := uint64(binary.BigEndian.Uint32(data))
// 		fmt.Println(galEl)
// 		data = data[4:]
// 		// cut data by 4?
// 		swk := new(rlwe.SwitchingKey)
// 		var inc int
// 		if inc, err = DecodeSwitchingKey(swk, data); err != nil {
// 			return err
// 		}

// 		if keyLen == 0 {
// 			keyLen = 4 + inc
// 		}

// 		data = data[inc:]
// 		rtks.Keys[galEl] = swk
// 		pointer += 4 + inc

// 		if len(data) < keyLen {
// 			if (int(fileLen) - pointer) < 1073741824 {

// 				data = ReadFromDesinatedPointer(pointer, (int(fileLen) - pointer), keyFile)
// 				continue
// 			}

// 			data = ReadFromDesinatedPointer(pointer, 1073741824, keyFile)
// 		}
// 	}

// 	return nil
// }

func UnmarshalGaloisFromS3(rtks *rlwe.RotationKeySet, keyS3key string, s3Client *s3.Client, bucketName string) (err error) {

	headParams := &s3.HeadObjectInput{
		Bucket: &bucketName,
		Key: &keyS3key,
	}

	headObject, s3HeadErr := s3Client.HeadObject(context.TODO(), headParams)
	if s3HeadErr != nil {
		return s3HeadErr
	}

	fileLen := headObject.ContentLength

	data, reqByteErr := requestBytesFromS3(s3Client, bucketName, keyS3key, 0, 1073741824, int(fileLen))
	if reqByteErr != nil {
		return reqByteErr
	}

	keyLen := 0
	pointer := 0

	for len(data) > 0 {

		galEl := uint64(binary.BigEndian.Uint32(data))

		data = data[4:]
		swk := new(rlwe.SwitchingKey)
		var inc int
		if inc, err = DecodeSwitchingKey(swk, data); err != nil {
			return err
		}

		if keyLen == 0 {
			keyLen = 4 + inc
		}

		data = data[inc:]
		rtks.Keys[galEl] = swk
		pointer += 4 + inc

		if len(data) < keyLen {
			data, reqByteErr = requestBytesFromS3(s3Client, bucketName, keyS3key, pointer, 1073741824, int(fileLen))
			if reqByteErr != nil {
				return reqByteErr
			}
		}
	}

	return nil
}


func requestBytesFromS3(s3Client *s3.Client, bucketName string, key string, start int, getLen int, fileLength int) ([]byte, error) {

	var byteRange string

	if start + 1073741824 < fileLength{
		byteRange = fmt.Sprintf("%d-%d", start, start + getLen)
	} else {
		byteRange = fmt.Sprintf("%d-%d", start, fileLength)
	}

	getParams := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key: &key,
		Range: &byteRange,
	}

	s3Object, err:= s3Client.GetObject(context.TODO(), getParams)
	if err != nil {
		return nil, err
	}

	data, readErr := ioutil.ReadAll(s3Object.Body)
	if readErr != nil {
		return nil, readErr
	}

	s3Object.Body.Close()

	return data, nil

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