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

// AegisContractMetaData contains all meta data concerning the AegisContract contract.
var AegisContractMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_usdc\",\"type\":\"address\",\"internalType\":\"contractIUSDC\"},{\"name\":\"_registry\",\"type\":\"address\",\"internalType\":\"contractIVerifierRegistry\"},{\"name\":\"_treasury\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_CLAIM_WINDOW\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEFAULT_COMMIT_WINDOW\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"DEFAULT_REVEAL_WINDOW\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_REVEALS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SLASH_EXECUTOR_ON_FAIL\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"SLASH_PER_DISSENT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"cancelStaleJob\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"commitVote\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"commitHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"commits\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"jobRevealers\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"jobs\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"client\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executor\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executorReimbursement\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"executorFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifierBounty\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"specHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"txHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"reportedOutcomeHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"claimDeadline\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"commitDeadline\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"revealDeadline\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumAegisContract.Status\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"nextJobId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"postJob\",\"inputs\":[{\"name\":\"specHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"executor\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"executorReimbursement\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"executorFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifierBounty\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIVerifierRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"revealCount\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"revealVote\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verdict\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nonce\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"reveals\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"verdict\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"revealed\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"settle\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"submitClaim\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"txHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"reportedOutcomeHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"treasury\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"usdc\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIUSDC\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"ClaimSubmitted\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"txHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"reportedOutcomeHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"JobCancelled\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"JobPosted\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"client\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"executor\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"specHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"},{\"name\":\"executorReimbursement\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"executorFee\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"verifierBounty\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"JobSettled\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"finalVerdict\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"},{\"name\":\"forVotes\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"totalReveals\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VoteCommitted\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"commitHash\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"VoteRevealed\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"verifier\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"verdict\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AddressInsufficientBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"AlreadyCommitted\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"AlreadyRevealed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"CommitMismatch\",\"inputs\":[{\"name\":\"expected\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"actual\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"DeadlineNotPassed\",\"inputs\":[{\"name\":\"deadline\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"nowTs\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"type\":\"error\",\"name\":\"DeadlinePassed\",\"inputs\":[{\"name\":\"deadline\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"nowTs\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"type\":\"error\",\"name\":\"FailedInnerCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"JobNotInStatus\",\"inputs\":[{\"name\":\"jobId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"expected\",\"type\":\"uint8\",\"internalType\":\"enumAegisContract.Status\"},{\"name\":\"actual\",\"type\":\"uint8\",\"internalType\":\"enumAegisContract.Status\"}]},{\"type\":\"error\",\"name\":\"NoCommit\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotClient\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotExecutor\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"TooFewReveals\",\"inputs\":[{\"name\":\"got\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"required\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"VerifierNotActive\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ZeroAmount\",\"inputs\":[]}]",
}

// AegisContractABI is the input ABI used to generate the binding from.
// Deprecated: Use AegisContractMetaData.ABI instead.
var AegisContractABI = AegisContractMetaData.ABI

// AegisContract is an auto generated Go binding around an Ethereum contract.
type AegisContract struct {
	AegisContractCaller     // Read-only binding to the contract
	AegisContractTransactor // Write-only binding to the contract
	AegisContractFilterer   // Log filterer for contract events
}

// AegisContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type AegisContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AegisContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AegisContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AegisContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AegisContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AegisContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AegisContractSession struct {
	Contract     *AegisContract    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AegisContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AegisContractCallerSession struct {
	Contract *AegisContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// AegisContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AegisContractTransactorSession struct {
	Contract     *AegisContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// AegisContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type AegisContractRaw struct {
	Contract *AegisContract // Generic contract binding to access the raw methods on
}

// AegisContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AegisContractCallerRaw struct {
	Contract *AegisContractCaller // Generic read-only contract binding to access the raw methods on
}

// AegisContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AegisContractTransactorRaw struct {
	Contract *AegisContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAegisContract creates a new instance of AegisContract, bound to a specific deployed contract.
func NewAegisContract(address common.Address, backend bind.ContractBackend) (*AegisContract, error) {
	contract, err := bindAegisContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AegisContract{AegisContractCaller: AegisContractCaller{contract: contract}, AegisContractTransactor: AegisContractTransactor{contract: contract}, AegisContractFilterer: AegisContractFilterer{contract: contract}}, nil
}

// NewAegisContractCaller creates a new read-only instance of AegisContract, bound to a specific deployed contract.
func NewAegisContractCaller(address common.Address, caller bind.ContractCaller) (*AegisContractCaller, error) {
	contract, err := bindAegisContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AegisContractCaller{contract: contract}, nil
}

// NewAegisContractTransactor creates a new write-only instance of AegisContract, bound to a specific deployed contract.
func NewAegisContractTransactor(address common.Address, transactor bind.ContractTransactor) (*AegisContractTransactor, error) {
	contract, err := bindAegisContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AegisContractTransactor{contract: contract}, nil
}

// NewAegisContractFilterer creates a new log filterer instance of AegisContract, bound to a specific deployed contract.
func NewAegisContractFilterer(address common.Address, filterer bind.ContractFilterer) (*AegisContractFilterer, error) {
	contract, err := bindAegisContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AegisContractFilterer{contract: contract}, nil
}

// bindAegisContract binds a generic wrapper to an already deployed contract.
func bindAegisContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AegisContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AegisContract *AegisContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AegisContract.Contract.AegisContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AegisContract *AegisContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AegisContract.Contract.AegisContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AegisContract *AegisContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AegisContract.Contract.AegisContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AegisContract *AegisContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AegisContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AegisContract *AegisContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AegisContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AegisContract *AegisContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AegisContract.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTCLAIMWINDOW is a free data retrieval call binding the contract method 0x4e333b25.
//
// Solidity: function DEFAULT_CLAIM_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractCaller) DEFAULTCLAIMWINDOW(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "DEFAULT_CLAIM_WINDOW")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// DEFAULTCLAIMWINDOW is a free data retrieval call binding the contract method 0x4e333b25.
//
// Solidity: function DEFAULT_CLAIM_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractSession) DEFAULTCLAIMWINDOW() (uint64, error) {
	return _AegisContract.Contract.DEFAULTCLAIMWINDOW(&_AegisContract.CallOpts)
}

// DEFAULTCLAIMWINDOW is a free data retrieval call binding the contract method 0x4e333b25.
//
// Solidity: function DEFAULT_CLAIM_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractCallerSession) DEFAULTCLAIMWINDOW() (uint64, error) {
	return _AegisContract.Contract.DEFAULTCLAIMWINDOW(&_AegisContract.CallOpts)
}

// DEFAULTCOMMITWINDOW is a free data retrieval call binding the contract method 0xbed80efb.
//
// Solidity: function DEFAULT_COMMIT_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractCaller) DEFAULTCOMMITWINDOW(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "DEFAULT_COMMIT_WINDOW")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// DEFAULTCOMMITWINDOW is a free data retrieval call binding the contract method 0xbed80efb.
//
// Solidity: function DEFAULT_COMMIT_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractSession) DEFAULTCOMMITWINDOW() (uint64, error) {
	return _AegisContract.Contract.DEFAULTCOMMITWINDOW(&_AegisContract.CallOpts)
}

// DEFAULTCOMMITWINDOW is a free data retrieval call binding the contract method 0xbed80efb.
//
// Solidity: function DEFAULT_COMMIT_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractCallerSession) DEFAULTCOMMITWINDOW() (uint64, error) {
	return _AegisContract.Contract.DEFAULTCOMMITWINDOW(&_AegisContract.CallOpts)
}

// DEFAULTREVEALWINDOW is a free data retrieval call binding the contract method 0x8416c1f8.
//
// Solidity: function DEFAULT_REVEAL_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractCaller) DEFAULTREVEALWINDOW(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "DEFAULT_REVEAL_WINDOW")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// DEFAULTREVEALWINDOW is a free data retrieval call binding the contract method 0x8416c1f8.
//
// Solidity: function DEFAULT_REVEAL_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractSession) DEFAULTREVEALWINDOW() (uint64, error) {
	return _AegisContract.Contract.DEFAULTREVEALWINDOW(&_AegisContract.CallOpts)
}

// DEFAULTREVEALWINDOW is a free data retrieval call binding the contract method 0x8416c1f8.
//
// Solidity: function DEFAULT_REVEAL_WINDOW() view returns(uint64)
func (_AegisContract *AegisContractCallerSession) DEFAULTREVEALWINDOW() (uint64, error) {
	return _AegisContract.Contract.DEFAULTREVEALWINDOW(&_AegisContract.CallOpts)
}

// MINREVEALS is a free data retrieval call binding the contract method 0x2c78a525.
//
// Solidity: function MIN_REVEALS() view returns(uint256)
func (_AegisContract *AegisContractCaller) MINREVEALS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "MIN_REVEALS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINREVEALS is a free data retrieval call binding the contract method 0x2c78a525.
//
// Solidity: function MIN_REVEALS() view returns(uint256)
func (_AegisContract *AegisContractSession) MINREVEALS() (*big.Int, error) {
	return _AegisContract.Contract.MINREVEALS(&_AegisContract.CallOpts)
}

// MINREVEALS is a free data retrieval call binding the contract method 0x2c78a525.
//
// Solidity: function MIN_REVEALS() view returns(uint256)
func (_AegisContract *AegisContractCallerSession) MINREVEALS() (*big.Int, error) {
	return _AegisContract.Contract.MINREVEALS(&_AegisContract.CallOpts)
}

// SLASHEXECUTORONFAIL is a free data retrieval call binding the contract method 0x3a5a3f28.
//
// Solidity: function SLASH_EXECUTOR_ON_FAIL() view returns(uint256)
func (_AegisContract *AegisContractCaller) SLASHEXECUTORONFAIL(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "SLASH_EXECUTOR_ON_FAIL")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SLASHEXECUTORONFAIL is a free data retrieval call binding the contract method 0x3a5a3f28.
//
// Solidity: function SLASH_EXECUTOR_ON_FAIL() view returns(uint256)
func (_AegisContract *AegisContractSession) SLASHEXECUTORONFAIL() (*big.Int, error) {
	return _AegisContract.Contract.SLASHEXECUTORONFAIL(&_AegisContract.CallOpts)
}

// SLASHEXECUTORONFAIL is a free data retrieval call binding the contract method 0x3a5a3f28.
//
// Solidity: function SLASH_EXECUTOR_ON_FAIL() view returns(uint256)
func (_AegisContract *AegisContractCallerSession) SLASHEXECUTORONFAIL() (*big.Int, error) {
	return _AegisContract.Contract.SLASHEXECUTORONFAIL(&_AegisContract.CallOpts)
}

// SLASHPERDISSENT is a free data retrieval call binding the contract method 0xf96ac30f.
//
// Solidity: function SLASH_PER_DISSENT() view returns(uint256)
func (_AegisContract *AegisContractCaller) SLASHPERDISSENT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "SLASH_PER_DISSENT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SLASHPERDISSENT is a free data retrieval call binding the contract method 0xf96ac30f.
//
// Solidity: function SLASH_PER_DISSENT() view returns(uint256)
func (_AegisContract *AegisContractSession) SLASHPERDISSENT() (*big.Int, error) {
	return _AegisContract.Contract.SLASHPERDISSENT(&_AegisContract.CallOpts)
}

// SLASHPERDISSENT is a free data retrieval call binding the contract method 0xf96ac30f.
//
// Solidity: function SLASH_PER_DISSENT() view returns(uint256)
func (_AegisContract *AegisContractCallerSession) SLASHPERDISSENT() (*big.Int, error) {
	return _AegisContract.Contract.SLASHPERDISSENT(&_AegisContract.CallOpts)
}

// Commits is a free data retrieval call binding the contract method 0x885d68a8.
//
// Solidity: function commits(uint256 , address ) view returns(bytes32)
func (_AegisContract *AegisContractCaller) Commits(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) ([32]byte, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "commits", arg0, arg1)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Commits is a free data retrieval call binding the contract method 0x885d68a8.
//
// Solidity: function commits(uint256 , address ) view returns(bytes32)
func (_AegisContract *AegisContractSession) Commits(arg0 *big.Int, arg1 common.Address) ([32]byte, error) {
	return _AegisContract.Contract.Commits(&_AegisContract.CallOpts, arg0, arg1)
}

// Commits is a free data retrieval call binding the contract method 0x885d68a8.
//
// Solidity: function commits(uint256 , address ) view returns(bytes32)
func (_AegisContract *AegisContractCallerSession) Commits(arg0 *big.Int, arg1 common.Address) ([32]byte, error) {
	return _AegisContract.Contract.Commits(&_AegisContract.CallOpts, arg0, arg1)
}

// JobRevealers is a free data retrieval call binding the contract method 0x6e66169f.
//
// Solidity: function jobRevealers(uint256 , uint256 ) view returns(address)
func (_AegisContract *AegisContractCaller) JobRevealers(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "jobRevealers", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// JobRevealers is a free data retrieval call binding the contract method 0x6e66169f.
//
// Solidity: function jobRevealers(uint256 , uint256 ) view returns(address)
func (_AegisContract *AegisContractSession) JobRevealers(arg0 *big.Int, arg1 *big.Int) (common.Address, error) {
	return _AegisContract.Contract.JobRevealers(&_AegisContract.CallOpts, arg0, arg1)
}

// JobRevealers is a free data retrieval call binding the contract method 0x6e66169f.
//
// Solidity: function jobRevealers(uint256 , uint256 ) view returns(address)
func (_AegisContract *AegisContractCallerSession) JobRevealers(arg0 *big.Int, arg1 *big.Int) (common.Address, error) {
	return _AegisContract.Contract.JobRevealers(&_AegisContract.CallOpts, arg0, arg1)
}

// Jobs is a free data retrieval call binding the contract method 0x180aedf3.
//
// Solidity: function jobs(uint256 ) view returns(address client, address executor, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty, bytes32 specHash, bytes32 txHash, bytes32 reportedOutcomeHash, uint64 claimDeadline, uint64 commitDeadline, uint64 revealDeadline, uint8 status)
func (_AegisContract *AegisContractCaller) Jobs(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Client                common.Address
	Executor              common.Address
	ExecutorReimbursement *big.Int
	ExecutorFee           *big.Int
	VerifierBounty        *big.Int
	SpecHash              [32]byte
	TxHash                [32]byte
	ReportedOutcomeHash   [32]byte
	ClaimDeadline         uint64
	CommitDeadline        uint64
	RevealDeadline        uint64
	Status                uint8
}, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "jobs", arg0)

	outstruct := new(struct {
		Client                common.Address
		Executor              common.Address
		ExecutorReimbursement *big.Int
		ExecutorFee           *big.Int
		VerifierBounty        *big.Int
		SpecHash              [32]byte
		TxHash                [32]byte
		ReportedOutcomeHash   [32]byte
		ClaimDeadline         uint64
		CommitDeadline        uint64
		RevealDeadline        uint64
		Status                uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Client = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Executor = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.ExecutorReimbursement = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.ExecutorFee = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifierBounty = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.SpecHash = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.TxHash = *abi.ConvertType(out[6], new([32]byte)).(*[32]byte)
	outstruct.ReportedOutcomeHash = *abi.ConvertType(out[7], new([32]byte)).(*[32]byte)
	outstruct.ClaimDeadline = *abi.ConvertType(out[8], new(uint64)).(*uint64)
	outstruct.CommitDeadline = *abi.ConvertType(out[9], new(uint64)).(*uint64)
	outstruct.RevealDeadline = *abi.ConvertType(out[10], new(uint64)).(*uint64)
	outstruct.Status = *abi.ConvertType(out[11], new(uint8)).(*uint8)

	return *outstruct, err

}

// Jobs is a free data retrieval call binding the contract method 0x180aedf3.
//
// Solidity: function jobs(uint256 ) view returns(address client, address executor, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty, bytes32 specHash, bytes32 txHash, bytes32 reportedOutcomeHash, uint64 claimDeadline, uint64 commitDeadline, uint64 revealDeadline, uint8 status)
func (_AegisContract *AegisContractSession) Jobs(arg0 *big.Int) (struct {
	Client                common.Address
	Executor              common.Address
	ExecutorReimbursement *big.Int
	ExecutorFee           *big.Int
	VerifierBounty        *big.Int
	SpecHash              [32]byte
	TxHash                [32]byte
	ReportedOutcomeHash   [32]byte
	ClaimDeadline         uint64
	CommitDeadline        uint64
	RevealDeadline        uint64
	Status                uint8
}, error) {
	return _AegisContract.Contract.Jobs(&_AegisContract.CallOpts, arg0)
}

// Jobs is a free data retrieval call binding the contract method 0x180aedf3.
//
// Solidity: function jobs(uint256 ) view returns(address client, address executor, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty, bytes32 specHash, bytes32 txHash, bytes32 reportedOutcomeHash, uint64 claimDeadline, uint64 commitDeadline, uint64 revealDeadline, uint8 status)
func (_AegisContract *AegisContractCallerSession) Jobs(arg0 *big.Int) (struct {
	Client                common.Address
	Executor              common.Address
	ExecutorReimbursement *big.Int
	ExecutorFee           *big.Int
	VerifierBounty        *big.Int
	SpecHash              [32]byte
	TxHash                [32]byte
	ReportedOutcomeHash   [32]byte
	ClaimDeadline         uint64
	CommitDeadline        uint64
	RevealDeadline        uint64
	Status                uint8
}, error) {
	return _AegisContract.Contract.Jobs(&_AegisContract.CallOpts, arg0)
}

// NextJobId is a free data retrieval call binding the contract method 0xb0c2aa5e.
//
// Solidity: function nextJobId() view returns(uint256)
func (_AegisContract *AegisContractCaller) NextJobId(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "nextJobId")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NextJobId is a free data retrieval call binding the contract method 0xb0c2aa5e.
//
// Solidity: function nextJobId() view returns(uint256)
func (_AegisContract *AegisContractSession) NextJobId() (*big.Int, error) {
	return _AegisContract.Contract.NextJobId(&_AegisContract.CallOpts)
}

// NextJobId is a free data retrieval call binding the contract method 0xb0c2aa5e.
//
// Solidity: function nextJobId() view returns(uint256)
func (_AegisContract *AegisContractCallerSession) NextJobId() (*big.Int, error) {
	return _AegisContract.Contract.NextJobId(&_AegisContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AegisContract *AegisContractCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AegisContract *AegisContractSession) Owner() (common.Address, error) {
	return _AegisContract.Contract.Owner(&_AegisContract.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AegisContract *AegisContractCallerSession) Owner() (common.Address, error) {
	return _AegisContract.Contract.Owner(&_AegisContract.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_AegisContract *AegisContractCaller) Registry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "registry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_AegisContract *AegisContractSession) Registry() (common.Address, error) {
	return _AegisContract.Contract.Registry(&_AegisContract.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_AegisContract *AegisContractCallerSession) Registry() (common.Address, error) {
	return _AegisContract.Contract.Registry(&_AegisContract.CallOpts)
}

// RevealCount is a free data retrieval call binding the contract method 0xedcf55a2.
//
// Solidity: function revealCount(uint256 jobId) view returns(uint256)
func (_AegisContract *AegisContractCaller) RevealCount(opts *bind.CallOpts, jobId *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "revealCount", jobId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RevealCount is a free data retrieval call binding the contract method 0xedcf55a2.
//
// Solidity: function revealCount(uint256 jobId) view returns(uint256)
func (_AegisContract *AegisContractSession) RevealCount(jobId *big.Int) (*big.Int, error) {
	return _AegisContract.Contract.RevealCount(&_AegisContract.CallOpts, jobId)
}

// RevealCount is a free data retrieval call binding the contract method 0xedcf55a2.
//
// Solidity: function revealCount(uint256 jobId) view returns(uint256)
func (_AegisContract *AegisContractCallerSession) RevealCount(jobId *big.Int) (*big.Int, error) {
	return _AegisContract.Contract.RevealCount(&_AegisContract.CallOpts, jobId)
}

// Reveals is a free data retrieval call binding the contract method 0x4ef14888.
//
// Solidity: function reveals(uint256 , address ) view returns(bool verdict, bool revealed)
func (_AegisContract *AegisContractCaller) Reveals(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (struct {
	Verdict  bool
	Revealed bool
}, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "reveals", arg0, arg1)

	outstruct := new(struct {
		Verdict  bool
		Revealed bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Verdict = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.Revealed = *abi.ConvertType(out[1], new(bool)).(*bool)

	return *outstruct, err

}

// Reveals is a free data retrieval call binding the contract method 0x4ef14888.
//
// Solidity: function reveals(uint256 , address ) view returns(bool verdict, bool revealed)
func (_AegisContract *AegisContractSession) Reveals(arg0 *big.Int, arg1 common.Address) (struct {
	Verdict  bool
	Revealed bool
}, error) {
	return _AegisContract.Contract.Reveals(&_AegisContract.CallOpts, arg0, arg1)
}

// Reveals is a free data retrieval call binding the contract method 0x4ef14888.
//
// Solidity: function reveals(uint256 , address ) view returns(bool verdict, bool revealed)
func (_AegisContract *AegisContractCallerSession) Reveals(arg0 *big.Int, arg1 common.Address) (struct {
	Verdict  bool
	Revealed bool
}, error) {
	return _AegisContract.Contract.Reveals(&_AegisContract.CallOpts, arg0, arg1)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_AegisContract *AegisContractCaller) Treasury(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "treasury")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_AegisContract *AegisContractSession) Treasury() (common.Address, error) {
	return _AegisContract.Contract.Treasury(&_AegisContract.CallOpts)
}

// Treasury is a free data retrieval call binding the contract method 0x61d027b3.
//
// Solidity: function treasury() view returns(address)
func (_AegisContract *AegisContractCallerSession) Treasury() (common.Address, error) {
	return _AegisContract.Contract.Treasury(&_AegisContract.CallOpts)
}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_AegisContract *AegisContractCaller) Usdc(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AegisContract.contract.Call(opts, &out, "usdc")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_AegisContract *AegisContractSession) Usdc() (common.Address, error) {
	return _AegisContract.Contract.Usdc(&_AegisContract.CallOpts)
}

// Usdc is a free data retrieval call binding the contract method 0x3e413bee.
//
// Solidity: function usdc() view returns(address)
func (_AegisContract *AegisContractCallerSession) Usdc() (common.Address, error) {
	return _AegisContract.Contract.Usdc(&_AegisContract.CallOpts)
}

// CancelStaleJob is a paid mutator transaction binding the contract method 0x9514bb75.
//
// Solidity: function cancelStaleJob(uint256 jobId) returns()
func (_AegisContract *AegisContractTransactor) CancelStaleJob(opts *bind.TransactOpts, jobId *big.Int) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "cancelStaleJob", jobId)
}

// CancelStaleJob is a paid mutator transaction binding the contract method 0x9514bb75.
//
// Solidity: function cancelStaleJob(uint256 jobId) returns()
func (_AegisContract *AegisContractSession) CancelStaleJob(jobId *big.Int) (*types.Transaction, error) {
	return _AegisContract.Contract.CancelStaleJob(&_AegisContract.TransactOpts, jobId)
}

// CancelStaleJob is a paid mutator transaction binding the contract method 0x9514bb75.
//
// Solidity: function cancelStaleJob(uint256 jobId) returns()
func (_AegisContract *AegisContractTransactorSession) CancelStaleJob(jobId *big.Int) (*types.Transaction, error) {
	return _AegisContract.Contract.CancelStaleJob(&_AegisContract.TransactOpts, jobId)
}

// CommitVote is a paid mutator transaction binding the contract method 0x9da69180.
//
// Solidity: function commitVote(uint256 jobId, bytes32 commitHash) returns()
func (_AegisContract *AegisContractTransactor) CommitVote(opts *bind.TransactOpts, jobId *big.Int, commitHash [32]byte) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "commitVote", jobId, commitHash)
}

// CommitVote is a paid mutator transaction binding the contract method 0x9da69180.
//
// Solidity: function commitVote(uint256 jobId, bytes32 commitHash) returns()
func (_AegisContract *AegisContractSession) CommitVote(jobId *big.Int, commitHash [32]byte) (*types.Transaction, error) {
	return _AegisContract.Contract.CommitVote(&_AegisContract.TransactOpts, jobId, commitHash)
}

// CommitVote is a paid mutator transaction binding the contract method 0x9da69180.
//
// Solidity: function commitVote(uint256 jobId, bytes32 commitHash) returns()
func (_AegisContract *AegisContractTransactorSession) CommitVote(jobId *big.Int, commitHash [32]byte) (*types.Transaction, error) {
	return _AegisContract.Contract.CommitVote(&_AegisContract.TransactOpts, jobId, commitHash)
}

// PostJob is a paid mutator transaction binding the contract method 0xc6d28ef8.
//
// Solidity: function postJob(bytes32 specHash, address executor, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty) returns(uint256 jobId)
func (_AegisContract *AegisContractTransactor) PostJob(opts *bind.TransactOpts, specHash [32]byte, executor common.Address, executorReimbursement *big.Int, executorFee *big.Int, verifierBounty *big.Int) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "postJob", specHash, executor, executorReimbursement, executorFee, verifierBounty)
}

// PostJob is a paid mutator transaction binding the contract method 0xc6d28ef8.
//
// Solidity: function postJob(bytes32 specHash, address executor, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty) returns(uint256 jobId)
func (_AegisContract *AegisContractSession) PostJob(specHash [32]byte, executor common.Address, executorReimbursement *big.Int, executorFee *big.Int, verifierBounty *big.Int) (*types.Transaction, error) {
	return _AegisContract.Contract.PostJob(&_AegisContract.TransactOpts, specHash, executor, executorReimbursement, executorFee, verifierBounty)
}

// PostJob is a paid mutator transaction binding the contract method 0xc6d28ef8.
//
// Solidity: function postJob(bytes32 specHash, address executor, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty) returns(uint256 jobId)
func (_AegisContract *AegisContractTransactorSession) PostJob(specHash [32]byte, executor common.Address, executorReimbursement *big.Int, executorFee *big.Int, verifierBounty *big.Int) (*types.Transaction, error) {
	return _AegisContract.Contract.PostJob(&_AegisContract.TransactOpts, specHash, executor, executorReimbursement, executorFee, verifierBounty)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AegisContract *AegisContractTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AegisContract *AegisContractSession) RenounceOwnership() (*types.Transaction, error) {
	return _AegisContract.Contract.RenounceOwnership(&_AegisContract.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AegisContract *AegisContractTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AegisContract.Contract.RenounceOwnership(&_AegisContract.TransactOpts)
}

// RevealVote is a paid mutator transaction binding the contract method 0xfacc092f.
//
// Solidity: function revealVote(uint256 jobId, bool verdict, bytes32 nonce) returns()
func (_AegisContract *AegisContractTransactor) RevealVote(opts *bind.TransactOpts, jobId *big.Int, verdict bool, nonce [32]byte) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "revealVote", jobId, verdict, nonce)
}

// RevealVote is a paid mutator transaction binding the contract method 0xfacc092f.
//
// Solidity: function revealVote(uint256 jobId, bool verdict, bytes32 nonce) returns()
func (_AegisContract *AegisContractSession) RevealVote(jobId *big.Int, verdict bool, nonce [32]byte) (*types.Transaction, error) {
	return _AegisContract.Contract.RevealVote(&_AegisContract.TransactOpts, jobId, verdict, nonce)
}

// RevealVote is a paid mutator transaction binding the contract method 0xfacc092f.
//
// Solidity: function revealVote(uint256 jobId, bool verdict, bytes32 nonce) returns()
func (_AegisContract *AegisContractTransactorSession) RevealVote(jobId *big.Int, verdict bool, nonce [32]byte) (*types.Transaction, error) {
	return _AegisContract.Contract.RevealVote(&_AegisContract.TransactOpts, jobId, verdict, nonce)
}

// Settle is a paid mutator transaction binding the contract method 0x8df82800.
//
// Solidity: function settle(uint256 jobId) returns()
func (_AegisContract *AegisContractTransactor) Settle(opts *bind.TransactOpts, jobId *big.Int) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "settle", jobId)
}

// Settle is a paid mutator transaction binding the contract method 0x8df82800.
//
// Solidity: function settle(uint256 jobId) returns()
func (_AegisContract *AegisContractSession) Settle(jobId *big.Int) (*types.Transaction, error) {
	return _AegisContract.Contract.Settle(&_AegisContract.TransactOpts, jobId)
}

// Settle is a paid mutator transaction binding the contract method 0x8df82800.
//
// Solidity: function settle(uint256 jobId) returns()
func (_AegisContract *AegisContractTransactorSession) Settle(jobId *big.Int) (*types.Transaction, error) {
	return _AegisContract.Contract.Settle(&_AegisContract.TransactOpts, jobId)
}

// SubmitClaim is a paid mutator transaction binding the contract method 0xd82460d2.
//
// Solidity: function submitClaim(uint256 jobId, bytes32 txHash, bytes32 reportedOutcomeHash) returns()
func (_AegisContract *AegisContractTransactor) SubmitClaim(opts *bind.TransactOpts, jobId *big.Int, txHash [32]byte, reportedOutcomeHash [32]byte) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "submitClaim", jobId, txHash, reportedOutcomeHash)
}

// SubmitClaim is a paid mutator transaction binding the contract method 0xd82460d2.
//
// Solidity: function submitClaim(uint256 jobId, bytes32 txHash, bytes32 reportedOutcomeHash) returns()
func (_AegisContract *AegisContractSession) SubmitClaim(jobId *big.Int, txHash [32]byte, reportedOutcomeHash [32]byte) (*types.Transaction, error) {
	return _AegisContract.Contract.SubmitClaim(&_AegisContract.TransactOpts, jobId, txHash, reportedOutcomeHash)
}

// SubmitClaim is a paid mutator transaction binding the contract method 0xd82460d2.
//
// Solidity: function submitClaim(uint256 jobId, bytes32 txHash, bytes32 reportedOutcomeHash) returns()
func (_AegisContract *AegisContractTransactorSession) SubmitClaim(jobId *big.Int, txHash [32]byte, reportedOutcomeHash [32]byte) (*types.Transaction, error) {
	return _AegisContract.Contract.SubmitClaim(&_AegisContract.TransactOpts, jobId, txHash, reportedOutcomeHash)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AegisContract *AegisContractTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AegisContract.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AegisContract *AegisContractSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AegisContract.Contract.TransferOwnership(&_AegisContract.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AegisContract *AegisContractTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AegisContract.Contract.TransferOwnership(&_AegisContract.TransactOpts, newOwner)
}

// AegisContractClaimSubmittedIterator is returned from FilterClaimSubmitted and is used to iterate over the raw logs and unpacked data for ClaimSubmitted events raised by the AegisContract contract.
type AegisContractClaimSubmittedIterator struct {
	Event *AegisContractClaimSubmitted // Event containing the contract specifics and raw log

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
func (it *AegisContractClaimSubmittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AegisContractClaimSubmitted)
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
		it.Event = new(AegisContractClaimSubmitted)
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
func (it *AegisContractClaimSubmittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AegisContractClaimSubmittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AegisContractClaimSubmitted represents a ClaimSubmitted event raised by the AegisContract contract.
type AegisContractClaimSubmitted struct {
	JobId               *big.Int
	TxHash              [32]byte
	ReportedOutcomeHash [32]byte
	Raw                 types.Log // Blockchain specific contextual infos
}

// FilterClaimSubmitted is a free log retrieval operation binding the contract event 0x3e9086131cd56a11cb2fc7a195ff02167e166171a0357a11f6ee490775b51119.
//
// Solidity: event ClaimSubmitted(uint256 indexed jobId, bytes32 txHash, bytes32 reportedOutcomeHash)
func (_AegisContract *AegisContractFilterer) FilterClaimSubmitted(opts *bind.FilterOpts, jobId []*big.Int) (*AegisContractClaimSubmittedIterator, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}

	logs, sub, err := _AegisContract.contract.FilterLogs(opts, "ClaimSubmitted", jobIdRule)
	if err != nil {
		return nil, err
	}
	return &AegisContractClaimSubmittedIterator{contract: _AegisContract.contract, event: "ClaimSubmitted", logs: logs, sub: sub}, nil
}

// WatchClaimSubmitted is a free log subscription operation binding the contract event 0x3e9086131cd56a11cb2fc7a195ff02167e166171a0357a11f6ee490775b51119.
//
// Solidity: event ClaimSubmitted(uint256 indexed jobId, bytes32 txHash, bytes32 reportedOutcomeHash)
func (_AegisContract *AegisContractFilterer) WatchClaimSubmitted(opts *bind.WatchOpts, sink chan<- *AegisContractClaimSubmitted, jobId []*big.Int) (event.Subscription, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}

	logs, sub, err := _AegisContract.contract.WatchLogs(opts, "ClaimSubmitted", jobIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AegisContractClaimSubmitted)
				if err := _AegisContract.contract.UnpackLog(event, "ClaimSubmitted", log); err != nil {
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

// ParseClaimSubmitted is a log parse operation binding the contract event 0x3e9086131cd56a11cb2fc7a195ff02167e166171a0357a11f6ee490775b51119.
//
// Solidity: event ClaimSubmitted(uint256 indexed jobId, bytes32 txHash, bytes32 reportedOutcomeHash)
func (_AegisContract *AegisContractFilterer) ParseClaimSubmitted(log types.Log) (*AegisContractClaimSubmitted, error) {
	event := new(AegisContractClaimSubmitted)
	if err := _AegisContract.contract.UnpackLog(event, "ClaimSubmitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AegisContractJobCancelledIterator is returned from FilterJobCancelled and is used to iterate over the raw logs and unpacked data for JobCancelled events raised by the AegisContract contract.
type AegisContractJobCancelledIterator struct {
	Event *AegisContractJobCancelled // Event containing the contract specifics and raw log

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
func (it *AegisContractJobCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AegisContractJobCancelled)
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
		it.Event = new(AegisContractJobCancelled)
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
func (it *AegisContractJobCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AegisContractJobCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AegisContractJobCancelled represents a JobCancelled event raised by the AegisContract contract.
type AegisContractJobCancelled struct {
	JobId *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterJobCancelled is a free log retrieval operation binding the contract event 0x58243f872c550ad02802199ade99eb4250f6bb19e9abbf2d3de7b969c0eb5e4f.
//
// Solidity: event JobCancelled(uint256 indexed jobId)
func (_AegisContract *AegisContractFilterer) FilterJobCancelled(opts *bind.FilterOpts, jobId []*big.Int) (*AegisContractJobCancelledIterator, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}

	logs, sub, err := _AegisContract.contract.FilterLogs(opts, "JobCancelled", jobIdRule)
	if err != nil {
		return nil, err
	}
	return &AegisContractJobCancelledIterator{contract: _AegisContract.contract, event: "JobCancelled", logs: logs, sub: sub}, nil
}

// WatchJobCancelled is a free log subscription operation binding the contract event 0x58243f872c550ad02802199ade99eb4250f6bb19e9abbf2d3de7b969c0eb5e4f.
//
// Solidity: event JobCancelled(uint256 indexed jobId)
func (_AegisContract *AegisContractFilterer) WatchJobCancelled(opts *bind.WatchOpts, sink chan<- *AegisContractJobCancelled, jobId []*big.Int) (event.Subscription, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}

	logs, sub, err := _AegisContract.contract.WatchLogs(opts, "JobCancelled", jobIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AegisContractJobCancelled)
				if err := _AegisContract.contract.UnpackLog(event, "JobCancelled", log); err != nil {
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

// ParseJobCancelled is a log parse operation binding the contract event 0x58243f872c550ad02802199ade99eb4250f6bb19e9abbf2d3de7b969c0eb5e4f.
//
// Solidity: event JobCancelled(uint256 indexed jobId)
func (_AegisContract *AegisContractFilterer) ParseJobCancelled(log types.Log) (*AegisContractJobCancelled, error) {
	event := new(AegisContractJobCancelled)
	if err := _AegisContract.contract.UnpackLog(event, "JobCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AegisContractJobPostedIterator is returned from FilterJobPosted and is used to iterate over the raw logs and unpacked data for JobPosted events raised by the AegisContract contract.
type AegisContractJobPostedIterator struct {
	Event *AegisContractJobPosted // Event containing the contract specifics and raw log

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
func (it *AegisContractJobPostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AegisContractJobPosted)
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
		it.Event = new(AegisContractJobPosted)
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
func (it *AegisContractJobPostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AegisContractJobPostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AegisContractJobPosted represents a JobPosted event raised by the AegisContract contract.
type AegisContractJobPosted struct {
	JobId                 *big.Int
	Client                common.Address
	Executor              common.Address
	SpecHash              [32]byte
	ExecutorReimbursement *big.Int
	ExecutorFee           *big.Int
	VerifierBounty        *big.Int
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterJobPosted is a free log retrieval operation binding the contract event 0x5ad6ccf1e6883768e3798bf9a601432abb56dd593d1f039a0bec3f3c0a2feed4.
//
// Solidity: event JobPosted(uint256 indexed jobId, address indexed client, address indexed executor, bytes32 specHash, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty)
func (_AegisContract *AegisContractFilterer) FilterJobPosted(opts *bind.FilterOpts, jobId []*big.Int, client []common.Address, executor []common.Address) (*AegisContractJobPostedIterator, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}
	var clientRule []interface{}
	for _, clientItem := range client {
		clientRule = append(clientRule, clientItem)
	}
	var executorRule []interface{}
	for _, executorItem := range executor {
		executorRule = append(executorRule, executorItem)
	}

	logs, sub, err := _AegisContract.contract.FilterLogs(opts, "JobPosted", jobIdRule, clientRule, executorRule)
	if err != nil {
		return nil, err
	}
	return &AegisContractJobPostedIterator{contract: _AegisContract.contract, event: "JobPosted", logs: logs, sub: sub}, nil
}

// WatchJobPosted is a free log subscription operation binding the contract event 0x5ad6ccf1e6883768e3798bf9a601432abb56dd593d1f039a0bec3f3c0a2feed4.
//
// Solidity: event JobPosted(uint256 indexed jobId, address indexed client, address indexed executor, bytes32 specHash, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty)
func (_AegisContract *AegisContractFilterer) WatchJobPosted(opts *bind.WatchOpts, sink chan<- *AegisContractJobPosted, jobId []*big.Int, client []common.Address, executor []common.Address) (event.Subscription, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}
	var clientRule []interface{}
	for _, clientItem := range client {
		clientRule = append(clientRule, clientItem)
	}
	var executorRule []interface{}
	for _, executorItem := range executor {
		executorRule = append(executorRule, executorItem)
	}

	logs, sub, err := _AegisContract.contract.WatchLogs(opts, "JobPosted", jobIdRule, clientRule, executorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AegisContractJobPosted)
				if err := _AegisContract.contract.UnpackLog(event, "JobPosted", log); err != nil {
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

// ParseJobPosted is a log parse operation binding the contract event 0x5ad6ccf1e6883768e3798bf9a601432abb56dd593d1f039a0bec3f3c0a2feed4.
//
// Solidity: event JobPosted(uint256 indexed jobId, address indexed client, address indexed executor, bytes32 specHash, uint256 executorReimbursement, uint256 executorFee, uint256 verifierBounty)
func (_AegisContract *AegisContractFilterer) ParseJobPosted(log types.Log) (*AegisContractJobPosted, error) {
	event := new(AegisContractJobPosted)
	if err := _AegisContract.contract.UnpackLog(event, "JobPosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AegisContractJobSettledIterator is returned from FilterJobSettled and is used to iterate over the raw logs and unpacked data for JobSettled events raised by the AegisContract contract.
type AegisContractJobSettledIterator struct {
	Event *AegisContractJobSettled // Event containing the contract specifics and raw log

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
func (it *AegisContractJobSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AegisContractJobSettled)
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
		it.Event = new(AegisContractJobSettled)
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
func (it *AegisContractJobSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AegisContractJobSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AegisContractJobSettled represents a JobSettled event raised by the AegisContract contract.
type AegisContractJobSettled struct {
	JobId        *big.Int
	FinalVerdict bool
	ForVotes     *big.Int
	TotalReveals *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterJobSettled is a free log retrieval operation binding the contract event 0xe9aad52ec0f73a7e38d0d8af379e6efd0ba77116533ce0d37e1931b4c9412ac3.
//
// Solidity: event JobSettled(uint256 indexed jobId, bool finalVerdict, uint256 forVotes, uint256 totalReveals)
func (_AegisContract *AegisContractFilterer) FilterJobSettled(opts *bind.FilterOpts, jobId []*big.Int) (*AegisContractJobSettledIterator, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}

	logs, sub, err := _AegisContract.contract.FilterLogs(opts, "JobSettled", jobIdRule)
	if err != nil {
		return nil, err
	}
	return &AegisContractJobSettledIterator{contract: _AegisContract.contract, event: "JobSettled", logs: logs, sub: sub}, nil
}

// WatchJobSettled is a free log subscription operation binding the contract event 0xe9aad52ec0f73a7e38d0d8af379e6efd0ba77116533ce0d37e1931b4c9412ac3.
//
// Solidity: event JobSettled(uint256 indexed jobId, bool finalVerdict, uint256 forVotes, uint256 totalReveals)
func (_AegisContract *AegisContractFilterer) WatchJobSettled(opts *bind.WatchOpts, sink chan<- *AegisContractJobSettled, jobId []*big.Int) (event.Subscription, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}

	logs, sub, err := _AegisContract.contract.WatchLogs(opts, "JobSettled", jobIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AegisContractJobSettled)
				if err := _AegisContract.contract.UnpackLog(event, "JobSettled", log); err != nil {
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

// ParseJobSettled is a log parse operation binding the contract event 0xe9aad52ec0f73a7e38d0d8af379e6efd0ba77116533ce0d37e1931b4c9412ac3.
//
// Solidity: event JobSettled(uint256 indexed jobId, bool finalVerdict, uint256 forVotes, uint256 totalReveals)
func (_AegisContract *AegisContractFilterer) ParseJobSettled(log types.Log) (*AegisContractJobSettled, error) {
	event := new(AegisContractJobSettled)
	if err := _AegisContract.contract.UnpackLog(event, "JobSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AegisContractOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AegisContract contract.
type AegisContractOwnershipTransferredIterator struct {
	Event *AegisContractOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AegisContractOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AegisContractOwnershipTransferred)
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
		it.Event = new(AegisContractOwnershipTransferred)
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
func (it *AegisContractOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AegisContractOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AegisContractOwnershipTransferred represents a OwnershipTransferred event raised by the AegisContract contract.
type AegisContractOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AegisContract *AegisContractFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AegisContractOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AegisContract.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AegisContractOwnershipTransferredIterator{contract: _AegisContract.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AegisContract *AegisContractFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AegisContractOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AegisContract.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AegisContractOwnershipTransferred)
				if err := _AegisContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_AegisContract *AegisContractFilterer) ParseOwnershipTransferred(log types.Log) (*AegisContractOwnershipTransferred, error) {
	event := new(AegisContractOwnershipTransferred)
	if err := _AegisContract.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AegisContractVoteCommittedIterator is returned from FilterVoteCommitted and is used to iterate over the raw logs and unpacked data for VoteCommitted events raised by the AegisContract contract.
type AegisContractVoteCommittedIterator struct {
	Event *AegisContractVoteCommitted // Event containing the contract specifics and raw log

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
func (it *AegisContractVoteCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AegisContractVoteCommitted)
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
		it.Event = new(AegisContractVoteCommitted)
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
func (it *AegisContractVoteCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AegisContractVoteCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AegisContractVoteCommitted represents a VoteCommitted event raised by the AegisContract contract.
type AegisContractVoteCommitted struct {
	JobId      *big.Int
	Verifier   common.Address
	CommitHash [32]byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterVoteCommitted is a free log retrieval operation binding the contract event 0x1f4e2aa7825ef02e85293e2de66cfbcb326af2632b558996e546bb1ef3e8764d.
//
// Solidity: event VoteCommitted(uint256 indexed jobId, address indexed verifier, bytes32 commitHash)
func (_AegisContract *AegisContractFilterer) FilterVoteCommitted(opts *bind.FilterOpts, jobId []*big.Int, verifier []common.Address) (*AegisContractVoteCommittedIterator, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}
	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _AegisContract.contract.FilterLogs(opts, "VoteCommitted", jobIdRule, verifierRule)
	if err != nil {
		return nil, err
	}
	return &AegisContractVoteCommittedIterator{contract: _AegisContract.contract, event: "VoteCommitted", logs: logs, sub: sub}, nil
}

// WatchVoteCommitted is a free log subscription operation binding the contract event 0x1f4e2aa7825ef02e85293e2de66cfbcb326af2632b558996e546bb1ef3e8764d.
//
// Solidity: event VoteCommitted(uint256 indexed jobId, address indexed verifier, bytes32 commitHash)
func (_AegisContract *AegisContractFilterer) WatchVoteCommitted(opts *bind.WatchOpts, sink chan<- *AegisContractVoteCommitted, jobId []*big.Int, verifier []common.Address) (event.Subscription, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}
	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _AegisContract.contract.WatchLogs(opts, "VoteCommitted", jobIdRule, verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AegisContractVoteCommitted)
				if err := _AegisContract.contract.UnpackLog(event, "VoteCommitted", log); err != nil {
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

// ParseVoteCommitted is a log parse operation binding the contract event 0x1f4e2aa7825ef02e85293e2de66cfbcb326af2632b558996e546bb1ef3e8764d.
//
// Solidity: event VoteCommitted(uint256 indexed jobId, address indexed verifier, bytes32 commitHash)
func (_AegisContract *AegisContractFilterer) ParseVoteCommitted(log types.Log) (*AegisContractVoteCommitted, error) {
	event := new(AegisContractVoteCommitted)
	if err := _AegisContract.contract.UnpackLog(event, "VoteCommitted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AegisContractVoteRevealedIterator is returned from FilterVoteRevealed and is used to iterate over the raw logs and unpacked data for VoteRevealed events raised by the AegisContract contract.
type AegisContractVoteRevealedIterator struct {
	Event *AegisContractVoteRevealed // Event containing the contract specifics and raw log

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
func (it *AegisContractVoteRevealedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AegisContractVoteRevealed)
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
		it.Event = new(AegisContractVoteRevealed)
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
func (it *AegisContractVoteRevealedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AegisContractVoteRevealedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AegisContractVoteRevealed represents a VoteRevealed event raised by the AegisContract contract.
type AegisContractVoteRevealed struct {
	JobId    *big.Int
	Verifier common.Address
	Verdict  bool
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterVoteRevealed is a free log retrieval operation binding the contract event 0x943b1e1e6ea36d3a013c9f482eb56e6fec8904129c7ed1869bcfe1f04727cd4c.
//
// Solidity: event VoteRevealed(uint256 indexed jobId, address indexed verifier, bool verdict)
func (_AegisContract *AegisContractFilterer) FilterVoteRevealed(opts *bind.FilterOpts, jobId []*big.Int, verifier []common.Address) (*AegisContractVoteRevealedIterator, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}
	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _AegisContract.contract.FilterLogs(opts, "VoteRevealed", jobIdRule, verifierRule)
	if err != nil {
		return nil, err
	}
	return &AegisContractVoteRevealedIterator{contract: _AegisContract.contract, event: "VoteRevealed", logs: logs, sub: sub}, nil
}

// WatchVoteRevealed is a free log subscription operation binding the contract event 0x943b1e1e6ea36d3a013c9f482eb56e6fec8904129c7ed1869bcfe1f04727cd4c.
//
// Solidity: event VoteRevealed(uint256 indexed jobId, address indexed verifier, bool verdict)
func (_AegisContract *AegisContractFilterer) WatchVoteRevealed(opts *bind.WatchOpts, sink chan<- *AegisContractVoteRevealed, jobId []*big.Int, verifier []common.Address) (event.Subscription, error) {

	var jobIdRule []interface{}
	for _, jobIdItem := range jobId {
		jobIdRule = append(jobIdRule, jobIdItem)
	}
	var verifierRule []interface{}
	for _, verifierItem := range verifier {
		verifierRule = append(verifierRule, verifierItem)
	}

	logs, sub, err := _AegisContract.contract.WatchLogs(opts, "VoteRevealed", jobIdRule, verifierRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AegisContractVoteRevealed)
				if err := _AegisContract.contract.UnpackLog(event, "VoteRevealed", log); err != nil {
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

// ParseVoteRevealed is a log parse operation binding the contract event 0x943b1e1e6ea36d3a013c9f482eb56e6fec8904129c7ed1869bcfe1f04727cd4c.
//
// Solidity: event VoteRevealed(uint256 indexed jobId, address indexed verifier, bool verdict)
func (_AegisContract *AegisContractFilterer) ParseVoteRevealed(log types.Log) (*AegisContractVoteRevealed, error) {
	event := new(AegisContractVoteRevealed)
	if err := _AegisContract.contract.UnpackLog(event, "VoteRevealed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
