package p2p

import (
	"errors"
)

var (
	ErrConnectSelf        = errors.New("can't connect yourself")
	ErrGenesisNotMatch    = errors.New("can't match genesis block")
	ErrBadRemoteID        = errors.New("bad remoteID")
	ErrNilRemoteID        = errors.New("remoteID can't be nil")
	ErrUnavailablePackage = errors.New("unavailable net package")
	ErrBadPubKey          = errors.New("invalid public key")
	ErrRecoveryFailed     = errors.New("recovery public key failed")
	ErrAlreadyRunning     = errors.New("has already running")
	ErrNilPrvKey          = errors.New("privateKey can't be nil")

	ErrRlpDecode = errors.New("rlp decode failed")

	ErrSrvHasStopped = errors.New("server has stopped")
)
