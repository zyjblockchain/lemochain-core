// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
)

var _ = (*tradingAssetMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (t TradingAsset) MarshalJSON() ([]byte, error) {
	type TradingAsset struct {
		AssetId common.Hash    `json:"assetId" gencodec:"required"`
		Value   *hexutil.Big10 `json:"transferAsset" gencodec:"required"`
		Input   []byte         `json:"input"`
	}
	var enc TradingAsset
	enc.AssetId = t.AssetId
	enc.Value = (*hexutil.Big10)(t.Value)
	enc.Input = t.Input
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (t *TradingAsset) UnmarshalJSON(input []byte) error {
	type TradingAsset struct {
		AssetId *common.Hash   `json:"assetId" gencodec:"required"`
		Value   *hexutil.Big10 `json:"transferAsset" gencodec:"required"`
		Input   []byte         `json:"input"`
	}
	var dec TradingAsset
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.AssetId == nil {
		return errors.New("missing required field 'assetId' for TradingAsset")
	}
	t.AssetId = *dec.AssetId
	if dec.Value == nil {
		return errors.New("missing required field 'transferAsset' for TradingAsset")
	}
	t.Value = (*big.Int)(dec.Value)
	if dec.Input != nil {
		t.Input = dec.Input
	}
	return nil
}
