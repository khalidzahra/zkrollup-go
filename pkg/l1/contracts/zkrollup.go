// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ZKRollupABI is the input ABI used to generate the binding from.
const ZKRollupABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"batchNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"BatchSubmitted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"batchNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"bool\",\"name\":\"verified\",\"type\":\"bool\"}],\"name\":\"BatchVerified\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"batches\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"verified\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentBatchNumber\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNumber\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"stateRoot\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32[]\",\"name\":\"txHashes\",\"type\":\"bytes32[]\"},{\"internalType\":\"bytes\",\"name\":\"proof\",\"type\":\"bytes\"}],\"name\":\"submitBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"batchNumber\",\"type\":\"uint256\"}],\"name\":\"verifyBatch\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// ZKRollup is an auto generated Go binding around an Ethereum contract.
type ZKRollup struct {
	ZKRollupCaller     // Read-only binding to the contract
	ZKRollupTransactor // Write-only binding to the contract
	ZKRollupFilterer   // Log filterer for contract events
}

// ZKRollupCaller is an auto generated read-only Go binding around an Ethereum contract.
type ZKRollupCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZKRollupTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ZKRollupTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ZKRollupFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ZKRollupFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// NewZKRollup creates a new instance of ZKRollup, bound to a specific deployed contract.
func NewZKRollup(address common.Address, backend bind.ContractBackend) (*ZKRollup, error) {
	contract, err := bindZKRollup(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ZKRollup{ZKRollupCaller: ZKRollupCaller{contract: contract}, ZKRollupTransactor: ZKRollupTransactor{contract: contract}, ZKRollupFilterer: ZKRollupFilterer{contract: contract}}, nil
}

// bindZKRollup binds a generic wrapper to an already deployed contract.
func bindZKRollup(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ZKRollupABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// SubmitBatch is a paid mutator transaction binding the contract method 0x8f1d3776.
func (_ZKRollup *ZKRollupTransactor) SubmitBatch(opts *bind.TransactOpts, batchNumber *big.Int, stateRoot [32]byte, txHashes [][32]byte, proof []byte) (*types.Transaction, error) {
	return _ZKRollup.contract.Transact(opts, "submitBatch", batchNumber, stateRoot, txHashes, proof)
}

// VerifyBatch is a free data retrieval call binding the contract method 0x5e8a791d.
func (_ZKRollup *ZKRollupCaller) VerifyBatch(opts *bind.CallOpts, batchNumber *big.Int) (bool, error) {
	var out []interface{}
	err := _ZKRollup.contract.Call(opts, &out, "verifyBatch", batchNumber)
	if err != nil {
		return false, err
	}
	return out[0].(bool), err
}

// DeployZKRollup deploys a new Ethereum contract, binding an instance of ZKRollup to it.
func DeployZKRollup(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ZKRollup, error) {
	parsed, err := abi.JSON(strings.NewReader(ZKRollupABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex("608060405234801561001057600080fd5b5061057f806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80635e8a791d146100515780638f1d3776146100815780639fa6a6e3146100a1578063c5b1d9aa146100bf575b600080fd5b61006b60048036038101906100669190610341565b6100dd565b60405161007891906103a3565b60405180910390f35b61009b600480360381019061009691906103be565b610132565b005b6100a9610261565b6040516100b6919061046a565b60405180910390f35b6100d960048036038101906100d49190610341565b610267565b005b60006001600083815260200190815260200160002060010160009054906101000a900460ff169050919050565b6000805490506000811161017a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161017190610502565b60405180910390fd5b60008111801561018c5750600081115b6101cb576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016101c290610502565b60405180910390fd5b60006001600085815260200190815260200160002060010160009054906101000a900460ff1690506101fd85858585610267565b7f5a978f4f6ea5d4a5a575f1e07f74f7e89a1367e9e09d3f7e9c93237c10c2d46b8583604051610230929190610522565b60405180910390a1505050505050565b60005481565b6040518060600160405280838152602001600115158152602001428152506001600085815260200190815260200160002060008201518160000155602082015181600101600a81111561028f577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b60ff1660ff16815260200160408201518160020155905050600160008082825461029a9190610551565b925050819055507f33a88a5e8eeccf39bfcf2b7574032a7564fd793bbaa33371cdb9e5f2f9aef9f2848360405161032f929190610522565b60405180910390a150505050565b60008135905061034b81610565565b92915050565b60006020828403121561036357600080fd5b60006103718482850161033c565b91505092915050565b61038381610485565b82525050565b61039281610491565b82525050565b60006020820190506103ad6000830184610389565b92915050565b600080600080608085870312156103d457600080fd5b60006103e28782880161033c565b94505060206103f38782880161033c565b935050604061040487828801610341565b925050606061041587828801610341565b91505092959194509250565b61042a8161049d565b82525050565b6000610449601c836104a7565b9150610454826104b8565b602082019050919050565b61046881610491565b82525050565b60006020820190506104836000830184610421565b92915050565b60008115159050919050565b6000819050919050565b6000819050919050565b600082825260208201905092915050565b7f496e76616c696420626174636820636f6e66696775726174696f6e000000000060008201525060006104e3826104a7565b91506104ef836104b8565b602082019050919050565b6000602082019050818103600083015261051381610430565b9050919050565b600060408201905061052f600083018561045f565b61053c6020830184610421565b9392505050565b6000819050919050565b600061055c82610542565b915061056783610542565b9250828201905080821115610565576105646104d6565b5b92915050565b600081905091905056fea26469706673582212209a9b9c9a9b9c9a9b9c9a9b9c9a9b9c9a9b9c9a9b9c9a9b9c9a9b9c9a9b9c9a9b64736f6c63430008070033"), nil)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ZKRollup{ZKRollupCaller: ZKRollupCaller{contract: contract}, ZKRollupTransactor: ZKRollupTransactor{contract: contract}, ZKRollupFilterer: ZKRollupFilterer{contract: contract}}, nil
}
