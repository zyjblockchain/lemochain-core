// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package deputynode

import (
	"encoding/json"

	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
)

var _ = (*deputyNodesRecordMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (d DeputyNodesRecord) MarshalJSON() ([]byte, error) {
	type DeputyNodesRecord struct {
		TermStartHeight hexutil.Uint32 `json:"height"`
		Nodes           DeputyNodes    `json:"nodes"`
	}
	var enc DeputyNodesRecord
	enc.TermStartHeight = hexutil.Uint32(d.TermStartHeight)
	enc.Nodes = d.Nodes
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (d *DeputyNodesRecord) UnmarshalJSON(input []byte) error {
	type DeputyNodesRecord struct {
		TermStartHeight *hexutil.Uint32 `json:"height"`
		Nodes           *DeputyNodes    `json:"nodes"`
	}
	var dec DeputyNodesRecord
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.TermStartHeight != nil {
		d.TermStartHeight = uint32(*dec.TermStartHeight)
	}
	if dec.Nodes != nil {
		d.Nodes = *dec.Nodes
	}
	return nil
}
