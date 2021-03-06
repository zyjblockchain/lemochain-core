package types

import (
	"errors"
)

var (
	// ErrKnownBlock is returned when a block to import is already known locally.
	ErrKnownBlock = errors.New("block already known")

	// ErrGasLimitReached is returned by the gas pool if the amount of gas required
	// by a transaction is higher than what's left in the block.
	ErrGasLimitReached = errors.New("block gas limit reached")

	// ErrBlacklistedHash is returned if a block to import is on the blacklist.
	ErrBlacklistedHash = errors.New("blacklisted hash")
	ErrInvalidSig      = errors.New("invalid transaction sig")
	ErrInvalidVersion  = errors.New("invalid transaction version")
	ErrToNameLength    = errors.New("the length of toName field in transaction is out of max length limit")
	ErrToNameCharacter = errors.New("toName field in transaction contains illegal characters")
	ErrTxMessage       = errors.New("the length of message field in transaction is out of max length limit")
	ErrCreateContract  = errors.New("the data of create contract transaction can't be null")
	ErrSpecialTx       = errors.New("the data of special transaction can't be null")
	ErrTxType          = errors.New("the transaction type does not exit")
	ErrGasPrice        = errors.New("the transaction gas price is to low")
	ErrTxExpired       = errors.New("received transaction expiration time less than current time")
	ErrTxExpiration    = errors.New("received transaction expiration time can't more than 30 minutes")
	ErrNegativeValue   = errors.New("transaction amount can't be negative")
	ErrTxChainID       = errors.New("transaction chainID is incorrect")
	ErrBoxTx           = errors.New("box tx expiration time error")
	ErrVerifyBoxTx     = errors.New("box transaction cannot be a sub transaction")
	ErrToExist         = errors.New("verifyTx: the to of transaction is incorrect")
)
