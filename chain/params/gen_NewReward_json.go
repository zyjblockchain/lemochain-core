// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package params

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/LemoFoundationLtd/lemochain-go/common/hexutil"
)

var _ = (*NewRewardMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (n NewReward) MarshalJSON() ([]byte, error) {
	type NewReward struct {
		Term  hexutil.Uint32 `json:"term" gencodec:"required"`
		Value *hexutil.Big10 `json:"value" gencodec:"required"`
	}
	var enc NewReward
	enc.Term = hexutil.Uint32(n.Term)
	enc.Value = (*hexutil.Big10)(n.Value)
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (n *NewReward) UnmarshalJSON(input []byte) error {
	type NewReward struct {
		Term  *hexutil.Uint32 `json:"term" gencodec:"required"`
		Value *hexutil.Big10  `json:"value" gencodec:"required"`
	}
	var dec NewReward
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Term == nil {
		return errors.New("missing required field 'term' for NewReward")
	}
	n.Term = uint32(*dec.Term)
	if dec.Value == nil {
		return errors.New("missing required field 'value' for NewReward")
	}
	n.Value = (*big.Int)(dec.Value)
	return nil
}
