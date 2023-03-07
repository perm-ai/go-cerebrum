package management

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/tuneinsight/lattigo/v4/ckks"
	"github.com/tuneinsight/lattigo/v4/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
)


type LoginRequest struct {
	Email 	 string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status 	string `json:"status"`
	Message string `json:"message"`
	Token 	string `json:"token"`
	Id 		string `json:"id"`
}

type DownloadKeyRequest struct {
	KeyId 	 string `json:"keyId"`
}

type DownloadKeyResponse struct {
	Status 	string   `json:"status"`
	Message string   `json:"message"`
	Pk		string   `json:"pk"`
	Rlk		string   `json:"rlk"`
	Rotk	[]string `json:"rotk"`
}

func GetDownloadKeyUrl(email string, password string, keyId string) (DownloadKeyResponse, error) {

	baseUrl := "https://edmt-api.perm.ai"
	loginUrl := baseUrl + "/user/login"
	downloadUrl := baseUrl + "/key/download"

	client := http.Client{}

	// Encode login request body
	loginRequestBody, err := json.Marshal(LoginRequest{
		Email: email,
		Password: password,
	})

	if err != nil{
		return DownloadKeyResponse{}, err
	}

	loginRequestBodyBuffer := bytes.NewBuffer(loginRequestBody)

	// Make request
	loginRequest, err := http.NewRequest("POST", loginUrl, loginRequestBodyBuffer)
	loginRequest.Header.Add("Content-Type", "application/json")

	if err != nil{
		return DownloadKeyResponse{}, err
	}

	// Send login request
	loginResponse, err := client.Do(loginRequest)

	if err != nil{
		return DownloadKeyResponse{}, err
	}

	// Parse return data
	loginResponseByte, err := ioutil.ReadAll(loginResponse.Body)

	if err != nil{
		return DownloadKeyResponse{}, err
	}

	var login LoginResponse
	err = json.Unmarshal(loginResponseByte, &login)
	if err != nil{
		return DownloadKeyResponse{}, err
	}

	if login.Status == "success"{

		// Encode login request body
		downloadRequestBody, err := json.Marshal(DownloadKeyRequest{
			KeyId: keyId,
		})

		if err != nil{
			return DownloadKeyResponse{}, err
		}

		downloadRequestBodyBuffer := bytes.NewBuffer(downloadRequestBody)

		// Make request
		downloadRequest, err := http.NewRequest("POST", downloadUrl, downloadRequestBodyBuffer)
		downloadRequest.Header.Add("Content-Type", "application/json")
		downloadRequest.Header.Add("Authorization", "Bearer " + login.Token)

		if err != nil{
			return DownloadKeyResponse{}, err
		}

		// Send download url request
		downloadResponse, err := client.Do(downloadRequest)

		if err != nil{
			return DownloadKeyResponse{}, err
		}

		// Parse return data
		downloadResponseByte, err := ioutil.ReadAll(downloadResponse.Body)

		if err != nil{
			return DownloadKeyResponse{}, err
		}

		var download DownloadKeyResponse
		err = json.Unmarshal(downloadResponseByte, &download)

		if err != nil{
			return DownloadKeyResponse{}, err
		} else {
			return download, nil
		}

	} else {
		fmt.Println(login)
		return DownloadKeyResponse{}, errors.New("unsuccessful login")
	}

}

func DownloadFromS3(url string) ([]byte, error) {

	client := http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil{
		return []byte{}, err
	}

	res, err := client.Do(req)

	if err != nil{
		return []byte{}, err
	}

	byteValue, err := ioutil.ReadAll(res.Body)

	return byteValue, err

}

func GetKeyChainFromS3(paramsIndex int, keyResponse DownloadKeyResponse, logEnabled bool) (key.KeyChain, error) {

	log := logger.NewLogger(logEnabled)

	paramSet := bootstrapping.DefaultParametersSparse[paramsIndex]
	ckksParams := paramSet.SchemeParams

	Params, _ := ckks.NewParametersFromLiteral(ckksParams)

	// Load pk
	pkByte, err := DownloadFromS3(keyResponse.Pk)
	log.Log("Downloading PK")

	if err != nil {
		return key.KeyChain{}, err
	}

	pk := rlwe.PublicKey{}
	err = pk.UnmarshalBinary(pkByte)
	log.Log("Downloaded PK")

	if err != nil {
		return key.KeyChain{}, err
	}

	// Load rlk
	rlkByte, err := DownloadFromS3(keyResponse.Rlk)
	log.Log("Downloading RLK")

	if err != nil {
		return key.KeyChain{}, err
	}

	rlk := rlwe.RelinearizationKey{}
	err = rlk.UnmarshalBinary(rlkByte)
	log.Log("Downloaded RLK")

	if err != nil {
		return key.KeyChain{}, err
	}

	// Load rot keys
	galEls := make([]uint64, len(keyResponse.Rotk))
	keys := make(map[uint64]*rlwe.SwitchingKey)

	for i := range galEls{

		seperatedUrls := strings.Split(keyResponse.Rotk[i], "/")
		keyName := strings.Split(strings.Split(seperatedUrls[len(seperatedUrls) - 1], "?")[0], "_")
		galEl, err := strconv.Atoi(keyName[len(keyName) - 1])
		if err != nil {
			fmt.Println(keyName)
			return key.KeyChain{}, err
		}

		galEls[i] = uint64(galEl)

		// Download key
		log.Log(fmt.Sprintf("Downloading ROTK GalEl: %d (%d/%d)", galEl, i+1, len(galEls)))
		rotkByte, err := DownloadFromS3(keyResponse.Rotk[i])
		if err != nil {
			return key.KeyChain{}, err
		}

		keys[galEls[i]] = &rlwe.SwitchingKey{}
		err = keys[galEls[i]].UnmarshalBinary(rotkByte)
		if err != nil {
			return key.KeyChain{}, err
		}
		log.Log(fmt.Sprintf("Downloaded ROTK GalEl: %d (%d/%d)", galEl, i+1, len(galEls)))

	}

	rotk := rlwe.NewRotationKeySet(Params.Parameters, galEls)
	rotk.Keys = keys

	return key.KeyChain{
		ParamsIndex: paramsIndex,
		SecretKey: nil, 
		PublicKey: &pk,
		RelinKey: &rlk,
		GaloisKey: rotk,
		BtspGalKey: rotk,
	}, nil

}