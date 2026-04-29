// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package aegis

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// VerifierRegistryMetaData contains all meta data concerning the VerifierRegistry contract.
var VerifierRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_usdc\",\"type\":\"address\",\"internalType\":\"contractIUSDC\"},{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_MIN_STAKE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"WITHDRAW_COOLDOWN_BLOCKS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"addStake\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"aegis\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAccuracy\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"bps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"iNftContract\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractVerifierINFT\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isActive\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minStake\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"recordVote\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"wasCorrect\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"register\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"iNftId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"requestWithdraw\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAegis\",\"inputs\":[{\"name\":\"_aegis\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setINftContract\",\"inputs\":[{\"name\":\"_iNft\",\"type\":\"address\",\"internalType\":\"contractVerifierINFT\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinStake\",\"inputs\":[{\"name\":\"_minStake\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"slash\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"actualSlashed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"stakeOf\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"usdc\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIUSDC\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifierAddresses\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifierCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"verifiers\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"stake\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"votesTotal\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"votesCorrect\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"iNftId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"withdrawableAt\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"active\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdraw\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AegisSet\",\"inputs\":[{\"name\":\"aegis\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MinStakeUpdated\",\"inputs\":[{\"name\":\"oldMin\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newMin\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StakeAdded\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newStake\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"StakeWithdrawn\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VerifierRegistered\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"stake\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"iNftId\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VerifierSlashed\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"newStake\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VoteRecorded\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"wasCorrect\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"},{\"name\":\"votesTotal\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"votesCorrect\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"WithdrawalRequested\",\"inputs\":[{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"unlockBlock\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AddressInsufficientBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AlreadyRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CooldownNotElapsed\",\"inputs\":[{\"name\":\"unlockBlock\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"currentBlock\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"FailedInnerCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientStake\",\"inputs\":[{\"name\":\"provided\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"required\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NotAegis\",\"inputs\":[{\"name\":\"caller\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"NotINftController\",\"inputs\":[{\"name\":\"tokenId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"claimedBy\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"NotRegistered\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"WithdrawNotRequested\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroRecipient\",\"inputs\":[]}]",
}

// VerifierRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use VerifierRegistryMetaData.ABI instead.
var VerifierRegistryABI = VerifierRegistryMetaData.ABI

// VerifierRegistry is an auto generated Go binding around an Ethereum contract.
type VerifierRegistry struct {
	VerifierRegistryCaller     // Read-only binding to the contract
	VerifierRegistryTransactor // Write-only binding to the contract
	VerifierRegistryFilterer   // Log filterer for contract events
}

// VerifierRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type VerifierRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifierRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type VerifierRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifierRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type VerifierRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// VerifierRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type VerifierRegistrySession struct {
	Contract     *VerifierRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// VerifierRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type VerifierRegistryCallerSession struct {
	Contract *VerifierRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// VerifierRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type VerifierRegistryTransactorSession struct {
	Contract     *VerifierRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// VerifierRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type VerifierRegistryRaw struct {
	Contract *VerifierRegistry // Generic contract binding to access the raw methods on
}

// VerifierRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type VerifierRegistryCallerRaw struct {
	Contract *VerifierRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// VerifierRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type VerifierRegistryTransactorRaw struct {
	Contract *VerifierRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewVerifierRegistry creates a new instance of VerifierRegistry, bound to a specific deployed contract.
func NewVerifierRegistry(address common.Address, backend bind.ContractBackend) (*VerifierRegistry, error) {
	contract, err := bindVerifierRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistry{VerifierRegistryCaller: VerifierRegistryCaller{contract: contract}, VerifierRegistryTransactor: VerifierRegistryTransactor{contract: contract}, VerifierRegistryFilterer: VerifierRegistryFilterer{contract: contract}}, nil
}

// NewVerifierRegistryCaller creates a new read-only instance of VerifierRegistry, bound to a specific deployed contract.
func NewVerifierRegistryCaller(address common.Address, caller bind.ContractCaller) (*VerifierRegistryCaller, error) {
	contract, err := bindVerifierRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryCaller{contract: contract}, nil
}

// NewVerifierRegistryTransactor creates a new write-only instance of VerifierRegistry, bound to a specific deployed contract.
func NewVerifierRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*VerifierRegistryTransactor, error) {
	contract, err := bindVerifierRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryTransactor{contract: contract}, nil
}

// NewVerifierRegistryFilterer creates a new log filterer instance of VerifierRegistry, bound to a specific deployed contract.
func NewVerifierRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*VerifierRegistryFilterer, error) {
	contract, err := bindVerifierRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryFilterer{contract: contract}, nil
}

// bindVerifierRegistry binds a generic wrapper to an already deployed contract.
func bindVerifierRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := VerifierRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VerifierRegistry *VerifierRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VerifierRegistry.Contract.VerifierRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VerifierRegistry *VerifierRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.VerifierRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VerifierRegistry *VerifierRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.VerifierRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_VerifierRegistry *VerifierRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _VerifierRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_VerifierRegistry *VerifierRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_VerifierRegistry *VerifierRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTMINSTAKE is a free data retrieval call binding the contract method 0xdff3ece9.
//
// Solidity: function DEFAULT_MIN_STAKE() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCaller) DEFAULTMINSTAKE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "DEFAULT_MIN_STAKE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DEFAULTMINSTAKE is a free data retrieval call binding the contract method 0xdff3ece9.
//
// Solidity: function DEFAULT_MIN_STAKE() view returns(uint256)
func (_VerifierRegistry *VerifierRegistrySession) DEFAULTMINSTAKE() (*big.Int, error) {
	return _VerifierRegistry.Contract.DEFAULTMINSTAKE(&_VerifierRegistry.CallOpts)
}

// DEFAULTMINSTAKE is a free data retrieval call binding the contract method 0xdff3ece9.
//
// Solidity: function DEFAULT_MIN_STAKE() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCallerSession) DEFAULTMINSTAKE() (*big.Int, error) {
	return _VerifierRegistry.Contract.DEFAULTMINSTAKE(&_VerifierRegistry.CallOpts)
}

// WITHDRAWCOOLDOWNBLOCKS is a free data retrieval call binding the contract method 0x97c67393.
//
// Solidity: function WITHDRAW_COOLDOWN_BLOCKS() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCaller) WITHDRAWCOOLDOWNBLOCKS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "WITHDRAW_COOLDOWN_BLOCKS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WITHDRAWCOOLDOWNBLOCKS is a free data retrieval call binding the contract method 0x97c67393.
//
// Solidity: function WITHDRAW_COOLDOWN_BLOCKS() view returns(uint256)
func (_VerifierRegistry *VerifierRegistrySession) WITHDRAWCOOLDOWNBLOCKS() (*big.Int, error) {
	return _VerifierRegistry.Contract.WITHDRAWCOOLDOWNBLOCKS(&_VerifierRegistry.CallOpts)
}

// WITHDRAWCOOLDOWNBLOCKS is a free data retrieval call binding the contract method 0x97c67393.
//
// Solidity: function WITHDRAW_COOLDOWN_BLOCKS() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCallerSession) WITHDRAWCOOLDOWNBLOCKS() (*big.Int, error) {
	return _VerifierRegistry.Contract.WITHDRAWCOOLDOWNBLOCKS(&_VerifierRegistry.CallOpts)
}

// Aegis is a free data retrieval call binding the contract method 0xa10d1dcd.
//
// Solidity: function aegis() view returns(address)
func (_VerifierRegistry *VerifierRegistryCaller) Aegis(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "aegis")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Aegis is a free data retrieval call binding the contract method 0xa10d1dcd.
//
// Solidity: function aegis() view returns(address)
func (_VerifierRegistry *VerifierRegistrySession) Aegis() (common.Address, error) {
	return _VerifierRegistry.Contract.Aegis(&_VerifierRegistry.CallOpts)
}

// Aegis is a free data retrieval call binding the contract method 0xa10d1dcd.
//
// Solidity: function aegis() view returns(address)
func (_VerifierRegistry *VerifierRegistryCallerSession) Aegis() (common.Address, error) {
	return _VerifierRegistry.Contract.Aegis(&_VerifierRegistry.CallOpts)
}

// GetAccuracy is a free data retrieval call binding the contract method 0x783df30a.
//
// Solidity: function getAccuracy(address verifier) view returns(uint256 bps)
func (_VerifierRegistry *VerifierRegistryCaller) GetAccuracy(opts *bind.CallOpts, verifier common.Address) (*big.Int, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "getAccuracy", verifier)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAccuracy is a free data retrieval call binding the contract method 0x783df30a.
//
// Solidity: function getAccuracy(address verifier) view returns(uint256 bps)
func (_VerifierRegistry *VerifierRegistrySession) GetAccuracy(verifier common.Address) (*big.Int, error) {
	return _VerifierRegistry.Contract.GetAccuracy(&_VerifierRegistry.CallOpts, verifier)
}

// GetAccuracy is a free data retrieval call binding the contract method 0x783df30a.
//
// Solidity: function getAccuracy(address verifier) view returns(uint256 bps)
func (_VerifierRegistry *VerifierRegistryCallerSession) GetAccuracy(verifier common.Address) (*big.Int, error) {
	return _VerifierRegistry.Contract.GetAccuracy(&_VerifierRegistry.CallOpts, verifier)
}

// INftContract is a free data retrieval call binding the contract method 0x0dc3d284.
//
// Solidity: function iNftContract() view returns(address)
func (_VerifierRegistry *VerifierRegistryCaller) INftContract(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "iNftContract")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// INftContract is a free data retrieval call binding the contract method 0x0dc3d284.
//
// Solidity: function iNftContract() view returns(address)
func (_VerifierRegistry *VerifierRegistrySession) INftContract() (common.Address, error) {
	return _VerifierRegistry.Contract.INftContract(&_VerifierRegistry.CallOpts)
}

// INftContract is a free data retrieval call binding the contract method 0x0dc3d284.
//
// Solidity: function iNftContract() view returns(address)
func (_VerifierRegistry *VerifierRegistryCallerSession) INftContract() (common.Address, error) {
	return _VerifierRegistry.Contract.INftContract(&_VerifierRegistry.CallOpts)
}

// IsActive is a free data retrieval call binding the contract method 0x9f8a13d7.
//
// Solidity: function isActive(address verifier) view returns(bool)
func (_VerifierRegistry *VerifierRegistryCaller) IsActive(opts *bind.CallOpts, verifier common.Address) (bool, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "isActive", verifier)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsActive is a free data retrieval call binding the contract method 0x9f8a13d7.
//
// Solidity: function isActive(address verifier) view returns(bool)
func (_VerifierRegistry *VerifierRegistrySession) IsActive(verifier common.Address) (bool, error) {
	return _VerifierRegistry.Contract.IsActive(&_VerifierRegistry.CallOpts, verifier)
}

// IsActive is a free data retrieval call binding the contract method 0x9f8a13d7.
//
// Solidity: function isActive(address verifier) view returns(bool)
func (_VerifierRegistry *VerifierRegistryCallerSession) IsActive(verifier common.Address) (bool, error) {
	return _VerifierRegistry.Contract.IsActive(&_VerifierRegistry.CallOpts, verifier)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCaller) MinStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "minStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_VerifierRegistry *VerifierRegistrySession) MinStake() (*big.Int, error) {
	return _VerifierRegistry.Contract.MinStake(&_VerifierRegistry.CallOpts)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCallerSession) MinStake() (*big.Int, error) {
	return _VerifierRegistry.Contract.MinStake(&_VerifierRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_VerifierRegistry *VerifierRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_VerifierRegistry *VerifierRegistrySession) Owner() (common.Address, error) {
	return _VerifierRegistry.Contract.Owner(&_VerifierRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_VerifierRegistry *VerifierRegistryCallerSession) Owner() (common.Address, error) {
	return _VerifierRegistry.Contract.Owner(&_VerifierRegistry.CallOpts)
}

// StakeOf is a free data retrieval call binding the contract method 0x42623360.
//
// Solidity: function stakeOf(address verifier) view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCaller) StakeOf(opts *bind.CallOpts, verifier common.Address) (*big.Int, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "stakeOf", verifier)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeOf is a free data retrieval call binding the contract method 0x42623360.
//
// Solidity: function stakeOf(address verifier) view returns(uint256)
func (_VerifierRegistry *VerifierRegistrySession) StakeOf(verifier common.Address) (*big.Int, error) {
	return _VerifierRegistry.Contract.StakeOf(&_VerifierRegistry.CallOpts, verifier)
}

// StakeOf is a free data retrieval call binding the contract method 0x42623360.
//
// Solidity: function stakeOf(address verifier) view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCallerSession) StakeOf(verifier common.Address) (*big.Int, error) {
	return _VerifierRegistry.Contract.StakeOf(&_VerifierRegistry.CallOpts, verifier)
}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_VerifierRegistry *VerifierRegistryCaller) Usdc(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "usdc")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_VerifierRegistry *VerifierRegistrySession) Usdc() (common.Address, error) {
	return _VerifierRegistry.Contract.Usdc(&_VerifierRegistry.CallOpts)
}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_VerifierRegistry *VerifierRegistryCallerSession) Usdc() (common.Address, error) {
	return _VerifierRegistry.Contract.Usdc(&_VerifierRegistry.CallOpts)
}

// VerifierAddresses is a free data retrieval call binding the contract method 0x6de05398.
//
// Solidity: function verifierAddresses(uint256 ) view returns(address)
func (_VerifierRegistry *VerifierRegistryCaller) VerifierAddresses(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "verifierAddresses", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// VerifierAddresses is a free data retrieval call binding the contract method 0x6de05398.
//
// Solidity: function verifierAddresses(uint256 ) view returns(address)
func (_VerifierRegistry *VerifierRegistrySession) VerifierAddresses(arg0 *big.Int) (common.Address, error) {
	return _VerifierRegistry.Contract.VerifierAddresses(&_VerifierRegistry.CallOpts, arg0)
}

// VerifierAddresses is a free data retrieval call binding the contract method 0x6de05398.
//
// Solidity: function verifierAddresses(uint256 ) view returns(address)
func (_VerifierRegistry *VerifierRegistryCallerSession) VerifierAddresses(arg0 *big.Int) (common.Address, error) {
	return _VerifierRegistry.Contract.VerifierAddresses(&_VerifierRegistry.CallOpts, arg0)
}

// VerifierCount is a free data retrieval call binding the contract method 0xfb4adf52.
//
// Solidity: function verifierCount() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCaller) VerifierCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "verifierCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VerifierCount is a free data retrieval call binding the contract method 0xfb4adf52.
//
// Solidity: function verifierCount() view returns(uint256)
func (_VerifierRegistry *VerifierRegistrySession) VerifierCount() (*big.Int, error) {
	return _VerifierRegistry.Contract.VerifierCount(&_VerifierRegistry.CallOpts)
}

// VerifierCount is a free data retrieval call binding the contract method 0xfb4adf52.
//
// Solidity: function verifierCount() view returns(uint256)
func (_VerifierRegistry *VerifierRegistryCallerSession) VerifierCount() (*big.Int, error) {
	return _VerifierRegistry.Contract.VerifierCount(&_VerifierRegistry.CallOpts)
}

// Verifiers is a free data retrieval call binding the contract method 0x6c824487.
//
// Solidity: function verifiers(address ) view returns(uint256 stake, uint256 votesTotal, uint256 votesCorrect, uint256 iNftId, uint256 withdrawableAt, bool active)
func (_VerifierRegistry *VerifierRegistryCaller) Verifiers(opts *bind.CallOpts, arg0 common.Address) (struct {
	Stake          *big.Int
	VotesTotal     *big.Int
	VotesCorrect   *big.Int
	INftId         *big.Int
	WithdrawableAt *big.Int
	Active         bool
}, error) {
	var out []interface{}
	err := _VerifierRegistry.contract.Call(opts, &out, "verifiers", arg0)

	outstruct := new(struct {
		Stake          *big.Int
		VotesTotal     *big.Int
		VotesCorrect   *big.Int
		INftId         *big.Int
		WithdrawableAt *big.Int
		Active         bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Stake = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.VotesTotal = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.VotesCorrect = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.INftId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.WithdrawableAt = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Active = *abi.ConvertType(out[5], new(bool)).(*bool)

	return *outstruct, err

}

// Verifiers is a free data retrieval call binding the contract method 0x6c824487.
//
// Solidity: function verifiers(address ) view returns(uint256 stake, uint256 votesTotal, uint256 votesCorrect, uint256 iNftId, uint256 withdrawableAt, bool active)
func (_VerifierRegistry *VerifierRegistrySession) Verifiers(arg0 common.Address) (struct {
	Stake          *big.Int
	VotesTotal     *big.Int
	VotesCorrect   *big.Int
	INftId         *big.Int
	WithdrawableAt *big.Int
	Active         bool
}, error) {
	return _VerifierRegistry.Contract.Verifiers(&_VerifierRegistry.CallOpts, arg0)
}

// Verifiers is a free data retrieval call binding the contract method 0x6c824487.
//
// Solidity: function verifiers(address ) view returns(uint256 stake, uint256 votesTotal, uint256 votesCorrect, uint256 iNftId, uint256 withdrawableAt, bool active)
func (_VerifierRegistry *VerifierRegistryCallerSession) Verifiers(arg0 common.Address) (struct {
	Stake          *big.Int
	VotesTotal     *big.Int
	VotesCorrect   *big.Int
	INftId         *big.Int
	WithdrawableAt *big.Int
	Active         bool
}, error) {
	return _VerifierRegistry.Contract.Verifiers(&_VerifierRegistry.CallOpts, arg0)
}

// AddStake is a paid mutator transaction binding the contract method 0xeb4f16b5.
//
// Solidity: function addStake(uint256 amount) returns()
func (_VerifierRegistry *VerifierRegistryTransactor) AddStake(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "addStake", amount)
}

// AddStake is a paid mutator transaction binding the contract method 0xeb4f16b5.
//
// Solidity: function addStake(uint256 amount) returns()
func (_VerifierRegistry *VerifierRegistrySession) AddStake(amount *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.AddStake(&_VerifierRegistry.TransactOpts, amount)
}

// AddStake is a paid mutator transaction binding the contract method 0xeb4f16b5.
//
// Solidity: function addStake(uint256 amount) returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) AddStake(amount *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.AddStake(&_VerifierRegistry.TransactOpts, amount)
}

// RecordVote is a paid mutator transaction binding the contract method 0x8cbbb50c.
//
// Solidity: function recordVote(address verifier, bool wasCorrect) returns()
func (_VerifierRegistry *VerifierRegistryTransactor) RecordVote(opts *bind.TransactOpts, verifier common.Address, wasCorrect bool) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "recordVote", verifier, wasCorrect)
}

// RecordVote is a paid mutator transaction binding the contract method 0x8cbbb50c.
//
// Solidity: function recordVote(address verifier, bool wasCorrect) returns()
func (_VerifierRegistry *VerifierRegistrySession) RecordVote(verifier common.Address, wasCorrect bool) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.RecordVote(&_VerifierRegistry.TransactOpts, verifier, wasCorrect)
}

// RecordVote is a paid mutator transaction binding the contract method 0x8cbbb50c.
//
// Solidity: function recordVote(address verifier, bool wasCorrect) returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) RecordVote(verifier common.Address, wasCorrect bool) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.RecordVote(&_VerifierRegistry.TransactOpts, verifier, wasCorrect)
}

// Register is a paid mutator transaction binding the contract method 0xd66d6c10.
//
// Solidity: function register(uint256 amount, uint256 iNftId) returns()
func (_VerifierRegistry *VerifierRegistryTransactor) Register(opts *bind.TransactOpts, amount *big.Int, iNftId *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "register", amount, iNftId)
}

// Register is a paid mutator transaction binding the contract method 0xd66d6c10.
//
// Solidity: function register(uint256 amount, uint256 iNftId) returns()
func (_VerifierRegistry *VerifierRegistrySession) Register(amount *big.Int, iNftId *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.Register(&_VerifierRegistry.TransactOpts, amount, iNftId)
}

// Register is a paid mutator transaction binding the contract method 0xd66d6c10.
//
// Solidity: function register(uint256 amount, uint256 iNftId) returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) Register(amount *big.Int, iNftId *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.Register(&_VerifierRegistry.TransactOpts, amount, iNftId)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_VerifierRegistry *VerifierRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_VerifierRegistry *VerifierRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _VerifierRegistry.Contract.RenounceOwnership(&_VerifierRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _VerifierRegistry.Contract.RenounceOwnership(&_VerifierRegistry.TransactOpts)
}

// RequestWithdraw is a paid mutator transaction binding the contract method 0xb3423eec.
//
// Solidity: function requestWithdraw() returns()
func (_VerifierRegistry *VerifierRegistryTransactor) RequestWithdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "requestWithdraw")
}

// RequestWithdraw is a paid mutator transaction binding the contract method 0xb3423eec.
//
// Solidity: function requestWithdraw() returns()
func (_VerifierRegistry *VerifierRegistrySession) RequestWithdraw() (*types.Transaction, error) {
	return _VerifierRegistry.Contract.RequestWithdraw(&_VerifierRegistry.TransactOpts)
}

// RequestWithdraw is a paid mutator transaction binding the contract method 0xb3423eec.
//
// Solidity: function requestWithdraw() returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) RequestWithdraw() (*types.Transaction, error) {
	return _VerifierRegistry.Contract.RequestWithdraw(&_VerifierRegistry.TransactOpts)
}

// SetAegis is a paid mutator transaction binding the contract method 0xd1d63601.
//
// Solidity: function setAegis(address _aegis) returns()
func (_VerifierRegistry *VerifierRegistryTransactor) SetAegis(opts *bind.TransactOpts, _aegis common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "setAegis", _aegis)
}

// SetAegis is a paid mutator transaction binding the contract method 0xd1d63601.
//
// Solidity: function setAegis(address _aegis) returns()
func (_VerifierRegistry *VerifierRegistrySession) SetAegis(_aegis common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.SetAegis(&_VerifierRegistry.TransactOpts, _aegis)
}

// SetAegis is a paid mutator transaction binding the contract method 0xd1d63601.
//
// Solidity: function setAegis(address _aegis) returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) SetAegis(_aegis common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.SetAegis(&_VerifierRegistry.TransactOpts, _aegis)
}

// SetINftContract is a paid mutator transaction binding the contract method 0x7ce4611f.
//
// Solidity: function setINftContract(address _iNft) returns()
func (_VerifierRegistry *VerifierRegistryTransactor) SetINftContract(opts *bind.TransactOpts, _iNft common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "setINftContract", _iNft)
}

// SetINftContract is a paid mutator transaction binding the contract method 0x7ce4611f.
//
// Solidity: function setINftContract(address _iNft) returns()
func (_VerifierRegistry *VerifierRegistrySession) SetINftContract(_iNft common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.SetINftContract(&_VerifierRegistry.TransactOpts, _iNft)
}

// SetINftContract is a paid mutator transaction binding the contract method 0x7ce4611f.
//
// Solidity: function setINftContract(address _iNft) returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) SetINftContract(_iNft common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.SetINftContract(&_VerifierRegistry.TransactOpts, _iNft)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_VerifierRegistry *VerifierRegistryTransactor) SetMinStake(opts *bind.TransactOpts, _minStake *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "setMinStake", _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_VerifierRegistry *VerifierRegistrySession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.SetMinStake(&_VerifierRegistry.TransactOpts, _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.SetMinStake(&_VerifierRegistry.TransactOpts, _minStake)
}

// Slash is a paid mutator transaction binding the contract method 0xbfac5990.
//
// Solidity: function slash(address verifier, uint256 amount, address recipient) returns(uint256 actualSlashed)
func (_VerifierRegistry *VerifierRegistryTransactor) Slash(opts *bind.TransactOpts, verifier common.Address, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "slash", verifier, amount, recipient)
}

// Slash is a paid mutator transaction binding the contract method 0xbfac5990.
//
// Solidity: function slash(address verifier, uint256 amount, address recipient) returns(uint256 actualSlashed)
func (_VerifierRegistry *VerifierRegistrySession) Slash(verifier common.Address, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.Slash(&_VerifierRegistry.TransactOpts, verifier, amount, recipient)
}

// Slash is a paid mutator transaction binding the contract method 0xbfac5990.
//
// Solidity: function slash(address verifier, uint256 amount, address recipient) returns(uint256 actualSlashed)
func (_VerifierRegistry *VerifierRegistryTransactorSession) Slash(verifier common.Address, amount *big.Int, recipient common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.Slash(&_VerifierRegistry.TransactOpts, verifier, amount, recipient)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_VerifierRegistry *VerifierRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_VerifierRegistry *VerifierRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.TransferOwnership(&_VerifierRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _VerifierRegistry.Contract.TransferOwnership(&_VerifierRegistry.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_VerifierRegistry *VerifierRegistryTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _VerifierRegistry.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_VerifierRegistry *VerifierRegistrySession) Withdraw() (*types.Transaction, error) {
	return _VerifierRegistry.Contract.Withdraw(&_VerifierRegistry.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_VerifierRegistry *VerifierRegistryTransactorSession) Withdraw() (*types.Transaction, error) {
	return _VerifierRegistry.Contract.Withdraw(&_VerifierRegistry.TransactOpts)
}

// VerifierRegistryAegisSetIterator is returned from FilterAegisSet and is used to iterate over the raw logs and unpacked data for AegisSet events raised by the VerifierRegistry contract.
type VerifierRegistryAegisSetIterator struct {
	Event *VerifierRegistryAegisSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryAegisSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryAegisSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryAegisSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryAegisSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryAegisSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryAegisSet represents a AegisSet event raised by the VerifierRegistry contract.
type VerifierRegistryAegisSet struct {
	Aegis common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterAegisSet is a free log retrieval operation binding the contract event 0x8a4f944711cc731fe00e14afd079cdcbd5d5b473d1215f5c32f05f9ee622a4f0.
//
// Solidity: event AegisSet(address indexed aegis)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterAegisSet(opts *bind.FilterOpts, aegis []common.Address) (*VerifierRegistryAegisSetIterator, error) {

	var aegisRule []interface{}
	for _, aegisItem := range aegis {
		aegisRule = append(aegisRule, aegisItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "AegisSet", aegisRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryAegisSetIterator{contract: _VerifierRegistry.contract, event: "AegisSet", logs: logs, sub: sub}, nil
}

// WatchAegisSet is a free log subscription operation binding the contract event 0x8a4f944711cc731fe00e14afd079cdcbd5d5b473d1215f5c32f05f9ee622a4f0.
//
// Solidity: event AegisSet(address indexed aegis)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchAegisSet(opts *bind.WatchOpts, sink chan<- *VerifierRegistryAegisSet, aegis []common.Address) (event.Subscription, error) {

	var aegisRule []interface{}
	for _, aegisItem := range aegis {
		aegisRule = append(aegisRule, aegisItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "AegisSet", aegisRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryAegisSet)
				if err := _VerifierRegistry.contract.UnpackLog(event, "AegisSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAegisSet is a log parse operation binding the contract event 0x8a4f944711cc731fe00e14afd079cdcbd5d5b473d1215f5c32f05f9ee622a4f0.
//
// Solidity: event AegisSet(address indexed aegis)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseAegisSet(log types.Log) (*VerifierRegistryAegisSet, error) {
	event := new(VerifierRegistryAegisSet)
	if err := _VerifierRegistry.contract.UnpackLog(event, "AegisSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryMinStakeUpdatedIterator is returned from FilterMinStakeUpdated and is used to iterate over the raw logs and unpacked data for MinStakeUpdated events raised by the VerifierRegistry contract.
type VerifierRegistryMinStakeUpdatedIterator struct {
	Event *VerifierRegistryMinStakeUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryMinStakeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryMinStakeUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryMinStakeUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryMinStakeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryMinStakeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryMinStakeUpdated represents a MinStakeUpdated event raised by the VerifierRegistry contract.
type VerifierRegistryMinStakeUpdated struct {
	OldMin *big.Int
	NewMin *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterMinStakeUpdated is a free log retrieval operation binding the contract event 0x171aabb8815c02fd00303450a77058600e3661eb75ce2e77972c0f080bc7099d.
//
// Solidity: event MinStakeUpdated(uint256 oldMin, uint256 newMin)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterMinStakeUpdated(opts *bind.FilterOpts) (*VerifierRegistryMinStakeUpdatedIterator, error) {

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryMinStakeUpdatedIterator{contract: _VerifierRegistry.contract, event: "MinStakeUpdated", logs: logs, sub: sub}, nil
}

// WatchMinStakeUpdated is a free log subscription operation binding the contract event 0x171aabb8815c02fd00303450a77058600e3661eb75ce2e77972c0f080bc7099d.
//
// Solidity: event MinStakeUpdated(uint256 oldMin, uint256 newMin)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchMinStakeUpdated(opts *bind.WatchOpts, sink chan<- *VerifierRegistryMinStakeUpdated) (event.Subscription, error) {

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryMinStakeUpdated)
				if err := _VerifierRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMinStakeUpdated is a log parse operation binding the contract event 0x171aabb8815c02fd00303450a77058600e3661eb75ce2e77972c0f080bc7099d.
//
// Solidity: event MinStakeUpdated(uint256 oldMin, uint256 newMin)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseMinStakeUpdated(log types.Log) (*VerifierRegistryMinStakeUpdated, error) {
	event := new(VerifierRegistryMinStakeUpdated)
	if err := _VerifierRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the VerifierRegistry contract.
type VerifierRegistryOwnershipTransferredIterator struct {
	Event *VerifierRegistryOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the VerifierRegistry contract.
type VerifierRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*VerifierRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryOwnershipTransferredIterator{contract: _VerifierRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *VerifierRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryOwnershipTransferred)
				if err := _VerifierRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*VerifierRegistryOwnershipTransferred, error) {
	event := new(VerifierRegistryOwnershipTransferred)
	if err := _VerifierRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryStakeAddedIterator is returned from FilterStakeAdded and is used to iterate over the raw logs and unpacked data for StakeAdded events raised by the VerifierRegistry contract.
type VerifierRegistryStakeAddedIterator struct {
	Event *VerifierRegistryStakeAdded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryStakeAddedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryStakeAdded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryStakeAdded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryStakeAddedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryStakeAddedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryStakeAdded represents a StakeAdded event raised by the VerifierRegistry contract.
type VerifierRegistryStakeAdded struct {
	Verifier common.Address
	Amount   *big.Int
	NewStake *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterStakeAdded is a free log retrieval operation binding the contract event 0x270d6dd254edd1d985c81cf7861b8f28fb06b6d719df04d90464034d43412440.
//
// Solidity: event StakeAdded(address indexed verifier, uint256 amount, uint256 newStake)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterStakeAdded(opts *bind.FilterOpts, verifier []common.Address) (*VerifierRegistryStakeAddedIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "StakeAdded", verifierRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryStakeAddedIterator{contract: _VerifierRegistry.contract, event: "StakeAdded", logs: logs, sub: sub}, nil
}

// WatchStakeAdded is a free log subscription operation binding the contract event 0x270d6dd254edd1d985c81cf7861b8f28fb06b6d719df04d90464034d43412440.
//
// Solidity: event StakeAdded(address indexed verifier, uint256 amount, uint256 newStake)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchStakeAdded(opts *bind.WatchOpts, sink chan<- *VerifierRegistryStakeAdded, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "StakeAdded", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryStakeAdded)
				if err := _VerifierRegistry.contract.UnpackLog(event, "StakeAdded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeAdded is a log parse operation binding the contract event 0x270d6dd254edd1d985c81cf7861b8f28fb06b6d719df04d90464034d43412440.
//
// Solidity: event StakeAdded(address indexed verifier, uint256 amount, uint256 newStake)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseStakeAdded(log types.Log) (*VerifierRegistryStakeAdded, error) {
	event := new(VerifierRegistryStakeAdded)
	if err := _VerifierRegistry.contract.UnpackLog(event, "StakeAdded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryStakeWithdrawnIterator is returned from FilterStakeWithdrawn and is used to iterate over the raw logs and unpacked data for StakeWithdrawn events raised by the VerifierRegistry contract.
type VerifierRegistryStakeWithdrawnIterator struct {
	Event *VerifierRegistryStakeWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryStakeWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryStakeWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryStakeWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryStakeWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryStakeWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryStakeWithdrawn represents a StakeWithdrawn event raised by the VerifierRegistry contract.
type VerifierRegistryStakeWithdrawn struct {
	Verifier common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterStakeWithdrawn is a free log retrieval operation binding the contract event 0x8108595eb6bad3acefa9da467d90cc2217686d5c5ac85460f8b7849c840645fc.
//
// Solidity: event StakeWithdrawn(address indexed verifier, uint256 amount)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterStakeWithdrawn(opts *bind.FilterOpts, verifier []common.Address) (*VerifierRegistryStakeWithdrawnIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "StakeWithdrawn", verifierRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryStakeWithdrawnIterator{contract: _VerifierRegistry.contract, event: "StakeWithdrawn", logs: logs, sub: sub}, nil
}

// WatchStakeWithdrawn is a free log subscription operation binding the contract event 0x8108595eb6bad3acefa9da467d90cc2217686d5c5ac85460f8b7849c840645fc.
//
// Solidity: event StakeWithdrawn(address indexed verifier, uint256 amount)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchStakeWithdrawn(opts *bind.WatchOpts, sink chan<- *VerifierRegistryStakeWithdrawn, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "StakeWithdrawn", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryStakeWithdrawn)
				if err := _VerifierRegistry.contract.UnpackLog(event, "StakeWithdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStakeWithdrawn is a log parse operation binding the contract event 0x8108595eb6bad3acefa9da467d90cc2217686d5c5ac85460f8b7849c840645fc.
//
// Solidity: event StakeWithdrawn(address indexed verifier, uint256 amount)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseStakeWithdrawn(log types.Log) (*VerifierRegistryStakeWithdrawn, error) {
	event := new(VerifierRegistryStakeWithdrawn)
	if err := _VerifierRegistry.contract.UnpackLog(event, "StakeWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryVerifierRegisteredIterator is returned from FilterVerifierRegistered and is used to iterate over the raw logs and unpacked data for VerifierRegistered events raised by the VerifierRegistry contract.
type VerifierRegistryVerifierRegisteredIterator struct {
	Event *VerifierRegistryVerifierRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryVerifierRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryVerifierRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryVerifierRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryVerifierRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryVerifierRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryVerifierRegistered represents a VerifierRegistered event raised by the VerifierRegistry contract.
type VerifierRegistryVerifierRegistered struct {
	Verifier common.Address
	Stake    *big.Int
	INftId   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVerifierRegistered is a free log retrieval operation binding the contract event 0xbd7a897cd32cbc28d279a7e71b516ec53c747f2ca382101ff71b3105db1c98d9.
//
// Solidity: event VerifierRegistered(address indexed verifier, uint256 stake, uint256 iNftId)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterVerifierRegistered(opts *bind.FilterOpts, verifier []common.Address) (*VerifierRegistryVerifierRegisteredIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "VerifierRegistered", verifierRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryVerifierRegisteredIterator{contract: _VerifierRegistry.contract, event: "VerifierRegistered", logs: logs, sub: sub}, nil
}

// WatchVerifierRegistered is a free log subscription operation binding the contract event 0xbd7a897cd32cbc28d279a7e71b516ec53c747f2ca382101ff71b3105db1c98d9.
//
// Solidity: event VerifierRegistered(address indexed verifier, uint256 stake, uint256 iNftId)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchVerifierRegistered(opts *bind.WatchOpts, sink chan<- *VerifierRegistryVerifierRegistered, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "VerifierRegistered", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryVerifierRegistered)
				if err := _VerifierRegistry.contract.UnpackLog(event, "VerifierRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVerifierRegistered is a log parse operation binding the contract event 0xbd7a897cd32cbc28d279a7e71b516ec53c747f2ca382101ff71b3105db1c98d9.
//
// Solidity: event VerifierRegistered(address indexed verifier, uint256 stake, uint256 iNftId)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseVerifierRegistered(log types.Log) (*VerifierRegistryVerifierRegistered, error) {
	event := new(VerifierRegistryVerifierRegistered)
	if err := _VerifierRegistry.contract.UnpackLog(event, "VerifierRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryVerifierSlashedIterator is returned from FilterVerifierSlashed and is used to iterate over the raw logs and unpacked data for VerifierSlashed events raised by the VerifierRegistry contract.
type VerifierRegistryVerifierSlashedIterator struct {
	Event *VerifierRegistryVerifierSlashed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryVerifierSlashedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryVerifierSlashed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryVerifierSlashed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryVerifierSlashedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryVerifierSlashedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryVerifierSlashed represents a VerifierSlashed event raised by the VerifierRegistry contract.
type VerifierRegistryVerifierSlashed struct {
	Verifier  common.Address
	Recipient common.Address
	Amount    *big.Int
	NewStake  *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterVerifierSlashed is a free log retrieval operation binding the contract event 0x3b1f5512e0b2fb04b3aeb21dbc89b7bdd91902b28a20dd85605318c373bb9e58.
//
// Solidity: event VerifierSlashed(address indexed verifier, address indexed recipient, uint256 amount, uint256 newStake)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterVerifierSlashed(opts *bind.FilterOpts, verifier []common.Address, recipient []common.Address) (*VerifierRegistryVerifierSlashedIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "VerifierSlashed", verifierRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryVerifierSlashedIterator{contract: _VerifierRegistry.contract, event: "VerifierSlashed", logs: logs, sub: sub}, nil
}

// WatchVerifierSlashed is a free log subscription operation binding the contract event 0x3b1f5512e0b2fb04b3aeb21dbc89b7bdd91902b28a20dd85605318c373bb9e58.
//
// Solidity: event VerifierSlashed(address indexed verifier, address indexed recipient, uint256 amount, uint256 newStake)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchVerifierSlashed(opts *bind.WatchOpts, sink chan<- *VerifierRegistryVerifierSlashed, verifier []common.Address, recipient []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}
	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "VerifierSlashed", verifierRule, recipientRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryVerifierSlashed)
				if err := _VerifierRegistry.contract.UnpackLog(event, "VerifierSlashed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVerifierSlashed is a log parse operation binding the contract event 0x3b1f5512e0b2fb04b3aeb21dbc89b7bdd91902b28a20dd85605318c373bb9e58.
//
// Solidity: event VerifierSlashed(address indexed verifier, address indexed recipient, uint256 amount, uint256 newStake)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseVerifierSlashed(log types.Log) (*VerifierRegistryVerifierSlashed, error) {
	event := new(VerifierRegistryVerifierSlashed)
	if err := _VerifierRegistry.contract.UnpackLog(event, "VerifierSlashed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryVoteRecordedIterator is returned from FilterVoteRecorded and is used to iterate over the raw logs and unpacked data for VoteRecorded events raised by the VerifierRegistry contract.
type VerifierRegistryVoteRecordedIterator struct {
	Event *VerifierRegistryVoteRecorded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryVoteRecordedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryVoteRecorded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryVoteRecorded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryVoteRecordedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryVoteRecordedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryVoteRecorded represents a VoteRecorded event raised by the VerifierRegistry contract.
type VerifierRegistryVoteRecorded struct {
	Verifier     common.Address
	WasCorrect   bool
	VotesTotal   *big.Int
	VotesCorrect *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterVoteRecorded is a free log retrieval operation binding the contract event 0x683c0a863ec62edc2aef867991e3394b55693d54789fea3b7e46f9d27fdef99e.
//
// Solidity: event VoteRecorded(address indexed verifier, bool wasCorrect, uint256 votesTotal, uint256 votesCorrect)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterVoteRecorded(opts *bind.FilterOpts, verifier []common.Address) (*VerifierRegistryVoteRecordedIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "VoteRecorded", verifierRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryVoteRecordedIterator{contract: _VerifierRegistry.contract, event: "VoteRecorded", logs: logs, sub: sub}, nil
}

// WatchVoteRecorded is a free log subscription operation binding the contract event 0x683c0a863ec62edc2aef867991e3394b55693d54789fea3b7e46f9d27fdef99e.
//
// Solidity: event VoteRecorded(address indexed verifier, bool wasCorrect, uint256 votesTotal, uint256 votesCorrect)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchVoteRecorded(opts *bind.WatchOpts, sink chan<- *VerifierRegistryVoteRecorded, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "VoteRecorded", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryVoteRecorded)
				if err := _VerifierRegistry.contract.UnpackLog(event, "VoteRecorded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseVoteRecorded is a log parse operation binding the contract event 0x683c0a863ec62edc2aef867991e3394b55693d54789fea3b7e46f9d27fdef99e.
//
// Solidity: event VoteRecorded(address indexed verifier, bool wasCorrect, uint256 votesTotal, uint256 votesCorrect)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseVoteRecorded(log types.Log) (*VerifierRegistryVoteRecorded, error) {
	event := new(VerifierRegistryVoteRecorded)
	if err := _VerifierRegistry.contract.UnpackLog(event, "VoteRecorded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// VerifierRegistryWithdrawalRequestedIterator is returned from FilterWithdrawalRequested and is used to iterate over the raw logs and unpacked data for WithdrawalRequested events raised by the VerifierRegistry contract.
type VerifierRegistryWithdrawalRequestedIterator struct {
	Event *VerifierRegistryWithdrawalRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *VerifierRegistryWithdrawalRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(VerifierRegistryWithdrawalRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(VerifierRegistryWithdrawalRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *VerifierRegistryWithdrawalRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *VerifierRegistryWithdrawalRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// VerifierRegistryWithdrawalRequested represents a WithdrawalRequested event raised by the VerifierRegistry contract.
type VerifierRegistryWithdrawalRequested struct {
	Verifier    common.Address
	UnlockBlock *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterWithdrawalRequested is a free log retrieval operation binding the contract event 0xe670e4e82118d22a1f9ee18920455ebc958bae26a90a05d31d3378788b1b0e44.
//
// Solidity: event WithdrawalRequested(address indexed verifier, uint256 unlockBlock)
func (_VerifierRegistry *VerifierRegistryFilterer) FilterWithdrawalRequested(opts *bind.FilterOpts, verifier []common.Address) (*VerifierRegistryWithdrawalRequestedIterator, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.FilterLogs(opts, "WithdrawalRequested", verifierRule)
	if err != nil {
		return nil, err
	}
	return &VerifierRegistryWithdrawalRequestedIterator{contract: _VerifierRegistry.contract, event: "WithdrawalRequested", logs: logs, sub: sub}, nil
}

// WatchWithdrawalRequested is a free log subscription operation binding the contract event 0xe670e4e82118d22a1f9ee18920455ebc958bae26a90a05d31d3378788b1b0e44.
//
// Solidity: event WithdrawalRequested(address indexed verifier, uint256 unlockBlock)
func (_VerifierRegistry *VerifierRegistryFilterer) WatchWithdrawalRequested(opts *bind.WatchOpts, sink chan<- *VerifierRegistryWithdrawalRequested, verifier []common.Address) (event.Subscription, error) {

	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _VerifierRegistry.contract.WatchLogs(opts, "WithdrawalRequested", verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(VerifierRegistryWithdrawalRequested)
				if err := _VerifierRegistry.contract.UnpackLog(event, "WithdrawalRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawalRequested is a log parse operation binding the contract event 0xe670e4e82118d22a1f9ee18920455ebc958bae26a90a05d31d3378788b1b0e44.
//
// Solidity: event WithdrawalRequested(address indexed verifier, uint256 unlockBlock)
func (_VerifierRegistry *VerifierRegistryFilterer) ParseWithdrawalRequested(log types.Log) (*VerifierRegistryWithdrawalRequested, error) {
	event := new(VerifierRegistryWithdrawalRequested)
	if err := _VerifierRegistry.contract.UnpackLog(event, "WithdrawalRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
