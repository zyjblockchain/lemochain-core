// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package deputynode

import (
	"encoding/json"

	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
)

var _ = (*termRecordMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (t TermRecord) MarshalJSON() ([]byte, error) {
	type TermRecord struct {
		StartHeight hexutil.Uint32 `json:"height"`
		Nodes       DeputyNodes    `json:"nodes"`
	}
	var enc TermRecord
	enc.StartHeight = hexutil.Uint32(t.StartHeight)
	enc.Nodes = t.Nodes
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (t *TermRecord) UnmarshalJSON(input []byte) error {
	type TermRecord struct {
		StartHeight *hexutil.Uint32 `json:"height"`
		Nodes       *DeputyNodes    `json:"nodes"`
	}
	var dec TermRecord
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.StartHeight != nil {
		t.StartHeight = uint32(*dec.StartHeight)
	}
	if dec.Nodes != nil {
		t.Nodes = *dec.Nodes
	}
	return nil
}
