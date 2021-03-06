// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package chain

import (
	"encoding/json"
	"errors"

	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
)

var _ = (*candidateInfoMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (c CandidateInfo) MarshalJSON() ([]byte, error) {
	type CandidateInfo struct {
		MinerAddress  common.Address `json:"minerAddress" gencodec:"required"`
		IncomeAddress common.Address `json:"incomeAddress" gencodec:"required"`
		NodeID        hexutil.Bytes  `json:"nodeID" gencodec:"required"`
		Host          string         `json:"host" gencodec:"required"`
		Port          string         `json:"port" gencodec:"required"`
		Introduction  string         `json:"introduction"`
	}
	var enc CandidateInfo
	enc.MinerAddress = c.MinerAddress
	enc.IncomeAddress = c.IncomeAddress
	enc.NodeID = c.NodeID
	enc.Host = c.Host
	enc.Port = c.Port
	enc.Introduction = c.Introduction
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (c *CandidateInfo) UnmarshalJSON(input []byte) error {
	type CandidateInfo struct {
		MinerAddress  *common.Address `json:"minerAddress" gencodec:"required"`
		IncomeAddress *common.Address `json:"incomeAddress" gencodec:"required"`
		NodeID        *hexutil.Bytes  `json:"nodeID" gencodec:"required"`
		Host          *string         `json:"host" gencodec:"required"`
		Port          *string         `json:"port" gencodec:"required"`
		Introduction  *string         `json:"introduction"`
	}
	var dec CandidateInfo
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.MinerAddress == nil {
		return errors.New("missing required field 'minerAddress' for CandidateInfo")
	}
	c.MinerAddress = *dec.MinerAddress
	if dec.IncomeAddress == nil {
		return errors.New("missing required field 'incomeAddress' for CandidateInfo")
	}
	c.IncomeAddress = *dec.IncomeAddress
	if dec.NodeID == nil {
		return errors.New("missing required field 'nodeID' for CandidateInfo")
	}
	c.NodeID = *dec.NodeID
	if dec.Host == nil {
		return errors.New("missing required field 'host' for CandidateInfo")
	}
	c.Host = *dec.Host
	if dec.Port == nil {
		return errors.New("missing required field 'port' for CandidateInfo")
	}
	c.Port = *dec.Port
	if dec.Introduction != nil {
		c.Introduction = *dec.Introduction
	}
	return nil
}
