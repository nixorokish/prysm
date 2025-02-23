//go:build go1.18
// +build go1.18

package powchain_test

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/beacon"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/powchain"
	pb "github.com/prysmaticlabs/prysm/proto/engine/v1"
	"github.com/prysmaticlabs/prysm/testing/assert"
)

func FuzzForkChoiceResponse(f *testing.F) {
	valHash := common.Hash([32]byte{0xFF, 0x01})
	payloadID := beacon.PayloadID([8]byte{0x01, 0xFF, 0xAA, 0x00, 0xEE, 0xFE, 0x00, 0x00})
	valErr := "asjajshjahsaj"
	seed := &beacon.ForkChoiceResponse{
		PayloadStatus: beacon.PayloadStatusV1{
			Status:          "INVALID_TERMINAL_BLOCK",
			LatestValidHash: &valHash,
			ValidationError: &valErr,
		},
		PayloadID: &payloadID,
	}
	output, err := json.Marshal(seed)
	assert.NoError(f, err)
	f.Add(output)
	f.Fuzz(func(t *testing.T, jsonBlob []byte) {
		gethResp := &beacon.ForkChoiceResponse{}
		prysmResp := &powchain.ForkchoiceUpdatedResponse{}
		gethErr := json.Unmarshal(jsonBlob, gethResp)
		prysmErr := json.Unmarshal(jsonBlob, prysmResp)
		assert.Equal(t, gethErr != nil, prysmErr != nil, fmt.Sprintf("geth and prysm unmarshaller return inconsistent errors. %v and %v", gethErr, prysmErr))
		// Nothing to marshal if we have an error.
		if gethErr != nil {
			return
		}
		gethBlob, gethErr := json.Marshal(gethResp)
		prysmBlob, prysmErr := json.Marshal(prysmResp)
		assert.Equal(t, gethErr != nil, prysmErr != nil, "geth and prysm unmarshaller return inconsistent errors")
		newGethResp := &beacon.ForkChoiceResponse{}
		newGethErr := json.Unmarshal(prysmBlob, newGethResp)
		assert.NoError(t, newGethErr)
		if newGethResp.PayloadStatus.Status == "UNKNOWN" {
			return
		}

		newGethResp2 := &beacon.ForkChoiceResponse{}
		newGethErr = json.Unmarshal(gethBlob, newGethResp2)
		assert.NoError(t, newGethErr)

		assert.DeepEqual(t, newGethResp.PayloadID, newGethResp2.PayloadID)
		assert.DeepEqual(t, newGethResp.PayloadStatus.Status, newGethResp2.PayloadStatus.Status)
		assert.DeepEqual(t, newGethResp.PayloadStatus.LatestValidHash, newGethResp2.PayloadStatus.LatestValidHash)
		isNilOrEmpty := newGethResp.PayloadStatus.ValidationError == nil || (*newGethResp.PayloadStatus.ValidationError == "")
		isNilOrEmpty2 := newGethResp2.PayloadStatus.ValidationError == nil || (*newGethResp2.PayloadStatus.ValidationError == "")
		assert.DeepEqual(t, isNilOrEmpty, isNilOrEmpty2)
		if !isNilOrEmpty {
			assert.DeepEqual(t, *newGethResp.PayloadStatus.ValidationError, *newGethResp2.PayloadStatus.ValidationError)
		}
	})
}

func FuzzExchangeTransitionConfiguration(f *testing.F) {
	valHash := common.Hash([32]byte{0xFF, 0x01})
	ttd := hexutil.Big(*big.NewInt(math.MaxInt))
	seed := &beacon.TransitionConfigurationV1{
		TerminalTotalDifficulty: &ttd,
		TerminalBlockHash:       valHash,
		TerminalBlockNumber:     hexutil.Uint64(math.MaxUint64),
	}

	output, err := json.Marshal(seed)
	assert.NoError(f, err)
	f.Add(output)
	f.Fuzz(func(t *testing.T, jsonBlob []byte) {
		gethResp := &beacon.TransitionConfigurationV1{}
		prysmResp := &pb.TransitionConfiguration{}
		gethErr := json.Unmarshal(jsonBlob, gethResp)
		prysmErr := json.Unmarshal(jsonBlob, prysmResp)
		assert.Equal(t, gethErr != nil, prysmErr != nil, fmt.Sprintf("geth and prysm unmarshaller return inconsistent errors. %v and %v", gethErr, prysmErr))
		// Nothing to marshal if we have an error.
		if gethErr != nil {
			return
		}
		gethBlob, gethErr := json.Marshal(gethResp)
		prysmBlob, prysmErr := json.Marshal(prysmResp)
		if gethErr != nil {
			t.Errorf("%s %s", gethResp.TerminalTotalDifficulty.String(), prysmResp.TerminalTotalDifficulty)
		}
		assert.Equal(t, gethErr != nil, prysmErr != nil, fmt.Sprintf("geth and prysm unmarshaller return inconsistent errors. %v and %v", gethErr, prysmErr))
		if gethErr != nil {
			t.Errorf("%s %s", gethResp.TerminalTotalDifficulty.String(), prysmResp.TerminalTotalDifficulty)
		}
		newGethResp := &beacon.TransitionConfigurationV1{}
		newGethErr := json.Unmarshal(prysmBlob, newGethResp)
		assert.NoError(t, newGethErr)

		newGethResp2 := &beacon.TransitionConfigurationV1{}
		newGethErr = json.Unmarshal(gethBlob, newGethResp2)
		assert.NoError(t, newGethErr)
	})
}

func FuzzExecutionPayload(f *testing.F) {
	logsBloom := [256]byte{'j', 'u', 'n', 'k'}
	execData := &beacon.ExecutableDataV1{
		ParentHash:    common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
		FeeRecipient:  common.Address([20]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF}),
		StateRoot:     common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
		ReceiptsRoot:  common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
		LogsBloom:     logsBloom[:],
		Random:        common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
		Number:        math.MaxUint64,
		GasLimit:      math.MaxUint64,
		GasUsed:       math.MaxUint64,
		Timestamp:     100,
		ExtraData:     nil,
		BaseFeePerGas: big.NewInt(math.MaxInt),
		BlockHash:     common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
		Transactions:  [][]byte{{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}, {0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}, {0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}, {0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}},
	}
	output, err := json.Marshal(execData)
	assert.NoError(f, err)
	f.Add(output)
	f.Fuzz(func(t *testing.T, jsonBlob []byte) {
		gethResp := &beacon.ExecutableDataV1{}
		prysmResp := &pb.ExecutionPayload{}
		gethErr := json.Unmarshal(jsonBlob, gethResp)
		prysmErr := json.Unmarshal(jsonBlob, prysmResp)
		assert.Equal(t, gethErr != nil, prysmErr != nil, fmt.Sprintf("geth and prysm unmarshaller return inconsistent errors. %v and %v", gethErr, prysmErr))
		// Nothing to marshal if we have an error.
		if gethErr != nil {
			return
		}
		gethBlob, gethErr := json.Marshal(gethResp)
		prysmBlob, prysmErr := json.Marshal(prysmResp)
		assert.Equal(t, gethErr != nil, prysmErr != nil, "geth and prysm unmarshaller return inconsistent errors")
		newGethResp := &beacon.ExecutableDataV1{}
		newGethErr := json.Unmarshal(prysmBlob, newGethResp)
		assert.NoError(t, newGethErr)
		newGethResp2 := &beacon.ExecutableDataV1{}
		newGethErr = json.Unmarshal(gethBlob, newGethResp2)
		assert.NoError(t, newGethErr)

		assert.DeepEqual(t, newGethResp, newGethResp2)
	})
}

func FuzzExecutionBlock(f *testing.F) {
	logsBloom := [256]byte{'j', 'u', 'n', 'k'}
	addr := common.HexToAddress("095e7baea6a6c7c4c2dfeb977efac326af552d87")
	innerData := &types.DynamicFeeTx{
		ChainID:   big.NewInt(math.MaxInt),
		Nonce:     math.MaxUint64,
		GasTipCap: big.NewInt(math.MaxInt),
		GasFeeCap: big.NewInt(math.MaxInt),
		Gas:       math.MaxUint64,
		To:        &addr,
		Value:     big.NewInt(math.MaxInt),
		Data:      []byte{'r', 'a', 'n', 'd', 'o', 'm'},

		// Signature values
		V: big.NewInt(0),
		R: big.NewInt(math.MaxInt),
		S: big.NewInt(math.MaxInt),
	}
	tx := types.NewTx(innerData)
	execBlock := &pb.ExecutionBlock{
		Header: types.Header{
			ParentHash:  common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			Root:        common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			ReceiptHash: common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
			Bloom:       types.Bloom(logsBloom),
			Number:      big.NewInt(math.MaxInt),
			GasLimit:    math.MaxUint64,
			GasUsed:     math.MaxUint64,
			Time:        100,
			Extra:       nil,
			BaseFee:     big.NewInt(math.MaxInt),
			Difficulty:  big.NewInt(math.MaxInt),
		},
		Hash:            common.Hash([32]byte{0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01, 0xFF, 0x01}),
		TotalDifficulty: "999999999999999999999999999999999999999",
		Transactions:    []*types.Transaction{tx, tx},
	}
	output, err := json.Marshal(execBlock)
	assert.NoError(f, err)

	f.Add(output)

	f.Fuzz(func(t *testing.T, jsonBlob []byte) {
		gethResp := make(map[string]interface{})
		prysmResp := &pb.ExecutionBlock{}
		gethErr := json.Unmarshal(jsonBlob, &gethResp)
		prysmErr := json.Unmarshal(jsonBlob, prysmResp)
		// Nothing to marshal if we have an error.
		if gethErr != nil || prysmErr != nil {
			return
		}

		assert.NoError(t, validateBlockConsistency(prysmResp, gethResp))

		gethBlob, gethErr := json.Marshal(gethResp)
		prysmBlob, prysmErr := json.Marshal(prysmResp)
		assert.Equal(t, gethErr != nil, prysmErr != nil, "geth and prysm unmarshaller return inconsistent errors")
		newGethResp := make(map[string]interface{})
		newGethErr := json.Unmarshal(prysmBlob, &newGethResp)
		assert.NoError(t, newGethErr)
		newGethResp2 := make(map[string]interface{})
		newGethErr = json.Unmarshal(gethBlob, &newGethResp2)
		assert.NoError(t, newGethErr)

		assert.DeepEqual(t, newGethResp, newGethResp2)
		compareHeaders(t, jsonBlob)
	})
}

func compareHeaders(t *testing.T, jsonBlob []byte) {
	gethResp := &types.Header{}
	prysmResp := &pb.ExecutionBlock{}
	gethErr := json.Unmarshal(jsonBlob, gethResp)
	prysmErr := json.Unmarshal(jsonBlob, prysmResp)
	assert.Equal(t, gethErr != nil, prysmErr != nil, fmt.Sprintf("geth and prysm unmarshaller return inconsistent errors. %v and %v", gethErr, prysmErr))
	// Nothing to marshal if we have an error.
	if gethErr != nil {
		return
	}

	gethBlob, gethErr := json.Marshal(gethResp)
	prysmBlob, prysmErr := json.Marshal(prysmResp.Header)
	assert.Equal(t, gethErr != nil, prysmErr != nil, "geth and prysm unmarshaller return inconsistent errors")
	newGethResp := &types.Header{}
	newGethErr := json.Unmarshal(prysmBlob, newGethResp)
	assert.NoError(t, newGethErr)
	newGethResp2 := &types.Header{}
	newGethErr = json.Unmarshal(gethBlob, newGethResp2)
	assert.NoError(t, newGethErr)

	assert.DeepEqual(t, newGethResp, newGethResp2)
}

func validateBlockConsistency(execBlock *pb.ExecutionBlock, jsonMap map[string]interface{}) error {
	blockVal := reflect.ValueOf(execBlock).Elem()
	bType := reflect.TypeOf(execBlock).Elem()

	fieldnum := bType.NumField()

	for i := 0; i < fieldnum; i++ {
		field := bType.Field(i)
		fName := field.Tag.Get("json")
		if field.Name == "Header" {
			continue
		}
		if fName == "" {
			return errors.Errorf("Field %s had no json tag", field.Name)
		}
		fVal, ok := jsonMap[fName]
		if !ok {
			return errors.Errorf("%s doesn't exist in json map for field %s", fName, field.Name)
		}
		jsonVal := fVal
		bVal := blockVal.Field(i).Interface()
		if field.Name == "Hash" {
			jsonVal = common.HexToHash(jsonVal.(string))
		}
		if field.Name == "Transactions" {
			continue
		}
		if !reflect.DeepEqual(jsonVal, bVal) {
			return errors.Errorf("fields dont match, %v and %v are not equal for field %s", jsonVal, bVal, field.Name)
		}
	}
	return nil
}
