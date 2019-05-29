// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
)

var _ = (*txdataMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (t txdata) MarshalJSON() ([]byte, error) {
	type txdata struct {
		Type          hexutil.Uint16  `json:"type" gencodec:"required"`
		Version       hexutil.Uint8   `json:"version" gencodec:"required"`
		ChainID       hexutil.Uint16  `json:"chainID" gencodec:"required"`
		From          common.Address  `json:"from" gencodec:"required"`
		GasPayer      *common.Address `json:"gasPayer" rlp:"nil"`
		Recipient     *common.Address `json:"to" rlp:"nil"`
		RecipientName string          `json:"toName"`
		GasPrice      *hexutil.Big10  `json:"gasPrice" gencodec:"required"`
		GasLimit      hexutil.Uint64  `json:"gasLimit" gencodec:"required"`
		Amount        *hexutil.Big10  `json:"amount" gencodec:"required"`
		Data          hexutil.Bytes   `json:"data"`
		Expiration    hexutil.Uint64  `json:"expirationTime" gencodec:"required"`
		Message       string          `json:"message"`
		Sigs          []hexutil.Bytes `json:"sigs" gencodec:"required"`
		Hash          *common.Hash    `json:"hash" rlp:"-"`
		GasPayerSigs  []hexutil.Bytes `json:"gasPayerSigs"`
	}
	var enc txdata
	enc.Type = hexutil.Uint16(t.Type)
	enc.Version = hexutil.Uint8(t.Version)
	enc.ChainID = hexutil.Uint16(t.ChainID)
	enc.From = t.From
	enc.GasPayer = t.GasPayer
	enc.Recipient = t.Recipient
	enc.RecipientName = t.RecipientName
	enc.GasPrice = (*hexutil.Big10)(t.GasPrice)
	enc.GasLimit = hexutil.Uint64(t.GasLimit)
	enc.Amount = (*hexutil.Big10)(t.Amount)
	enc.Data = t.Data
	enc.Expiration = hexutil.Uint64(t.Expiration)
	enc.Message = t.Message
	if t.Sigs != nil {
		enc.Sigs = make([]hexutil.Bytes, len(t.Sigs))
		for k, v := range t.Sigs {
			enc.Sigs[k] = v
		}
	}
	enc.Hash = t.Hash
	if t.GasPayerSigs != nil {
		enc.GasPayerSigs = make([]hexutil.Bytes, len(t.GasPayerSigs))
		for k, v := range t.GasPayerSigs {
			enc.GasPayerSigs[k] = v
		}
	}
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (t *txdata) UnmarshalJSON(input []byte) error {
	type txdata struct {
		Type          *hexutil.Uint16 `json:"type" gencodec:"required"`
		Version       *hexutil.Uint8  `json:"version" gencodec:"required"`
		ChainID       *hexutil.Uint16 `json:"chainID" gencodec:"required"`
		From          *common.Address `json:"from" gencodec:"required"`
		GasPayer      *common.Address `json:"gasPayer" rlp:"nil"`
		Recipient     *common.Address `json:"to" rlp:"nil"`
		RecipientName *string         `json:"toName"`
		GasPrice      *hexutil.Big10  `json:"gasPrice" gencodec:"required"`
		GasLimit      *hexutil.Uint64 `json:"gasLimit" gencodec:"required"`
		Amount        *hexutil.Big10  `json:"amount" gencodec:"required"`
		Data          *hexutil.Bytes  `json:"data"`
		Expiration    *hexutil.Uint64 `json:"expirationTime" gencodec:"required"`
		Message       *string         `json:"message"`
		Sigs          []hexutil.Bytes `json:"sigs" gencodec:"required"`
		Hash          *common.Hash    `json:"hash" rlp:"-"`
		GasPayerSigs  []hexutil.Bytes `json:"gasPayerSigs"`
	}
	var dec txdata
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Type == nil {
		return errors.New("missing required field 'type' for txdata")
	}
	t.Type = uint16(*dec.Type)
	if dec.Version == nil {
		return errors.New("missing required field 'version' for txdata")
	}
	t.Version = uint8(*dec.Version)
	if dec.ChainID == nil {
		return errors.New("missing required field 'chainID' for txdata")
	}
	t.ChainID = uint16(*dec.ChainID)
	if dec.From == nil {
		return errors.New("missing required field 'from' for txdata")
	}
	t.From = *dec.From
	if dec.GasPayer != nil {
		t.GasPayer = dec.GasPayer
	}
	if dec.Recipient != nil {
		t.Recipient = dec.Recipient
	}
	if dec.RecipientName != nil {
		t.RecipientName = *dec.RecipientName
	}
	if dec.GasPrice == nil {
		return errors.New("missing required field 'gasPrice' for txdata")
	}
	t.GasPrice = (*big.Int)(dec.GasPrice)
	if dec.GasLimit == nil {
		return errors.New("missing required field 'gasLimit' for txdata")
	}
	t.GasLimit = uint64(*dec.GasLimit)
	if dec.Amount == nil {
		return errors.New("missing required field 'amount' for txdata")
	}
	t.Amount = (*big.Int)(dec.Amount)
	if dec.Data != nil {
		t.Data = *dec.Data
	}
	if dec.Expiration == nil {
		return errors.New("missing required field 'expirationTime' for txdata")
	}
	t.Expiration = uint64(*dec.Expiration)
	if dec.Message != nil {
		t.Message = *dec.Message
	}
	if dec.Sigs == nil {
		return errors.New("missing required field 'sigs' for txdata")
	}
	t.Sigs = make([][]byte, len(dec.Sigs))
	for k, v := range dec.Sigs {
		t.Sigs[k] = v
	}
	if dec.Hash != nil {
		t.Hash = dec.Hash
	}
	if dec.GasPayerSigs != nil {
		t.GasPayerSigs = make([][]byte, len(dec.GasPayerSigs))
		for k, v := range dec.GasPayerSigs {
			t.GasPayerSigs[k] = v
		}
	}
	return nil
}
