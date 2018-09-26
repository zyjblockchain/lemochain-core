package account

import (
	"github.com/LemoFoundationLtd/lemochain-go/chain/types"
	"github.com/LemoFoundationLtd/lemochain-go/common"
	"github.com/LemoFoundationLtd/lemochain-go/store"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func loadAccount(address common.Address) *Account {
	db := newDB()
	data, _ := db.GetAccount(defaultBlock.hash, address)
	return NewAccount(db, address, data)
}

func TestAccount_GetAddress(t *testing.T) {
	db := newDB()

	// load default account
	account := loadAccount(defaultAccounts[0].Address)
	assert.Equal(t, uint32(100), account.GetVersion())

	// load not exist account
	account = loadAccount(common.HexToAddress("0xaaa"))
	assert.Equal(t, common.HexToAddress("0xaaa"), account.GetAddress())

	// load from genesis' parent block
	_, err := db.GetAccount(common.Hash{}, common.HexToAddress("0xaaa"))
	assert.Equal(t, store.ErrNotExist, err)
}

func TestAccount_SetBalance_GetBalance(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)
	assert.Equal(t, big.NewInt(100), account.GetBalance())

	account.SetBalance(big.NewInt(200))
	assert.Equal(t, big.NewInt(200), account.GetBalance())
}

func TestAccount_SetVersion_GetVersion(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)
	assert.Equal(t, uint32(100), account.GetVersion())

	account.SetVersion(200)
	assert.Equal(t, uint32(200), account.GetVersion())
}

func TestAccount_SetSuicide_GetSuicide(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)
	assert.Equal(t, false, account.GetSuicide())

	account.SetSuicide(true)
	assert.Equal(t, true, account.GetSuicide())
}

func TestAccount_SetCodeHash_GetCodeHash(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)
	assert.Equal(t, defaultCodes[0].hash, account.GetCodeHash())

	account.code = types.Code{0x12}
	account.SetCodeHash(c(2))
	assert.Equal(t, c(2), account.GetCodeHash())
	assert.Empty(t, account.code)

	// set to empty
	account.SetCodeHash(common.Hash{})
	assert.Equal(t, sha3Nil, account.GetCodeHash())
}

func TestAccount_SetCode_GetCode(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)
	readCode, err := account.GetCode()
	assert.NoError(t, err)
	assert.Equal(t, types.Code{12, 34}, readCode)

	account.SetCode(types.Code{0x12})
	readCode, err = account.GetCode()
	assert.NoError(t, err)
	assert.Equal(t, types.Code{0x12}, readCode)
	assert.Equal(t, common.HexToHash("0x5fa2358263196dbbf23d1ca7a509451f7a2f64c15837bfbb81298b1e3e24e4fa"), account.GetCodeHash())
	assert.Equal(t, true, account.codeIsDirty)

	// clear code
	account.codeIsDirty = false
	account.SetCode(nil)
	readCode, err = account.GetCode()
	assert.NoError(t, err)
	assert.Empty(t, readCode)
	assert.Equal(t, sha3Nil, account.GetCodeHash())
	assert.Equal(t, true, account.codeIsDirty)

	// set nil to new account
	account = loadAccount(common.HexToAddress("0xaaa"))
	account.SetCode(nil)
	readCode, err = account.GetCode()
	assert.NoError(t, err)
	assert.Empty(t, readCode)
	assert.Equal(t, sha3Nil, account.GetCodeHash())
	assert.Equal(t, false, account.codeIsDirty)
}

func TestAccount_SetStorageRoot_GetStorageRoot(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)
	assert.Equal(t, defaultAccounts[0].StorageRoot, account.GetStorageRoot())

	account.dirtyStorage[k(1)] = []byte{12}
	account.SetStorageRoot(h(200))
	assert.Equal(t, h(200), account.GetStorageRoot())
	assert.Empty(t, account.dirtyStorage)
}

func TestAccount_SetStorageState_GetStorageState(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)

	// exist in db
	readValue, err := account.GetStorageState(defaultStorage[0].key)
	assert.NoError(t, err)
	assert.Equal(t, defaultStorage[0].value, readValue)

	// exist in cache
	key1 := k(1)
	value1 := []byte{11}
	account.cachedStorage[key1] = value1
	readValue, err = account.GetStorageState(key1)
	assert.NoError(t, err)
	assert.Equal(t, value1, readValue)

	// not exist value
	readValue, err = account.GetStorageState(k(2))
	assert.NoError(t, err)
	assert.Empty(t, readValue) // []byte(nil)

	// set
	key3 := k(3)
	value3 := []byte{22}
	account.SetStorageState(key3, value3)
	assert.Equal(t, value3, account.cachedStorage[key3])
	assert.Equal(t, value3, account.dirtyStorage[key3])

	// set empty
	key4 := k(4)
	value4 := []byte{}
	account.SetStorageState(key4, value4)
	readValue, err = account.GetStorageState(key4)
	assert.NoError(t, err)
	assert.Equal(t, value4, readValue)
	// set nil
	account.SetStorageState(key4, nil)
	readValue, err = account.GetStorageState(key4)
	assert.NoError(t, err)
	assert.Empty(t, readValue) // []byte(nil)

	// set with empty key
	key5 := common.Hash{}
	value5 := []byte{55}
	account.SetStorageState(key5, value5)
	readValue, err = account.GetStorageState(key5)
	assert.NoError(t, err)
	assert.Equal(t, value5, readValue)

	// invalid root
	account.SetStorageRoot(h(1))
	readValue, err = account.GetStorageState(k(6))
	assert.Equal(t, ErrTrieFail, err)
	assert.Empty(t, readValue) // []byte(nil)
}

func TestAccount_IsEmpty(t *testing.T) {
	account := loadAccount(common.HexToAddress("0x1"))
	assert.Equal(t, true, account.IsEmpty())
	account.SetVersion(100)
	assert.Equal(t, false, account.IsEmpty())
}

func TestAccount_Finalise_Save(t *testing.T) {
	account := loadAccount(defaultAccounts[0].Address)

	// nothing to finalise
	value, err := account.GetStorageState(defaultStorage[0].key)
	assert.NoError(t, err)
	assert.Equal(t, defaultStorage[0].value, value)
	assert.Equal(t, 1, len(account.cachedStorage))
	assert.Equal(t, 0, len(account.dirtyStorage))
	err = account.Finalise(1)
	assert.NoError(t, err)
	assert.Equal(t, defaultAccounts[0].StorageRoot, account.GetStorageRoot())
	// save
	err = account.Save()
	assert.NoError(t, err)

	// finalise dirty storage
	key := k(1)
	value = []byte{11, 22, 33}
	blockHeight := uint32(1)
	err = account.SetStorageState(key, value)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(account.cachedStorage))
	assert.Equal(t, 1, len(account.dirtyStorage))
	assert.Equal(t, value, account.dirtyStorage[key])
	account.SetVersion(10)
	err = account.Finalise(blockHeight)
	assert.NoError(t, err)
	assert.Equal(t, "0xfb4fbcae2c19f15b34c53b059a4af53d8d793607bd8ca5868eeb9c817c4e5bc7", account.GetStorageRoot().Hex())
	assert.Equal(t, 1, len(account.data.VersionRecords))
	assert.Equal(t, blockHeight, account.data.VersionRecords[0].Height)
	assert.Equal(t, uint32(10), account.data.VersionRecords[0].Version)
	assert.Equal(t, 0, len(account.dirtyStorage))
	account.SetVersion(11)
	err = account.Finalise(blockHeight)
	assert.Equal(t, 1, len(account.data.VersionRecords))
	assert.Equal(t, blockHeight, account.data.VersionRecords[0].Height)
	assert.Equal(t, uint32(11), account.data.VersionRecords[0].Version)
	account.SetVersion(12)
	err = account.Finalise(blockHeight + 1)
	assert.Equal(t, 2, len(account.data.VersionRecords))
	assert.Equal(t, blockHeight+1, account.data.VersionRecords[1].Height)
	assert.Equal(t, uint32(12), account.data.VersionRecords[1].Version)
	// save
	err = account.Save()
	assert.NoError(t, err)
	account2 := loadAccount(defaultAccounts[0].Address)
	account2.SetStorageRoot(account.GetStorageRoot())
	readValue, err := account2.GetStorageState(key)
	assert.NoError(t, err)
	assert.Equal(t, value, readValue)

	// finalise after modify value
	value = []byte{44, 55}
	err = account.SetStorageState(key, value)
	assert.NoError(t, err)
	assert.Equal(t, value, account.dirtyStorage[key])
	err = account.Finalise(blockHeight)
	assert.NoError(t, err)
	assert.Equal(t, "0x0adade766035e43ef12b9ac1a84db5eae1c9a3501d81510cdc8cbd0fb3a4b922", account.GetStorageRoot().Hex())
	// save
	err = account.Save()
	assert.NoError(t, err)
	account2 = loadAccount(defaultAccounts[0].Address)
	account2.SetStorageRoot(account.GetStorageRoot())
	readValue, err = account2.GetStorageState(key)
	assert.NoError(t, err)
	assert.Equal(t, value, readValue)

	// finalise after remove value
	err = account.SetStorageState(key, nil)
	assert.NoError(t, err)
	assert.Empty(t, account.dirtyStorage[key])
	err = account.Finalise(blockHeight)
	assert.NoError(t, err)
	assert.Equal(t, defaultAccounts[0].StorageRoot, account.GetStorageRoot())
	// save
	err = account.Save()
	assert.NoError(t, err)
	account2 = loadAccount(defaultAccounts[0].Address)
	account2.SetStorageRoot(account.GetStorageRoot())
	readValue, err = account2.GetStorageState(key)
	assert.NoError(t, err)
	assert.Empty(t, readValue)

	// finalise after remove value with empty []byte
	value = []byte{}
	err = account.SetStorageState(key, value)
	assert.Equal(t, value, account.dirtyStorage[key])
	assert.NoError(t, err)
	err = account.Finalise(blockHeight)
	assert.Equal(t, defaultAccounts[0].StorageRoot, account.GetStorageRoot())
	assert.NoError(t, err)
	// save
	err = account.Save()
	assert.NoError(t, err)
	account2 = loadAccount(defaultAccounts[0].Address)
	account2.SetStorageRoot(account.GetStorageRoot())
	readValue, err = account2.GetStorageState(key)
	assert.NoError(t, err)
	assert.Empty(t, readValue)

	// dirty code
	account.SetCode(types.Code{0x12})
	err = account.Save()
	assert.NoError(t, err)
	assert.Equal(t, false, account.codeIsDirty)
	account2 = loadAccount(defaultAccounts[0].Address)
	account2.SetCodeHash(common.HexToHash("0x5fa2358263196dbbf23d1ca7a509451f7a2f64c15837bfbb81298b1e3e24e4fa"))
	readCode, err := account2.GetCode()
	assert.NoError(t, err)
	assert.Equal(t, types.Code{0x12}, readCode)

	// root changed after finalise
	key = k(2)
	value = []byte{44, 55}
	err = account.SetStorageState(key, value)
	assert.NoError(t, err)
	err = account.Finalise(blockHeight)
	assert.NoError(t, err)
	account.data.StorageRoot = defaultAccounts[0].StorageRoot
	err = account.Save()
	assert.Equal(t, ErrTrieChanged, err)

	// invalid root
	account.SetStorageRoot(h(1))
	value = []byte{11}
	err = account.SetStorageState(key, value)
	assert.NoError(t, err)
	err = account.Finalise(blockHeight)
	assert.Equal(t, ErrTrieFail, err)
	err = account.Save()
	assert.Equal(t, ErrTrieFail, err)
}

func TestAccount_LoadChangeLogs(t *testing.T) {
	// TODO
}
