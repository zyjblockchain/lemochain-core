package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LemoFoundationLtd/lemochain-core/common"
	"github.com/LemoFoundationLtd/lemochain-core/common/crypto/sha3"
	"github.com/LemoFoundationLtd/lemochain-core/common/hexutil"
	"github.com/LemoFoundationLtd/lemochain-core/common/log"
	"github.com/LemoFoundationLtd/lemochain-core/common/merkle"
	"github.com/LemoFoundationLtd/lemochain-core/common/rlp"
	"io"
	"math/big"
	"reflect"
	"strings"
)

var (
	ErrUnknownChangeLogType = errors.New("unknown change log type")
	// ErrWrongChangeLogVersion is returned by the ChangeLog Undo/Redo if account has an unexpected version
	ErrWrongChangeLogVersion = errors.New("the version of change log and account is not match")
	ErrAlreadyRedo           = errors.New("the change log's version is lower than account's. maybe it has redid")
	ErrWrongChangeLogData    = errors.New("change log data is incorrect")
)

// ChangeLogProcessor is used to access account, and the intermediate data generated by transactions.
// It is implemented by account.Manager
type ChangeLogProcessor interface {
	GetAccount(addr common.Address) AccountAccessor
}

type ChangeLogType uint32
type changeLogDecoder func(*rlp.Stream) (interface{}, error)
type changeLogDoFunc func(*ChangeLog, ChangeLogProcessor) error

type logConfig struct {
	TypeName      string
	NewValDecoder changeLogDecoder
	ExtraDecoder  changeLogDecoder
	Redo          changeLogDoFunc
	Undo          changeLogDoFunc
}

// logConfigs define how the log type map to action functions
var logConfigs = make(map[ChangeLogType]logConfig)

func RegisterChangeLog(logType ChangeLogType, TypeName string, newValDecoder, extraDecoder changeLogDecoder, redo, undo changeLogDoFunc) {
	logConfigs[logType] = logConfig{TypeName, newValDecoder, extraDecoder, redo, undo}
}

func (t ChangeLogType) String() string {
	config, ok := logConfigs[t]
	if ok {
		return config.TypeName
	}
	return fmt.Sprintf("ChangeLogType(%d)", t)
}

type ChangeLog struct {
	LogType ChangeLogType  `json:"type"       gencodec:"required"`
	Address common.Address `json:"address"    gencodec:"required"`
	// The No. of ChangeLog in an account
	Version uint32 `json:"version"    gencodec:"required"`

	// data pointer. Their content type depend on specific NewXXXLog function
	OldVal interface{} `json:"-"` // It's used for undo. So no need to save or send to others
	NewVal interface{} `json:"newValue"`
	Extra  interface{} `json:"extra"`
}

type rlpChangeLog struct {
	LogType ChangeLogType
	Address common.Address
	Version uint32
	NewVal  interface{}
	Extra   interface{}
}

// Hash returns the keccak256 hash of its RLP encoding.
func (c *ChangeLog) Hash() (h common.Hash) {
	hw := sha3.NewKeccak256()
	// this will call EncodeRLP
	if err := rlp.Encode(hw, c); err != nil {
		log.Error("hash changelog fail", "err", err)
	}
	hw.Sum(h[:0])
	return h
}

// EncodeRLP implements rlp.Encoder.
func (c *ChangeLog) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, rlpChangeLog{
		LogType: c.LogType,
		Address: c.Address,
		Version: c.Version,
		NewVal:  c.NewVal,
		Extra:   c.Extra,
	})
}

// DecodeRLP implements rlp.Decoder.
func (c *ChangeLog) DecodeRLP(s *rlp.Stream) (err error) {
	if _, err = s.List(); err != nil {
		return err
	}
	if err = s.Decode(&c.LogType); err != nil {
		return err
	}
	if err = s.Decode(&c.Address); err != nil {
		return err
	}
	if err = s.Decode(&c.Version); err != nil {
		return err
	}

	// decode the interface{}
	config, ok := logConfigs[c.LogType]
	if !ok {
		log.Errorf("unexpected LogType %T", c.LogType)
		return ErrUnknownChangeLogType
	}
	if c.NewVal, err = config.NewValDecoder(s); err != nil {
		return err
	}
	if c.Extra, err = config.ExtraDecoder(s); err != nil {
		return err
	}
	// This error means there are some data need to be decoded
	err = s.ListEnd()
	return err
}

// MarshalJSON marshals as JSON.
func (c ChangeLog) MarshalJSON() ([]byte, error) {
	type jsonChangeLog struct {
		LogType hexutil.Uint32 `json:"type"       gencodec:"required"`
		Address common.Address `json:"address"    gencodec:"required"`
		Version hexutil.Uint32 `json:"version"    gencodec:"required"`
		NewVal  interface{}    `json:"newValue"`
		Extra   interface{}    `json:"extra"`
	}
	var enc jsonChangeLog
	enc.LogType = hexutil.Uint32(c.LogType)
	enc.Address = c.Address
	enc.Version = hexutil.Uint32(c.Version)
	if c.NewVal != nil {
		// big.Int结构体没有实现Marshaler接口，所以需要直接转字符串
		if val, ok := c.NewVal.(big.Int); ok {
			enc.NewVal = val.String()
		} else {
			enc.NewVal = c.NewVal
		}
	}
	if c.Extra != nil {
		if val, ok := c.Extra.(big.Int); ok {
			enc.Extra = val.String()
		} else {
			enc.Extra = c.Extra
		}
	}
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (c *ChangeLog) UnmarshalJSON(input []byte) error {
	type jsonChangeLog struct {
		LogType *hexutil.Uint32 `json:"type"       gencodec:"required"`
		Address *common.Address `json:"address"    gencodec:"required"`
		Version *hexutil.Uint32 `json:"version"    gencodec:"required"`
		NewVal  interface{}     `json:"newValue"`
		Extra   interface{}     `json:"extra"`
	}
	var dec jsonChangeLog
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.LogType == nil {
		return errors.New("missing required field 'type' for ChangeLog")
	}
	c.LogType = ChangeLogType(*dec.LogType)
	if dec.Address == nil {
		return errors.New("missing required field 'address' for ChangeLog")
	}
	c.Address = *dec.Address
	if dec.Version == nil {
		return errors.New("missing required field 'version' for ChangeLog")
	}
	c.Version = uint32(*dec.Version)
	_, ok := logConfigs[c.LogType]
	if !ok {
		log.Errorf("unexpected LogType %T", c.LogType)
		return ErrUnknownChangeLogType
	}
	if dec.NewVal != nil {
		c.NewVal = dec.NewVal
	}
	if dec.Extra != nil {
		c.Extra = dec.Extra
	}
	return nil
}

func (c *ChangeLog) Copy() *ChangeLog {
	cpy := *c
	return &cpy
}

var bigIntType = reflect.TypeOf(big.Int{})
var addressType = reflect.TypeOf(common.Address{})
var signersType = reflect.TypeOf(Signers{})

func formatInterface(v interface{}) interface{} {
	result := v
	if reflect.TypeOf(result) == bigIntType {
		i := v.(big.Int)
		result = (&i).Text(10)
	} else if reflect.TypeOf(result) == addressType {
		i := v.(common.Address)
		result = i.String()
	} else if reflect.TypeOf(result) == signersType {
		i := v.(Signers)
		result = i.String()
	}
	return result
}

func (c *ChangeLog) String() string {
	set := []string{
		fmt.Sprintf("Account: %s", c.Address.String()),
		fmt.Sprintf("Version: %d", c.Version),
	}

	if !common.IsNil(c.OldVal) {
		set = append(set, fmt.Sprintf("OldVal: %v", formatInterface(c.OldVal)))
	}
	if !common.IsNil(c.NewVal) {
		set = append(set, fmt.Sprintf("NewVal: %v", formatInterface(c.NewVal)))
	}
	if !common.IsNil(c.Extra) {
		set = append(set, fmt.Sprintf("Extra: %v", formatInterface(c.Extra)))
	}

	return fmt.Sprintf("%s{%s}", c.LogType, strings.Join(set, ", "))
}

type ChangeLogSlice []*ChangeLog

func (c ChangeLogSlice) Len() int {
	return len(c)
}

func (c ChangeLogSlice) Less(i, j int) bool {
	return c[i].Version < c[j].Version
}

func (c ChangeLogSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c ChangeLogSlice) Search(version uint32) int {
	for i, value := range c {
		if value.Version == version {
			return i
		}
	}
	return -1
}

// FindByType find the first same type change log.
func (c ChangeLogSlice) FindByType(target *ChangeLog) *ChangeLog {
	for _, item := range c {
		if item.LogType == target.LogType {
			return item
		}
	}
	return nil
}

// MerkleRootSha compute the root hash of ChangeLog merkle trie
func (c ChangeLogSlice) MerkleRootSha() common.Hash {
	leaves := make([]common.Hash, len(c))
	for i, item := range c {
		leaves[i] = item.Hash()
	}
	return merkle.New(leaves).Root()
}

// Undo reverts the change. Its behavior depends on ChangeLog.ChangeLogType
func (c *ChangeLog) Undo(processor ChangeLogProcessor) error {
	config, ok := logConfigs[c.LogType]
	if !ok {
		log.Errorf("unexpected LogType %T", c.LogType)
		return ErrUnknownChangeLogType
	}

	if err := config.Undo(c, processor); err != nil {
		return err
	}
	return nil
}

// Redo reply the change for light client. Its behavior depends on ChangeLog.ChangeLogType
func (c *ChangeLog) Redo(processor ChangeLogProcessor) error {
	config, ok := logConfigs[c.LogType]
	if !ok {
		log.Errorf("unexpected LogType %T", c.LogType)
		return ErrUnknownChangeLogType
	}

	if err := config.Redo(c, processor); err != nil {
		return err
	}
	return nil
}
