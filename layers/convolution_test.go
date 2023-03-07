package layers

import (
	"fmt"
	"math"
	"testing"

	"github.com/tuneinsight/lattigo/v4/rlwe"
	"github.com/perm-ai/go-cerebrum/key"
	"github.com/perm-ai/go-cerebrum/logger"
	"github.com/perm-ai/go-cerebrum/utility"
)

func TestConv2dForward(t *testing.T) {

	keychain := key.GenerateKeys(0, false, true)
	utils := utility.NewUtils(keychain, math.Pow(2, 35), 0, true)

	testInput := [][][]*rlwe.Ciphertext{
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

	testKernel := [][][]*rlwe.Ciphertext{
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

	timer := logger.StartTimer("Conv forward")
	out := convLayer.Forward(testInput).Output
	timer.LogTimeTaken()

	for r := range out{
		one := (utils.Decrypt(out[r][0][0])[0])
		two := (utils.Decrypt(out[r][1][0])[0])
		three := (utils.Decrypt(out[r][2][0])[0])
		fmt.Printf("%f %f %f\n", one, two, three)
	}

}