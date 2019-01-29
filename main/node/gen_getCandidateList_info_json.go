// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package node

import (
	"encoding/json"
	"errors"

	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
)

var _ = (*getCandidateListMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (g GetCandidateList) MarshalJSON() ([]byte, error) {
	type GetCandidateList struct {
		CandidateInfoes []*CandidateInfo `json:"candidateInfoes" gencodec:"required"`
		Total           hexutil.Uint32   `json:"total" gencodec:"required"`
	}
	var enc GetCandidateList
	enc.CandidateInfoes = g.CandidateInfoes
	enc.Total = hexutil.Uint32(g.Total)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (g *GetCandidateList) UnmarshalJSON(input []byte) error {
	type GetCandidateList struct {
		CandidateInfoes []*CandidateInfo `json:"candidateInfoes" gencodec:"required"`
		Total           *hexutil.Uint32  `json:"total" gencodec:"required"`
	}
	var dec GetCandidateList
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.CandidateInfoes == nil {
		return errors.New("missing required field 'candidateInfoes' for GetCandidateList")
	}
	g.CandidateInfoes = dec.CandidateInfoes
	if dec.Total == nil {
		return errors.New("missing required field 'total' for GetCandidateList")
	}
	g.Total = uint32(*dec.Total)
	return nil
}