// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

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

// CRSManagerMetaData contains all meta data concerning the CRSManager contract.
var CRSManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_roundDuration\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_maxParticipants\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"newCRS\",\"type\":\"bytes\"}],\"name\":\"CRSContributed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"Committed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"crs\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"participants\",\"type\":\"address[]\"}],\"name\":\"Finalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"}],\"name\":\"Registered\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ceremonyActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"commitDeadline\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"commitments\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"participant\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"submitted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"newCRS\",\"type\":\"bytes\"}],\"name\":\"contributeCRS\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"crsHistory\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"crs\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentCRS\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentContributorIdx\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"currentRound\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"finalizeCRS\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentCRS\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentContributorIdx\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLatestCRS\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRegisteredParticipants\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"maxParticipants\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"registeredParticipants\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"roundDuration\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"commitment\",\"type\":\"bytes32\"}],\"name\":\"submitCommitment\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x60806040523480156200001157600080fd5b5060405162002612380380620026128339818101604052810190620000379190620000ab565b81600181905550806002819055506001600081905550600154426200005d919062000121565b60038190555050506200015c565b600080fd5b6000819050919050565b620000858162000070565b81146200009157600080fd5b50565b600081519050620000a5816200007a565b92915050565b60008060408385031215620000c557620000c46200006b565b5b6000620000d58582860162000094565b9250506020620000e88582860162000094565b9150509250929050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60006200012e8262000070565b91506200013b8362000070565b9250828201905080821115620001565762000155620000f2565b5b92915050565b6124a6806200016c6000396000f3fe608060405234801561001057600080fd5b50600436106101165760003560e01c806380734445116100a2578063ab520b3d11610071578063ab520b3d146102af578063b615f1e9146102cf578063caa1a46e146102ed578063e90d4c06146102f7578063f7cb789a1461032957610116565b806380734445146102375780638a19c8bc14610255578063a2ff020914610273578063a7b23ec61461029157610116565b806338448dd1116100e957806338448dd11461017f578063388ae83c146101b157806353f3eb8f146101e1578063699e6509146101fd57806378b0a7471461021957610116565b806316664c751461011b5780631aa3a008146101395780631d1904881461014357806324924bf714610161575b600080fd5b610123610347565b60405161013091906112a0565b60405180910390f35b61014161034d565b005b61014b6105c5565b60405161015891906112a0565b60405180910390f35b6101696105cf565b60405161017691906112a0565b60405180910390f35b610199600480360381019061019491906112f1565b6105d5565b6040516101a8939291906113ae565b60405180910390f35b6101cb60048036038101906101c691906113ec565b610687565b6040516101d8919061146d565b60405180910390f35b6101fb60048036038101906101f691906114be565b6106d5565b005b61021760048036038101906102129190611550565b610976565b005b610221610b9c565b60405161022e919061159d565b60405180910390f35b61023f610c2e565b60405161024c91906115da565b60405180910390f35b61025d610c41565b60405161026a91906112a0565b60405180910390f35b61027b610c47565b60405161028891906112a0565b60405180910390f35b610299610c4d565b6040516102a691906116b3565b60405180910390f35b6102b7610ced565b6040516102c6939291906116d5565b60405180910390f35b6102d7610e45565b6040516102e4919061159d565b60405180910390f35b6102f5610ed3565b005b610311600480360381019061030c9190611746565b6111c0565b60405161032093929190611795565b60405180910390f35b610331611224565b60405161033e91906112a0565b60405180910390f35b60085481565b6003544210610391576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161038890611829565b60405180910390fd5b600254600660008054815260200190815260200160002080549050106103ec576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103e390611895565b60405180910390fd5b60006006600080548152602001908152602001600020905060005b81805490508110156104d1573373ffffffffffffffffffffffffffffffffffffffff1682828154811061043d5761043c6118b5565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16036104be576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104b590611930565b60405180910390fd5b80806104c99061197f565b915050610407565b5080339080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff160217905550600181805490500361057257600060088190555060076000610556919061122a565b6001600960006101000a81548160ff0219169083151502179055505b3373ffffffffffffffffffffffffffffffffffffffff167f6f3bf3fa84e4763a43b3d23f9d79be242d6d5c834941ff4c1111b67469e1150c6000546040516105ba91906112a0565b60405180910390a250565b6000600854905090565b60025481565b60046020528060005260406000206000915090508060000154908060010180546105fe906119f6565b80601f016020809104026020016040519081016040528092919081815260200182805461062a906119f6565b80156106775780601f1061064c57610100808354040283529160200191610677565b820191906000526020600020905b81548152906001019060200180831161065a57829003601f168201915b5050505050908060020154905083565b600660205281600052604060002081815481106106a357600080fd5b906000526020600020016000915091509054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6003544210610719576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161071090611a73565b60405180910390fd5b6000600660008054815260200190815260200160002090506000805b82805490508110156107cd573373ffffffffffffffffffffffffffffffffffffffff1683828154811061076b5761076a6118b5565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16036107ba57600191506107cd565b80806107c59061197f565b915050610735565b508061080e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161080590611adf565b60405180910390fd5b6000600560008054815260200190815260200160002060003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060020160009054906101000a900460ff16156108b5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108ac90611b4b565b60405180910390fd5b338160000160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555083816001018190555060018160020160006101000a81548160ff0219169083151502179055503373ffffffffffffffffffffffffffffffffffffffff167f9dbfefa56f7a6ab8ace30e3e1d85b9df477517c5bbc4df07885d78b71b05519a60005486604051610968929190611b6b565b60405180910390a250505050565b600960009054906101000a900460ff166109c5576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109bc90611be0565b60405180910390fd5b6000600660008054815260200190815260200160002090506000818054905011610a24576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a1b90611c4c565b60405180910390fd5b808054905060085410610a6c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a6390611cb8565b60405180910390fd5b3373ffffffffffffffffffffffffffffffffffffffff168160085481548110610a9857610a976118b5565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614610b19576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b1090611d24565b60405180910390fd5b828260079182610b2a929190611f2a565b503373ffffffffffffffffffffffffffffffffffffffff167f5121fb45dedd7f35c1bafe81e1ab9743ecaa6e1c6ec5aec515266ac82f3ac7d86000548585604051610b7793929190612036565b60405180910390a260086000815480929190610b929061197f565b9190505550505050565b606060078054610bab906119f6565b80601f0160208091040260200160405190810160405280929190818152602001828054610bd7906119f6565b8015610c245780601f10610bf957610100808354040283529160200191610c24565b820191906000526020600020905b815481529060010190602001808311610c0757829003601f168201915b5050505050905090565b600960009054906101000a900460ff1681565b60005481565b60035481565b60606006600080548152602001908152602001600020805480602002602001604051908101604052809291908181526020018280548015610ce357602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610c99575b5050505050905090565b6060600060606000600460006001600054610d089190612068565b8152602001908152602001600020905080600101816002015482600301828054610d31906119f6565b80601f0160208091040260200160405190810160405280929190818152602001828054610d5d906119f6565b8015610daa5780601f10610d7f57610100808354040283529160200191610daa565b820191906000526020600020905b815481529060010190602001808311610d8d57829003601f168201915b5050505050925080805480602002602001604051908101604052809291908181526020018280548015610e3257602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019060010190808311610de8575b5050505050905093509350935050909192565b60078054610e52906119f6565b80601f0160208091040260200160405190810160405280929190818152602001828054610e7e906119f6565b8015610ecb5780601f10610ea057610100808354040283529160200191610ecb565b820191906000526020600020905b815481529060010190602001808311610eae57829003601f168201915b505050505081565b600960009054906101000a900460ff16610f22576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610f1990611be0565b60405180910390fd5b6000600660008054815260200190815260200160002090506000818054905011610f81576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610f7890611c4c565b60405180910390fd5b808054905060085414610fc9576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610fc0906120e8565b60405180910390fd5b6000600460008054815260200190815260200160002090506000816001018054610ff2906119f6565b905014611034576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161102b90612154565b60405180910390fd5b60005481600001819055506007816001019081611051919061219f565b5042816002018190555060005b828054905081101561111f5781600301838281548110611081576110806118b5565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555080806111179061197f565b91505061105e565b507f64ebe481729253d9df6960d56e1c4ec3f051ab146ae51e274360e25efcbf4aaa600054600784604051611156939291906123f7565b60405180910390a16001600080828254611170919061243c565b9250508190555060015442611185919061243c565b6003819055506000600960006101000a81548160ff021916908315150217905550600760006111b4919061122a565b60006008819055505050565b6005602052816000526040600020602052806000526040600020600091509150508060000160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16908060010154908060020160009054906101000a900460ff16905083565b60015481565b508054611236906119f6565b6000825580601f106112485750611267565b601f016020900490600052602060002090810190611266919061126a565b5b50565b5b8082111561128357600081600090555060010161126b565b5090565b6000819050919050565b61129a81611287565b82525050565b60006020820190506112b56000830184611291565b92915050565b600080fd5b600080fd5b6112ce81611287565b81146112d957600080fd5b50565b6000813590506112eb816112c5565b92915050565b600060208284031215611307576113066112bb565b5b6000611315848285016112dc565b91505092915050565b600081519050919050565b600082825260208201905092915050565b60005b8381101561135857808201518184015260208101905061133d565b60008484015250505050565b6000601f19601f8301169050919050565b60006113808261131e565b61138a8185611329565b935061139a81856020860161133a565b6113a381611364565b840191505092915050565b60006060820190506113c36000830186611291565b81810360208301526113d58185611375565b90506113e46040830184611291565b949350505050565b60008060408385031215611403576114026112bb565b5b6000611411858286016112dc565b9250506020611422858286016112dc565b9150509250929050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b60006114578261142c565b9050919050565b6114678161144c565b82525050565b6000602082019050611482600083018461145e565b92915050565b6000819050919050565b61149b81611488565b81146114a657600080fd5b50565b6000813590506114b881611492565b92915050565b6000602082840312156114d4576114d36112bb565b5b60006114e2848285016114a9565b91505092915050565b600080fd5b600080fd5b600080fd5b60008083601f8401126115105761150f6114eb565b5b8235905067ffffffffffffffff81111561152d5761152c6114f0565b5b602083019150836001820283011115611549576115486114f5565b5b9250929050565b60008060208385031215611567576115666112bb565b5b600083013567ffffffffffffffff811115611585576115846112c0565b5b611591858286016114fa565b92509250509250929050565b600060208201905081810360008301526115b78184611375565b905092915050565b60008115159050919050565b6115d4816115bf565b82525050565b60006020820190506115ef60008301846115cb565b92915050565b600081519050919050565b600082825260208201905092915050565b6000819050602082019050919050565b61162a8161144c565b82525050565b600061163c8383611621565b60208301905092915050565b6000602082019050919050565b6000611660826115f5565b61166a8185611600565b935061167583611611565b8060005b838110156116a657815161168d8882611630565b975061169883611648565b925050600181019050611679565b5085935050505092915050565b600060208201905081810360008301526116cd8184611655565b905092915050565b600060608201905081810360008301526116ef8186611375565b90506116fe6020830185611291565b81810360408301526117108184611655565b9050949350505050565b6117238161144c565b811461172e57600080fd5b50565b6000813590506117408161171a565b92915050565b6000806040838503121561175d5761175c6112bb565b5b600061176b858286016112dc565b925050602061177c85828601611731565b9150509250929050565b61178f81611488565b82525050565b60006060820190506117aa600083018661145e565b6117b76020830185611786565b6117c460408301846115cb565b949350505050565b600082825260208201905092915050565b7f526567697374726174696f6e20636c6f73656400000000000000000000000000600082015250565b60006118136013836117cc565b915061181e826117dd565b602082019050919050565b6000602082019050818103600083015261184281611806565b9050919050565b7f4d6178207061727469636970616e747320726561636865640000000000000000600082015250565b600061187f6018836117cc565b915061188a82611849565b602082019050919050565b600060208201905081810360008301526118ae81611872565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f416c726561647920726567697374657265640000000000000000000000000000600082015250565b600061191a6012836117cc565b9150611925826118e4565b602082019050919050565b600060208201905081810360008301526119498161190d565b9050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600061198a82611287565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036119bc576119bb611950565b5b600182019050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b60006002820490506001821680611a0e57607f821691505b602082108103611a2157611a206119c7565b5b50919050565b7f436f6d6d69746d656e7420706861736520656e64656400000000000000000000600082015250565b6000611a5d6016836117cc565b9150611a6882611a27565b602082019050919050565b60006020820190508181036000830152611a8c81611a50565b9050919050565b7f4e6f7420612072656769737465726564207061727469636970616e7400000000600082015250565b6000611ac9601c836117cc565b9150611ad482611a93565b602082019050919050565b60006020820190508181036000830152611af881611abc565b9050919050565b7f416c7265616479207375626d6974746564000000000000000000000000000000600082015250565b6000611b356011836117cc565b9150611b4082611aff565b602082019050919050565b60006020820190508181036000830152611b6481611b28565b9050919050565b6000604082019050611b806000830185611291565b611b8d6020830184611786565b9392505050565b7f4e6f2061637469766520636572656d6f6e790000000000000000000000000000600082015250565b6000611bca6012836117cc565b9150611bd582611b94565b602082019050919050565b60006020820190508181036000830152611bf981611bbd565b9050919050565b7f4e6f207061727469636970616e74730000000000000000000000000000000000600082015250565b6000611c36600f836117cc565b9150611c4182611c00565b602082019050919050565b60006020820190508181036000830152611c6581611c29565b9050919050565b7f416c6c206861766520636f6e7472696275746564000000000000000000000000600082015250565b6000611ca26014836117cc565b9150611cad82611c6c565b602082019050919050565b60006020820190508181036000830152611cd181611c95565b9050919050565b7f4e6f7420796f7572207475726e00000000000000000000000000000000000000600082015250565b6000611d0e600d836117cc565b9150611d1982611cd8565b602082019050919050565b60006020820190508181036000830152611d3d81611d01565b9050919050565b600082905092915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b60008190508160005260206000209050919050565b60006020601f8301049050919050565b600082821b905092915050565b600060088302611de07fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82611da3565b611dea8683611da3565b95508019841693508086168417925050509392505050565b6000819050919050565b6000611e27611e22611e1d84611287565b611e02565b611287565b9050919050565b6000819050919050565b611e4183611e0c565b611e55611e4d82611e2e565b848454611db0565b825550505050565b600090565b611e6a611e5d565b611e75818484611e38565b505050565b5b81811015611e9957611e8e600082611e62565b600181019050611e7b565b5050565b601f821115611ede57611eaf81611d7e565b611eb884611d93565b81016020851015611ec7578190505b611edb611ed385611d93565b830182611e7a565b50505b505050565b600082821c905092915050565b6000611f0160001984600802611ee3565b1980831691505092915050565b6000611f1a8383611ef0565b9150826002028217905092915050565b611f348383611d44565b67ffffffffffffffff811115611f4d57611f4c611d4f565b5b611f5782546119f6565b611f62828285611e9d565b6000601f831160018114611f915760008415611f7f578287013590505b611f898582611f0e565b865550611ff1565b601f198416611f9f86611d7e565b60005b82811015611fc757848901358255600182019150602085019450602081019050611fa2565b86831015611fe45784890135611fe0601f891682611ef0565b8355505b6001600288020188555050505b50505050505050565b82818337600083830152505050565b60006120158385611329565b9350612022838584611ffa565b61202b83611364565b840190509392505050565b600060408201905061204b6000830186611291565b818103602083015261205e818486612009565b9050949350505050565b600061207382611287565b915061207e83611287565b925082820390508181111561209657612095611950565b5b92915050565b7f4e6f7420616c6c206861766520636f6e74726962757465640000000000000000600082015250565b60006120d26018836117cc565b91506120dd8261209c565b602082019050919050565b60006020820190508181036000830152612101816120c5565b9050919050565b7f416c72656164792066696e616c697a6564000000000000000000000000000000600082015250565b600061213e6011836117cc565b915061214982612108565b602082019050919050565b6000602082019050818103600083015261216d81612131565b9050919050565b600081549050612183816119f6565b9050919050565b60008190508160005260206000209050919050565b8181036121ad575050612285565b6121b682612174565b67ffffffffffffffff8111156121cf576121ce611d4f565b5b6121d982546119f6565b6121e4828285611e9d565b6000601f8311600181146122135760008415612201578287015490505b61220b8582611f0e565b86555061227e565b601f1984166122218761218a565b965061222c86611d7e565b60005b828110156122545784890154825560018201915060018501945060208101905061222f565b86831015612271578489015461226d601f891682611ef0565b8355505b6001600288020188555050505b5050505050505b565b60008154612294816119f6565b61229e8186611329565b945060018216600081146122b957600181146122cf57612302565b60ff198316865281151560200286019350612302565b6122d885611d7e565b60005b838110156122fa578154818901526001820191506020810190506122db565b808801955050505b50505092915050565b600081549050919050565b60008190508160005260206000209050919050565b60008160001c9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b600061236b6123668361232b565b612338565b9050919050565b600061237e8254612358565b9050919050565b6000600182019050919050565b600061239d8261230b565b6123a78185611600565b93506123b283612316565b8060005b838110156123ea576123c782612372565b6123d18882611630565b97506123dc83612385565b9250506001810190506123b6565b5085935050505092915050565b600060608201905061240c6000830186611291565b818103602083015261241e8185612287565b905081810360408301526124328184612392565b9050949350505050565b600061244782611287565b915061245283611287565b925082820190508082111561246a57612469611950565b5b9291505056fea264697066735822122007d31ecba64930c697b10c4a6607b3d471c48cc68203ddda005eeb515a67acfb64736f6c63430008130033",
}

// CRSManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use CRSManagerMetaData.ABI instead.
var CRSManagerABI = CRSManagerMetaData.ABI

// CRSManagerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use CRSManagerMetaData.Bin instead.
var CRSManagerBin = CRSManagerMetaData.Bin

// DeployCRSManager deploys a new Ethereum contract, binding an instance of CRSManager to it.
func DeployCRSManager(auth *bind.TransactOpts, backend bind.ContractBackend, _roundDuration *big.Int, _maxParticipants *big.Int) (common.Address, *types.Transaction, *CRSManager, error) {
	parsed, err := CRSManagerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(CRSManagerBin), backend, _roundDuration, _maxParticipants)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CRSManager{CRSManagerCaller: CRSManagerCaller{contract: contract}, CRSManagerTransactor: CRSManagerTransactor{contract: contract}, CRSManagerFilterer: CRSManagerFilterer{contract: contract}}, nil
}

// CRSManager is an auto generated Go binding around an Ethereum contract.
type CRSManager struct {
	CRSManagerCaller     // Read-only binding to the contract
	CRSManagerTransactor // Write-only binding to the contract
	CRSManagerFilterer   // Log filterer for contract events
}

// CRSManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type CRSManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CRSManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CRSManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CRSManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CRSManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CRSManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CRSManagerSession struct {
	Contract     *CRSManager       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CRSManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CRSManagerCallerSession struct {
	Contract *CRSManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// CRSManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CRSManagerTransactorSession struct {
	Contract     *CRSManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// CRSManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type CRSManagerRaw struct {
	Contract *CRSManager // Generic contract binding to access the raw methods on
}

// CRSManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CRSManagerCallerRaw struct {
	Contract *CRSManagerCaller // Generic read-only contract binding to access the raw methods on
}

// CRSManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CRSManagerTransactorRaw struct {
	Contract *CRSManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCRSManager creates a new instance of CRSManager, bound to a specific deployed contract.
func NewCRSManager(address common.Address, backend bind.ContractBackend) (*CRSManager, error) {
	contract, err := bindCRSManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CRSManager{CRSManagerCaller: CRSManagerCaller{contract: contract}, CRSManagerTransactor: CRSManagerTransactor{contract: contract}, CRSManagerFilterer: CRSManagerFilterer{contract: contract}}, nil
}

// NewCRSManagerCaller creates a new read-only instance of CRSManager, bound to a specific deployed contract.
func NewCRSManagerCaller(address common.Address, caller bind.ContractCaller) (*CRSManagerCaller, error) {
	contract, err := bindCRSManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CRSManagerCaller{contract: contract}, nil
}

// NewCRSManagerTransactor creates a new write-only instance of CRSManager, bound to a specific deployed contract.
func NewCRSManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*CRSManagerTransactor, error) {
	contract, err := bindCRSManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CRSManagerTransactor{contract: contract}, nil
}

// NewCRSManagerFilterer creates a new log filterer instance of CRSManager, bound to a specific deployed contract.
func NewCRSManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*CRSManagerFilterer, error) {
	contract, err := bindCRSManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CRSManagerFilterer{contract: contract}, nil
}

// bindCRSManager binds a generic wrapper to an already deployed contract.
func bindCRSManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CRSManagerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CRSManager *CRSManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CRSManager.Contract.CRSManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CRSManager *CRSManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CRSManager.Contract.CRSManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CRSManager *CRSManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CRSManager.Contract.CRSManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CRSManager *CRSManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CRSManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CRSManager *CRSManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CRSManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CRSManager *CRSManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CRSManager.Contract.contract.Transact(opts, method, params...)
}

// CeremonyActive is a free data retrieval call binding the contract method 0x80734445.
//
// Solidity: function ceremonyActive() view returns(bool)
func (_CRSManager *CRSManagerCaller) CeremonyActive(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "ceremonyActive")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// CeremonyActive is a free data retrieval call binding the contract method 0x80734445.
//
// Solidity: function ceremonyActive() view returns(bool)
func (_CRSManager *CRSManagerSession) CeremonyActive() (bool, error) {
	return _CRSManager.Contract.CeremonyActive(&_CRSManager.CallOpts)
}

// CeremonyActive is a free data retrieval call binding the contract method 0x80734445.
//
// Solidity: function ceremonyActive() view returns(bool)
func (_CRSManager *CRSManagerCallerSession) CeremonyActive() (bool, error) {
	return _CRSManager.Contract.CeremonyActive(&_CRSManager.CallOpts)
}

// CommitDeadline is a free data retrieval call binding the contract method 0xa2ff0209.
//
// Solidity: function commitDeadline() view returns(uint256)
func (_CRSManager *CRSManagerCaller) CommitDeadline(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "commitDeadline")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CommitDeadline is a free data retrieval call binding the contract method 0xa2ff0209.
//
// Solidity: function commitDeadline() view returns(uint256)
func (_CRSManager *CRSManagerSession) CommitDeadline() (*big.Int, error) {
	return _CRSManager.Contract.CommitDeadline(&_CRSManager.CallOpts)
}

// CommitDeadline is a free data retrieval call binding the contract method 0xa2ff0209.
//
// Solidity: function commitDeadline() view returns(uint256)
func (_CRSManager *CRSManagerCallerSession) CommitDeadline() (*big.Int, error) {
	return _CRSManager.Contract.CommitDeadline(&_CRSManager.CallOpts)
}

// Commitments is a free data retrieval call binding the contract method 0xe90d4c06.
//
// Solidity: function commitments(uint256 , address ) view returns(address participant, bytes32 commitment, bool submitted)
func (_CRSManager *CRSManagerCaller) Commitments(opts *bind.CallOpts, arg0 *big.Int, arg1 common.Address) (struct {
	Participant common.Address
	Commitment  [32]byte
	Submitted   bool
}, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "commitments", arg0, arg1)

	outstruct := new(struct {
		Participant common.Address
		Commitment  [32]byte
		Submitted   bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Participant = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.Commitment = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.Submitted = *abi.ConvertType(out[2], new(bool)).(*bool)

	return *outstruct, err

}

// Commitments is a free data retrieval call binding the contract method 0xe90d4c06.
//
// Solidity: function commitments(uint256 , address ) view returns(address participant, bytes32 commitment, bool submitted)
func (_CRSManager *CRSManagerSession) Commitments(arg0 *big.Int, arg1 common.Address) (struct {
	Participant common.Address
	Commitment  [32]byte
	Submitted   bool
}, error) {
	return _CRSManager.Contract.Commitments(&_CRSManager.CallOpts, arg0, arg1)
}

// Commitments is a free data retrieval call binding the contract method 0xe90d4c06.
//
// Solidity: function commitments(uint256 , address ) view returns(address participant, bytes32 commitment, bool submitted)
func (_CRSManager *CRSManagerCallerSession) Commitments(arg0 *big.Int, arg1 common.Address) (struct {
	Participant common.Address
	Commitment  [32]byte
	Submitted   bool
}, error) {
	return _CRSManager.Contract.Commitments(&_CRSManager.CallOpts, arg0, arg1)
}

// CrsHistory is a free data retrieval call binding the contract method 0x38448dd1.
//
// Solidity: function crsHistory(uint256 ) view returns(uint256 round, bytes crs, uint256 timestamp)
func (_CRSManager *CRSManagerCaller) CrsHistory(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Round     *big.Int
	Crs       []byte
	Timestamp *big.Int
}, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "crsHistory", arg0)

	outstruct := new(struct {
		Round     *big.Int
		Crs       []byte
		Timestamp *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Round = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Crs = *abi.ConvertType(out[1], new([]byte)).(*[]byte)
	outstruct.Timestamp = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// CrsHistory is a free data retrieval call binding the contract method 0x38448dd1.
//
// Solidity: function crsHistory(uint256 ) view returns(uint256 round, bytes crs, uint256 timestamp)
func (_CRSManager *CRSManagerSession) CrsHistory(arg0 *big.Int) (struct {
	Round     *big.Int
	Crs       []byte
	Timestamp *big.Int
}, error) {
	return _CRSManager.Contract.CrsHistory(&_CRSManager.CallOpts, arg0)
}

// CrsHistory is a free data retrieval call binding the contract method 0x38448dd1.
//
// Solidity: function crsHistory(uint256 ) view returns(uint256 round, bytes crs, uint256 timestamp)
func (_CRSManager *CRSManagerCallerSession) CrsHistory(arg0 *big.Int) (struct {
	Round     *big.Int
	Crs       []byte
	Timestamp *big.Int
}, error) {
	return _CRSManager.Contract.CrsHistory(&_CRSManager.CallOpts, arg0)
}

// CurrentCRS is a free data retrieval call binding the contract method 0xb615f1e9.
//
// Solidity: function currentCRS() view returns(bytes)
func (_CRSManager *CRSManagerCaller) CurrentCRS(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "currentCRS")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// CurrentCRS is a free data retrieval call binding the contract method 0xb615f1e9.
//
// Solidity: function currentCRS() view returns(bytes)
func (_CRSManager *CRSManagerSession) CurrentCRS() ([]byte, error) {
	return _CRSManager.Contract.CurrentCRS(&_CRSManager.CallOpts)
}

// CurrentCRS is a free data retrieval call binding the contract method 0xb615f1e9.
//
// Solidity: function currentCRS() view returns(bytes)
func (_CRSManager *CRSManagerCallerSession) CurrentCRS() ([]byte, error) {
	return _CRSManager.Contract.CurrentCRS(&_CRSManager.CallOpts)
}

// CurrentContributorIdx is a free data retrieval call binding the contract method 0x16664c75.
//
// Solidity: function currentContributorIdx() view returns(uint256)
func (_CRSManager *CRSManagerCaller) CurrentContributorIdx(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "currentContributorIdx")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentContributorIdx is a free data retrieval call binding the contract method 0x16664c75.
//
// Solidity: function currentContributorIdx() view returns(uint256)
func (_CRSManager *CRSManagerSession) CurrentContributorIdx() (*big.Int, error) {
	return _CRSManager.Contract.CurrentContributorIdx(&_CRSManager.CallOpts)
}

// CurrentContributorIdx is a free data retrieval call binding the contract method 0x16664c75.
//
// Solidity: function currentContributorIdx() view returns(uint256)
func (_CRSManager *CRSManagerCallerSession) CurrentContributorIdx() (*big.Int, error) {
	return _CRSManager.Contract.CurrentContributorIdx(&_CRSManager.CallOpts)
}

// CurrentRound is a free data retrieval call binding the contract method 0x8a19c8bc.
//
// Solidity: function currentRound() view returns(uint256)
func (_CRSManager *CRSManagerCaller) CurrentRound(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "currentRound")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CurrentRound is a free data retrieval call binding the contract method 0x8a19c8bc.
//
// Solidity: function currentRound() view returns(uint256)
func (_CRSManager *CRSManagerSession) CurrentRound() (*big.Int, error) {
	return _CRSManager.Contract.CurrentRound(&_CRSManager.CallOpts)
}

// CurrentRound is a free data retrieval call binding the contract method 0x8a19c8bc.
//
// Solidity: function currentRound() view returns(uint256)
func (_CRSManager *CRSManagerCallerSession) CurrentRound() (*big.Int, error) {
	return _CRSManager.Contract.CurrentRound(&_CRSManager.CallOpts)
}

// GetCurrentCRS is a free data retrieval call binding the contract method 0x78b0a747.
//
// Solidity: function getCurrentCRS() view returns(bytes)
func (_CRSManager *CRSManagerCaller) GetCurrentCRS(opts *bind.CallOpts) ([]byte, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "getCurrentCRS")

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// GetCurrentCRS is a free data retrieval call binding the contract method 0x78b0a747.
//
// Solidity: function getCurrentCRS() view returns(bytes)
func (_CRSManager *CRSManagerSession) GetCurrentCRS() ([]byte, error) {
	return _CRSManager.Contract.GetCurrentCRS(&_CRSManager.CallOpts)
}

// GetCurrentCRS is a free data retrieval call binding the contract method 0x78b0a747.
//
// Solidity: function getCurrentCRS() view returns(bytes)
func (_CRSManager *CRSManagerCallerSession) GetCurrentCRS() ([]byte, error) {
	return _CRSManager.Contract.GetCurrentCRS(&_CRSManager.CallOpts)
}

// GetCurrentContributorIdx is a free data retrieval call binding the contract method 0x1d190488.
//
// Solidity: function getCurrentContributorIdx() view returns(uint256)
func (_CRSManager *CRSManagerCaller) GetCurrentContributorIdx(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "getCurrentContributorIdx")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCurrentContributorIdx is a free data retrieval call binding the contract method 0x1d190488.
//
// Solidity: function getCurrentContributorIdx() view returns(uint256)
func (_CRSManager *CRSManagerSession) GetCurrentContributorIdx() (*big.Int, error) {
	return _CRSManager.Contract.GetCurrentContributorIdx(&_CRSManager.CallOpts)
}

// GetCurrentContributorIdx is a free data retrieval call binding the contract method 0x1d190488.
//
// Solidity: function getCurrentContributorIdx() view returns(uint256)
func (_CRSManager *CRSManagerCallerSession) GetCurrentContributorIdx() (*big.Int, error) {
	return _CRSManager.Contract.GetCurrentContributorIdx(&_CRSManager.CallOpts)
}

// GetLatestCRS is a free data retrieval call binding the contract method 0xab520b3d.
//
// Solidity: function getLatestCRS() view returns(bytes, uint256, address[])
func (_CRSManager *CRSManagerCaller) GetLatestCRS(opts *bind.CallOpts) ([]byte, *big.Int, []common.Address, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "getLatestCRS")

	if err != nil {
		return *new([]byte), *new(*big.Int), *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	out2 := *abi.ConvertType(out[2], new([]common.Address)).(*[]common.Address)

	return out0, out1, out2, err

}

// GetLatestCRS is a free data retrieval call binding the contract method 0xab520b3d.
//
// Solidity: function getLatestCRS() view returns(bytes, uint256, address[])
func (_CRSManager *CRSManagerSession) GetLatestCRS() ([]byte, *big.Int, []common.Address, error) {
	return _CRSManager.Contract.GetLatestCRS(&_CRSManager.CallOpts)
}

// GetLatestCRS is a free data retrieval call binding the contract method 0xab520b3d.
//
// Solidity: function getLatestCRS() view returns(bytes, uint256, address[])
func (_CRSManager *CRSManagerCallerSession) GetLatestCRS() ([]byte, *big.Int, []common.Address, error) {
	return _CRSManager.Contract.GetLatestCRS(&_CRSManager.CallOpts)
}

// GetRegisteredParticipants is a free data retrieval call binding the contract method 0xa7b23ec6.
//
// Solidity: function getRegisteredParticipants() view returns(address[])
func (_CRSManager *CRSManagerCaller) GetRegisteredParticipants(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "getRegisteredParticipants")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetRegisteredParticipants is a free data retrieval call binding the contract method 0xa7b23ec6.
//
// Solidity: function getRegisteredParticipants() view returns(address[])
func (_CRSManager *CRSManagerSession) GetRegisteredParticipants() ([]common.Address, error) {
	return _CRSManager.Contract.GetRegisteredParticipants(&_CRSManager.CallOpts)
}

// GetRegisteredParticipants is a free data retrieval call binding the contract method 0xa7b23ec6.
//
// Solidity: function getRegisteredParticipants() view returns(address[])
func (_CRSManager *CRSManagerCallerSession) GetRegisteredParticipants() ([]common.Address, error) {
	return _CRSManager.Contract.GetRegisteredParticipants(&_CRSManager.CallOpts)
}

// MaxParticipants is a free data retrieval call binding the contract method 0x24924bf7.
//
// Solidity: function maxParticipants() view returns(uint256)
func (_CRSManager *CRSManagerCaller) MaxParticipants(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "maxParticipants")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MaxParticipants is a free data retrieval call binding the contract method 0x24924bf7.
//
// Solidity: function maxParticipants() view returns(uint256)
func (_CRSManager *CRSManagerSession) MaxParticipants() (*big.Int, error) {
	return _CRSManager.Contract.MaxParticipants(&_CRSManager.CallOpts)
}

// MaxParticipants is a free data retrieval call binding the contract method 0x24924bf7.
//
// Solidity: function maxParticipants() view returns(uint256)
func (_CRSManager *CRSManagerCallerSession) MaxParticipants() (*big.Int, error) {
	return _CRSManager.Contract.MaxParticipants(&_CRSManager.CallOpts)
}

// RegisteredParticipants is a free data retrieval call binding the contract method 0x388ae83c.
//
// Solidity: function registeredParticipants(uint256 , uint256 ) view returns(address)
func (_CRSManager *CRSManagerCaller) RegisteredParticipants(opts *bind.CallOpts, arg0 *big.Int, arg1 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "registeredParticipants", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RegisteredParticipants is a free data retrieval call binding the contract method 0x388ae83c.
//
// Solidity: function registeredParticipants(uint256 , uint256 ) view returns(address)
func (_CRSManager *CRSManagerSession) RegisteredParticipants(arg0 *big.Int, arg1 *big.Int) (common.Address, error) {
	return _CRSManager.Contract.RegisteredParticipants(&_CRSManager.CallOpts, arg0, arg1)
}

// RegisteredParticipants is a free data retrieval call binding the contract method 0x388ae83c.
//
// Solidity: function registeredParticipants(uint256 , uint256 ) view returns(address)
func (_CRSManager *CRSManagerCallerSession) RegisteredParticipants(arg0 *big.Int, arg1 *big.Int) (common.Address, error) {
	return _CRSManager.Contract.RegisteredParticipants(&_CRSManager.CallOpts, arg0, arg1)
}

// RoundDuration is a free data retrieval call binding the contract method 0xf7cb789a.
//
// Solidity: function roundDuration() view returns(uint256)
func (_CRSManager *CRSManagerCaller) RoundDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CRSManager.contract.Call(opts, &out, "roundDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// RoundDuration is a free data retrieval call binding the contract method 0xf7cb789a.
//
// Solidity: function roundDuration() view returns(uint256)
func (_CRSManager *CRSManagerSession) RoundDuration() (*big.Int, error) {
	return _CRSManager.Contract.RoundDuration(&_CRSManager.CallOpts)
}

// RoundDuration is a free data retrieval call binding the contract method 0xf7cb789a.
//
// Solidity: function roundDuration() view returns(uint256)
func (_CRSManager *CRSManagerCallerSession) RoundDuration() (*big.Int, error) {
	return _CRSManager.Contract.RoundDuration(&_CRSManager.CallOpts)
}

// ContributeCRS is a paid mutator transaction binding the contract method 0x699e6509.
//
// Solidity: function contributeCRS(bytes newCRS) returns()
func (_CRSManager *CRSManagerTransactor) ContributeCRS(opts *bind.TransactOpts, newCRS []byte) (*types.Transaction, error) {
	return _CRSManager.contract.Transact(opts, "contributeCRS", newCRS)
}

// ContributeCRS is a paid mutator transaction binding the contract method 0x699e6509.
//
// Solidity: function contributeCRS(bytes newCRS) returns()
func (_CRSManager *CRSManagerSession) ContributeCRS(newCRS []byte) (*types.Transaction, error) {
	return _CRSManager.Contract.ContributeCRS(&_CRSManager.TransactOpts, newCRS)
}

// ContributeCRS is a paid mutator transaction binding the contract method 0x699e6509.
//
// Solidity: function contributeCRS(bytes newCRS) returns()
func (_CRSManager *CRSManagerTransactorSession) ContributeCRS(newCRS []byte) (*types.Transaction, error) {
	return _CRSManager.Contract.ContributeCRS(&_CRSManager.TransactOpts, newCRS)
}

// FinalizeCRS is a paid mutator transaction binding the contract method 0xcaa1a46e.
//
// Solidity: function finalizeCRS() returns()
func (_CRSManager *CRSManagerTransactor) FinalizeCRS(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CRSManager.contract.Transact(opts, "finalizeCRS")
}

// FinalizeCRS is a paid mutator transaction binding the contract method 0xcaa1a46e.
//
// Solidity: function finalizeCRS() returns()
func (_CRSManager *CRSManagerSession) FinalizeCRS() (*types.Transaction, error) {
	return _CRSManager.Contract.FinalizeCRS(&_CRSManager.TransactOpts)
}

// FinalizeCRS is a paid mutator transaction binding the contract method 0xcaa1a46e.
//
// Solidity: function finalizeCRS() returns()
func (_CRSManager *CRSManagerTransactorSession) FinalizeCRS() (*types.Transaction, error) {
	return _CRSManager.Contract.FinalizeCRS(&_CRSManager.TransactOpts)
}

// Register is a paid mutator transaction binding the contract method 0x1aa3a008.
//
// Solidity: function register() returns()
func (_CRSManager *CRSManagerTransactor) Register(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CRSManager.contract.Transact(opts, "register")
}

// Register is a paid mutator transaction binding the contract method 0x1aa3a008.
//
// Solidity: function register() returns()
func (_CRSManager *CRSManagerSession) Register() (*types.Transaction, error) {
	return _CRSManager.Contract.Register(&_CRSManager.TransactOpts)
}

// Register is a paid mutator transaction binding the contract method 0x1aa3a008.
//
// Solidity: function register() returns()
func (_CRSManager *CRSManagerTransactorSession) Register() (*types.Transaction, error) {
	return _CRSManager.Contract.Register(&_CRSManager.TransactOpts)
}

// SubmitCommitment is a paid mutator transaction binding the contract method 0x53f3eb8f.
//
// Solidity: function submitCommitment(bytes32 commitment) returns()
func (_CRSManager *CRSManagerTransactor) SubmitCommitment(opts *bind.TransactOpts, commitment [32]byte) (*types.Transaction, error) {
	return _CRSManager.contract.Transact(opts, "submitCommitment", commitment)
}

// SubmitCommitment is a paid mutator transaction binding the contract method 0x53f3eb8f.
//
// Solidity: function submitCommitment(bytes32 commitment) returns()
func (_CRSManager *CRSManagerSession) SubmitCommitment(commitment [32]byte) (*types.Transaction, error) {
	return _CRSManager.Contract.SubmitCommitment(&_CRSManager.TransactOpts, commitment)
}

// SubmitCommitment is a paid mutator transaction binding the contract method 0x53f3eb8f.
//
// Solidity: function submitCommitment(bytes32 commitment) returns()
func (_CRSManager *CRSManagerTransactorSession) SubmitCommitment(commitment [32]byte) (*types.Transaction, error) {
	return _CRSManager.Contract.SubmitCommitment(&_CRSManager.TransactOpts, commitment)
}

// CRSManagerCRSContributedIterator is returned from FilterCRSContributed and is used to iterate over the raw logs and unpacked data for CRSContributed events raised by the CRSManager contract.
type CRSManagerCRSContributedIterator struct {
	Event *CRSManagerCRSContributed // Event containing the contract specifics and raw log

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
func (it *CRSManagerCRSContributedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CRSManagerCRSContributed)
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
		it.Event = new(CRSManagerCRSContributed)
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
func (it *CRSManagerCRSContributedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CRSManagerCRSContributedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CRSManagerCRSContributed represents a CRSContributed event raised by the CRSManager contract.
type CRSManagerCRSContributed struct {
	Participant common.Address
	Round       *big.Int
	NewCRS      []byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterCRSContributed is a free log retrieval operation binding the contract event 0x5121fb45dedd7f35c1bafe81e1ab9743ecaa6e1c6ec5aec515266ac82f3ac7d8.
//
// Solidity: event CRSContributed(address indexed participant, uint256 round, bytes newCRS)
func (_CRSManager *CRSManagerFilterer) FilterCRSContributed(opts *bind.FilterOpts, participant []common.Address) (*CRSManagerCRSContributedIterator, error) {

	var participantRule []interface{}
	for _, participantItem := range participant {
		participantRule = append(participantRule, participantItem)
	}

	logs, sub, err := _CRSManager.contract.FilterLogs(opts, "CRSContributed", participantRule)
	if err != nil {
		return nil, err
	}
	return &CRSManagerCRSContributedIterator{contract: _CRSManager.contract, event: "CRSContributed", logs: logs, sub: sub}, nil
}

// WatchCRSContributed is a free log subscription operation binding the contract event 0x5121fb45dedd7f35c1bafe81e1ab9743ecaa6e1c6ec5aec515266ac82f3ac7d8.
//
// Solidity: event CRSContributed(address indexed participant, uint256 round, bytes newCRS)
func (_CRSManager *CRSManagerFilterer) WatchCRSContributed(opts *bind.WatchOpts, sink chan<- *CRSManagerCRSContributed, participant []common.Address) (event.Subscription, error) {

	var participantRule []interface{}
	for _, participantItem := range participant {
		participantRule = append(participantRule, participantItem)
	}

	logs, sub, err := _CRSManager.contract.WatchLogs(opts, "CRSContributed", participantRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CRSManagerCRSContributed)
				if err := _CRSManager.contract.UnpackLog(event, "CRSContributed", log); err != nil {
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

// ParseCRSContributed is a log parse operation binding the contract event 0x5121fb45dedd7f35c1bafe81e1ab9743ecaa6e1c6ec5aec515266ac82f3ac7d8.
//
// Solidity: event CRSContributed(address indexed participant, uint256 round, bytes newCRS)
func (_CRSManager *CRSManagerFilterer) ParseCRSContributed(log types.Log) (*CRSManagerCRSContributed, error) {
	event := new(CRSManagerCRSContributed)
	if err := _CRSManager.contract.UnpackLog(event, "CRSContributed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CRSManagerCommittedIterator is returned from FilterCommitted and is used to iterate over the raw logs and unpacked data for Committed events raised by the CRSManager contract.
type CRSManagerCommittedIterator struct {
	Event *CRSManagerCommitted // Event containing the contract specifics and raw log

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
func (it *CRSManagerCommittedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CRSManagerCommitted)
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
		it.Event = new(CRSManagerCommitted)
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
func (it *CRSManagerCommittedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CRSManagerCommittedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CRSManagerCommitted represents a Committed event raised by the CRSManager contract.
type CRSManagerCommitted struct {
	Participant common.Address
	Round       *big.Int
	Commitment  [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterCommitted is a free log retrieval operation binding the contract event 0x9dbfefa56f7a6ab8ace30e3e1d85b9df477517c5bbc4df07885d78b71b05519a.
//
// Solidity: event Committed(address indexed participant, uint256 round, bytes32 commitment)
func (_CRSManager *CRSManagerFilterer) FilterCommitted(opts *bind.FilterOpts, participant []common.Address) (*CRSManagerCommittedIterator, error) {

	var participantRule []interface{}
	for _, participantItem := range participant {
		participantRule = append(participantRule, participantItem)
	}

	logs, sub, err := _CRSManager.contract.FilterLogs(opts, "Committed", participantRule)
	if err != nil {
		return nil, err
	}
	return &CRSManagerCommittedIterator{contract: _CRSManager.contract, event: "Committed", logs: logs, sub: sub}, nil
}

// WatchCommitted is a free log subscription operation binding the contract event 0x9dbfefa56f7a6ab8ace30e3e1d85b9df477517c5bbc4df07885d78b71b05519a.
//
// Solidity: event Committed(address indexed participant, uint256 round, bytes32 commitment)
func (_CRSManager *CRSManagerFilterer) WatchCommitted(opts *bind.WatchOpts, sink chan<- *CRSManagerCommitted, participant []common.Address) (event.Subscription, error) {

	var participantRule []interface{}
	for _, participantItem := range participant {
		participantRule = append(participantRule, participantItem)
	}

	logs, sub, err := _CRSManager.contract.WatchLogs(opts, "Committed", participantRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CRSManagerCommitted)
				if err := _CRSManager.contract.UnpackLog(event, "Committed", log); err != nil {
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

// ParseCommitted is a log parse operation binding the contract event 0x9dbfefa56f7a6ab8ace30e3e1d85b9df477517c5bbc4df07885d78b71b05519a.
//
// Solidity: event Committed(address indexed participant, uint256 round, bytes32 commitment)
func (_CRSManager *CRSManagerFilterer) ParseCommitted(log types.Log) (*CRSManagerCommitted, error) {
	event := new(CRSManagerCommitted)
	if err := _CRSManager.contract.UnpackLog(event, "Committed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CRSManagerFinalizedIterator is returned from FilterFinalized and is used to iterate over the raw logs and unpacked data for Finalized events raised by the CRSManager contract.
type CRSManagerFinalizedIterator struct {
	Event *CRSManagerFinalized // Event containing the contract specifics and raw log

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
func (it *CRSManagerFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CRSManagerFinalized)
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
		it.Event = new(CRSManagerFinalized)
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
func (it *CRSManagerFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CRSManagerFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CRSManagerFinalized represents a Finalized event raised by the CRSManager contract.
type CRSManagerFinalized struct {
	Round        *big.Int
	Crs          []byte
	Participants []common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterFinalized is a free log retrieval operation binding the contract event 0x64ebe481729253d9df6960d56e1c4ec3f051ab146ae51e274360e25efcbf4aaa.
//
// Solidity: event Finalized(uint256 round, bytes crs, address[] participants)
func (_CRSManager *CRSManagerFilterer) FilterFinalized(opts *bind.FilterOpts) (*CRSManagerFinalizedIterator, error) {

	logs, sub, err := _CRSManager.contract.FilterLogs(opts, "Finalized")
	if err != nil {
		return nil, err
	}
	return &CRSManagerFinalizedIterator{contract: _CRSManager.contract, event: "Finalized", logs: logs, sub: sub}, nil
}

// WatchFinalized is a free log subscription operation binding the contract event 0x64ebe481729253d9df6960d56e1c4ec3f051ab146ae51e274360e25efcbf4aaa.
//
// Solidity: event Finalized(uint256 round, bytes crs, address[] participants)
func (_CRSManager *CRSManagerFilterer) WatchFinalized(opts *bind.WatchOpts, sink chan<- *CRSManagerFinalized) (event.Subscription, error) {

	logs, sub, err := _CRSManager.contract.WatchLogs(opts, "Finalized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CRSManagerFinalized)
				if err := _CRSManager.contract.UnpackLog(event, "Finalized", log); err != nil {
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

// ParseFinalized is a log parse operation binding the contract event 0x64ebe481729253d9df6960d56e1c4ec3f051ab146ae51e274360e25efcbf4aaa.
//
// Solidity: event Finalized(uint256 round, bytes crs, address[] participants)
func (_CRSManager *CRSManagerFilterer) ParseFinalized(log types.Log) (*CRSManagerFinalized, error) {
	event := new(CRSManagerFinalized)
	if err := _CRSManager.contract.UnpackLog(event, "Finalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CRSManagerRegisteredIterator is returned from FilterRegistered and is used to iterate over the raw logs and unpacked data for Registered events raised by the CRSManager contract.
type CRSManagerRegisteredIterator struct {
	Event *CRSManagerRegistered // Event containing the contract specifics and raw log

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
func (it *CRSManagerRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CRSManagerRegistered)
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
		it.Event = new(CRSManagerRegistered)
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
func (it *CRSManagerRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CRSManagerRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CRSManagerRegistered represents a Registered event raised by the CRSManager contract.
type CRSManagerRegistered struct {
	Participant common.Address
	Round       *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRegistered is a free log retrieval operation binding the contract event 0x6f3bf3fa84e4763a43b3d23f9d79be242d6d5c834941ff4c1111b67469e1150c.
//
// Solidity: event Registered(address indexed participant, uint256 round)
func (_CRSManager *CRSManagerFilterer) FilterRegistered(opts *bind.FilterOpts, participant []common.Address) (*CRSManagerRegisteredIterator, error) {

	var participantRule []interface{}
	for _, participantItem := range participant {
		participantRule = append(participantRule, participantItem)
	}

	logs, sub, err := _CRSManager.contract.FilterLogs(opts, "Registered", participantRule)
	if err != nil {
		return nil, err
	}
	return &CRSManagerRegisteredIterator{contract: _CRSManager.contract, event: "Registered", logs: logs, sub: sub}, nil
}

// WatchRegistered is a free log subscription operation binding the contract event 0x6f3bf3fa84e4763a43b3d23f9d79be242d6d5c834941ff4c1111b67469e1150c.
//
// Solidity: event Registered(address indexed participant, uint256 round)
func (_CRSManager *CRSManagerFilterer) WatchRegistered(opts *bind.WatchOpts, sink chan<- *CRSManagerRegistered, participant []common.Address) (event.Subscription, error) {

	var participantRule []interface{}
	for _, participantItem := range participant {
		participantRule = append(participantRule, participantItem)
	}

	logs, sub, err := _CRSManager.contract.WatchLogs(opts, "Registered", participantRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CRSManagerRegistered)
				if err := _CRSManager.contract.UnpackLog(event, "Registered", log); err != nil {
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

// ParseRegistered is a log parse operation binding the contract event 0x6f3bf3fa84e4763a43b3d23f9d79be242d6d5c834941ff4c1111b67469e1150c.
//
// Solidity: event Registered(address indexed participant, uint256 round)
func (_CRSManager *CRSManagerFilterer) ParseRegistered(log types.Log) (*CRSManagerRegistered, error) {
	event := new(CRSManagerRegistered)
	if err := _CRSManager.contract.UnpackLog(event, "Registered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
