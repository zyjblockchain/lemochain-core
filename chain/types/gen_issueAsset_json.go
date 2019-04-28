// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
)

var _ = (*issueAssetMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (i IssueAsset) MarshalJSON() ([]byte, error) {
	type IssueAsset struct {
		AssetCode common.Hash    `json:"assetCode" gencodec:"required"`
		MetaData  string         `json:"metaData" `
		Amount    *hexutil.Big10 `json:"supplyAmount" gencodec:"required"`
	}
	var enc IssueAsset
	enc.AssetCode = i.AssetCode
	enc.MetaData = i.MetaData
	enc.Amount = (*hexutil.Big10)(i.Amount)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (i *IssueAsset) UnmarshalJSON(input []byte) error {
	type IssueAsset struct {
		AssetCode *common.Hash   `json:"assetCode" gencodec:"required"`
		MetaData  *string        `json:"metaData" `
		Amount    *hexutil.Big10 `json:"supplyAmount" gencodec:"required"`
	}
	var dec IssueAsset
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.AssetCode == nil {
		return errors.New("missing required field 'assetCode' for IssueAsset")
	}
	i.AssetCode = *dec.AssetCode
	if dec.MetaData != nil {
		i.MetaData = *dec.MetaData
	}
	if dec.Amount == nil {
		return errors.New("missing required field 'supplyAmount' for IssueAsset")
	}
	i.Amount = (*big.Int)(dec.Amount)
	return nil
}
