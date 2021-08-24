package layers

import (
	"fmt"
	"math"
	"testing"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/utility"
)

var keychain = key.GenerateKeys(2, false, true)
var utils = utility.NewUtils(keychain, math.Pow(2, 35), 0, true)

func TestConv2dForward(t *testing.T) {

	testInput := [][][]*ckks.Ciphertext{
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
			},
		},
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
		},
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
		},
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
		},
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(2)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
		},
	}

	testKernel := [][][]*ckks.Ciphertext{
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
		},
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
			},
		},
		{
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
				utils.EncryptToPointer(utils.GenerateFilledArray(1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
			{
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(-1)),
				utils.EncryptToPointer(utils.GenerateFilledArray(0)),
			},
		},
	}

	kernel := generate2dKernelFromArray(testKernel)

	convLayer := NewConv2D(utils, 1, []int{3,3}, []int{2,2}, true, nil, false, []int{5,5,3}, int(math.Pow(2, 15)))
	convLayer.LoadKernels([]conv2dKernel{kernel})

	out, _ := convLayer.Forward(testInput)

	for r := range out{
		one := (utils.Decrypt(out[r][0][0])[0])
		two := (utils.Decrypt(out[r][1][0])[0])
		three := (utils.Decrypt(out[r][2][0])[0])
		fmt.Printf("%f %f %f\n", one, two, three)
	}

}