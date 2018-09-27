// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package types

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
)

var _ = (*accountDataMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (a AccountData) MarshalJSON() ([]byte, error) {
	type AccountData struct {
		Address        common.Address `json:"address" gencodec:"required"`
		Balance        *hexutil.Big   `json:"balance" gencodec:"required"`
		Version        hexutil.Uint64 `json:"version" gencodec:"required"`
		CodeHash       common.Hash    `json:"codeHash" gencodec:"required"`
		StorageRoot    common.Hash    `json:"root" gencodec:"required"`
		VersionRecords []VersionRecord
	}
	var enc AccountData
	enc.Address = a.Address
	enc.Balance = (*hexutil.Big)(a.Balance)
	enc.Version = hexutil.Uint64(a.Version)
	enc.CodeHash = a.CodeHash
	enc.StorageRoot = a.StorageRoot
	enc.VersionRecords = a.VersionRecords
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (a *AccountData) UnmarshalJSON(input []byte) error {
	type AccountData struct {
		Address        *common.Address `json:"address" gencodec:"required"`
		Balance        *hexutil.Big    `json:"balance" gencodec:"required"`
		Version        *hexutil.Uint64 `json:"version" gencodec:"required"`
		CodeHash       *common.Hash    `json:"codeHash" gencodec:"required"`
		StorageRoot    *common.Hash    `json:"root" gencodec:"required"`
		VersionRecords []VersionRecord
	}
	var dec AccountData
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Address == nil {
		return errors.New("missing required field 'address' for AccountData")
	}
	a.Address = *dec.Address
	if dec.Balance == nil {
		return errors.New("missing required field 'balance' for AccountData")
	}
	a.Balance = (*big.Int)(dec.Balance)
	if dec.Version == nil {
		return errors.New("missing required field 'version' for AccountData")
	}
	a.Version = uint32(*dec.Version)
	if dec.CodeHash == nil {
		return errors.New("missing required field 'codeHash' for AccountData")
	}
	a.CodeHash = *dec.CodeHash
	if dec.StorageRoot == nil {
		return errors.New("missing required field 'root' for AccountData")
	}
	a.StorageRoot = *dec.StorageRoot
	if dec.VersionRecords != nil {
		a.VersionRecords = dec.VersionRecords
	}
	return nil
}
