package l1

import (
	"context"
	"math/big"

	"zkrollup/contracts/bindings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type CRSManager struct {
	address  common.Address
	client   bind.ContractBackend
	caller   *bindings.CRSManagerCaller
	transact *bindings.CRSManagerTransactor
}

// NewCRSManager creates a new CRSManager client
func NewCRSManager(address common.Address, client bind.ContractBackend) (*CRSManager, error) {
	caller, err := bindings.NewCRSManagerCaller(address, client)
	if err != nil {
		return nil, err
	}
	transact, err := bindings.NewCRSManagerTransactor(address, client)
	if err != nil {
		return nil, err
	}
	return &CRSManager{
		address:  address,
		client:   client,
		caller:   caller,
		transact: transact,
	}, nil
}

// Register registers the caller for the current CRS round
func (c *CRSManager) Register(auth *bind.TransactOpts) error {
	_, err := c.transact.Register(auth)
	return err
}

// SubmitCommitment submits a commitment for the current round
func (c *CRSManager) SubmitCommitment(auth *bind.TransactOpts, commitment [32]byte) error {
	_, err := c.transact.SubmitCommitment(auth, commitment)
	return err
}

// FinalizeCRS finalizes the round by submitting the CRS
func (c *CRSManager) FinalizeCRS(auth *bind.TransactOpts) error {
	_, err := c.transact.FinalizeCRS(auth)
	return err
}

// GetLatestCRS fetches the latest CRS, timestamp, and participants
func (c *CRSManager) GetLatestCRS(ctx context.Context) ([]byte, *big.Int, []common.Address, error) {
	crs, ts, participants, err := c.caller.GetLatestCRS(&bind.CallOpts{Context: ctx})
	return crs, ts, participants, err
}

// GetRegisteredParticipants fetches the current round's registered participants
func (c *CRSManager) GetRegisteredParticipants(ctx context.Context) ([]common.Address, error) {
	return c.caller.GetRegisteredParticipants(&bind.CallOpts{Context: ctx})
}

// CheckTurn checks if the given address is the current contributor
func (c *CRSManager) CheckTurn(ctx context.Context, addr common.Address) (bool, error) {
	// Fetch current contributor index
	currentIdx, err := c.caller.GetCurrentContributorIdx(&bind.CallOpts{Context: ctx})
	if err != nil {
		return false, err
	}
	// Fetch registered participants
	participants, err := c.caller.GetRegisteredParticipants(&bind.CallOpts{Context: ctx})
	if err != nil {
		return false, err
	}
	if int(currentIdx.Int64()) >= len(participants) {
		return false, nil
	}
	return participants[int(currentIdx.Int64())] == addr, nil
}

// ContributeCRS submits a CRS contribution
func (c *CRSManager) ContributeCRS(auth *bind.TransactOpts, crs []byte) error {
	_, err := c.transact.ContributeCRS(auth, crs)
	return err
}

// IsLastContributor checks if addr is the last registered participant
func (c *CRSManager) IsLastContributor(ctx context.Context, addr common.Address) (bool, error) {
	participants, err := c.GetRegisteredParticipants(ctx)
	if err != nil {
		return false, err
	}
	if len(participants) == 0 {
		return false, nil
	}
	return participants[len(participants)-1] == addr, nil
}

// GetCurrentCRS fetches the current in-progress CRS
func (c *CRSManager) GetCurrentCRS(ctx context.Context) ([]byte, error) {
	crs, err := c.caller.GetCurrentCRS(&bind.CallOpts{Context: ctx})
	return crs, err
}
