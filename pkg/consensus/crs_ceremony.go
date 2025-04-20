package consensus

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"sync"
	"bytes"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// CRSCeremonyMessage is sent between sequencers during the CRS ceremony
// Each step is signed by the contributor for auditability
// (Signature logic to be implemented as needed)
type CRSCeremonyMessage struct {
	EpochNumber     int64    // CRS epoch this ceremony is for
	Step            int      // Index of the sequencer in the participant list
	IntermediateCRS [][]byte // The current CRS state
	ContributorID   string   // NodeID of the contributor
	Signature       []byte   // Signature on (EpochNumber, Step, IntermediateCRS)
	Proof           *ChaumPedersenProof // Chaum-Pedersen proof of correct exponentiation
}

// CRSCeremonyState tracks the state of an ongoing CRS ceremony
// Each sequencer maintains this for the current epoch
type CRSCeremonyState struct {
	EpochNumber     int64
	Participants    []string // Ordered list of participant NodeIDs
	CurrentStep     int      // Whose turn (index in Participants)
	IntermediateCRS [][]byte // Current CRS state
	Completed       bool
	Mutex           sync.Mutex
}

// NewCRSCeremonyState creates a new ceremony state for an epoch
func NewCRSCeremonyState(epochNumber int64, participants []string, initialCRS [][]byte) *CRSCeremonyState {
	return &CRSCeremonyState{
		EpochNumber:     epochNumber,
		Participants:    participants,
		CurrentStep:     0,
		IntermediateCRS: initialCRS,
		Completed:       false,
	}
}

// CheckTurn returns true if the given nodeID is the current contributor
func (s *CRSCeremonyState) CheckTurn(nodeID string) bool {
	return s.Participants[s.CurrentStep] == nodeID
}

// AdvanceStep advances the ceremony to the next participant
func (s *CRSCeremonyState) AdvanceStep(newCRS [][]byte) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.CurrentStep++
	s.IntermediateCRS = newCRS
	if s.CurrentStep >= len(s.Participants) {
		s.Completed = true
	}
}

// HandleCRSCeremonyMessage processes an incoming ceremony message
// (Application should call this on message receipt)
func (s *CRSCeremonyState) HandleCRSCeremonyMessage(msg *CRSCeremonyMessage) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if s.Completed {
		return fmt.Errorf("ceremony already completed")
	}
	if msg.EpochNumber != s.EpochNumber {
		return fmt.Errorf("wrong epoch")
	}
	if msg.Step != s.CurrentStep {
		return fmt.Errorf("unexpected step: got %d, expected %d", msg.Step, s.CurrentStep)
	}
	// TODO: Verify signature and contributor ID
	// Accept the new CRS and advance
	s.IntermediateCRS = msg.IntermediateCRS
	s.CurrentStep++
	if s.CurrentStep >= len(s.Participants) {
		s.Completed = true
	}
	return nil
}

// GenerateInitialCRS creates the initial CRS state using BN254 and compressed points
// CRS is [g1, g1^s, g1^{s^2}, ..., g1^{s^n}] for random s
func GenerateInitialCRS(crsSize int) ([][]byte, *big.Int, error) {
	var crs [][]byte
	// Generate random scalar s
	var s fr.Element
	s.SetRandom()
	sBig := s.BigInt(new(big.Int))
	modulus := fr.Modulus()

	// Start with g1 (generator: (1, 2) for BN254)
	var g1Gen bn254.G1Affine
	g1Gen.X.SetString("1")
	g1Gen.Y.SetString("2")
	current := g1Gen
	b := current.Bytes()
	crs = append(crs, b[:]) // g1

	power := new(big.Int).Set(sBig)
	for i := 1; i < crsSize; i++ {
		// g1^{s^i}
		var tmpAffine bn254.G1Affine
		tmpAffine.ScalarMultiplication(&g1Gen, power)
		b := tmpAffine.Bytes()
		crs = append(crs, b[:])
		power.Mul(power, sBig)
		power.Mod(power, modulus)
	}
	return crs, sBig, nil
}

// ChaumPedersenProof represents a Chaum-Pedersen proof
// (A, B, C, S) for g, h, g^r, h^r
// A = g^k, B = h^k, C = challenge, S = response
// All points are compressed []byte, scalars are big.Int

type ChaumPedersenProof struct {
	A []byte
	B []byte
	C *big.Int
	S *big.Int
}

// AddContribution applies a participant's randomness to the CRS
// Each point is raised to the power of r (i.e., g1^{s^i} -> g1^{r*s^i})
// Returns new CRS as compressed points and a Chaum-Pedersen proof
func AddContribution(crs [][]byte) ([][]byte, *big.Int, *ChaumPedersenProof, error) {
	var r fr.Element
	r.SetRandom()
	rBig := r.BigInt(new(big.Int))

	var g1, h1 bn254.G1Affine
	if err := g1.Unmarshal(crs[0]); err != nil {
		return nil, nil, nil, err
	}
	if err := h1.Unmarshal(crs[1]); err != nil {
		return nil, nil, nil, err
	}

	var newCRS [][]byte
	for _, ptBytes := range crs {
		var pt bn254.G1Affine
		if err := pt.Unmarshal(ptBytes); err != nil {
			return nil, nil, nil, err
		}
		var jac bn254.G1Jac
		jac.FromAffine(&pt)
		jac.ScalarMultiplication(&jac, rBig)
		var newAffine bn254.G1Affine
		newAffine.FromJacobian(&jac)
		b := newAffine.Bytes()
		newCRS = append(newCRS, b[:])
	}

	var k fr.Element
	k.SetRandom()
	kBig := k.BigInt(new(big.Int))

	var Ajac bn254.G1Jac
	Ajac.FromAffine(&g1)
	Ajac.ScalarMultiplication(&Ajac, kBig)
	var A bn254.G1Affine
	A.FromJacobian(&Ajac)
	Araw := A.Bytes()
	Abytes := Araw[:]

	var Bjac bn254.G1Jac
	Bjac.FromAffine(&h1)
	Bjac.ScalarMultiplication(&Bjac, kBig)
	var B bn254.G1Affine
	B.FromJacobian(&Bjac)
	Braw := B.Bytes()
	Bbytes := Braw[:]

	cHash := sha256.New()
	g1Raw := g1.Bytes()
	h1Raw := h1.Bytes()
	cHash.Write(g1Raw[:])
	cHash.Write(h1Raw[:])
	cHash.Write(newCRS[0])
	cHash.Write(newCRS[1])
	cHash.Write(Abytes)
	cHash.Write(Bbytes)
	cBytes := cHash.Sum(nil)
	c := new(big.Int).SetBytes(cBytes)
	c.Mod(c, fr.Modulus())

	var cElem fr.Element
	cElem.SetBigInt(c)
	var s fr.Element
	s.Mul(&r, &cElem)
	s.Add(&s, &k)
	sBig := s.BigInt(new(big.Int))

	proof := &ChaumPedersenProof{
		A: Abytes,
		B: Bbytes,
		C: c,
		S: sBig,
	}

	return newCRS, rBig, proof, nil
}

// VerifyContributionProof verifies the Chaum-Pedersen proof for a CRS update
func VerifyContributionProof(prevCRS [][]byte, newCRS [][]byte, proof *ChaumPedersenProof) bool {
	if len(prevCRS) < 2 || len(newCRS) < 2 || proof == nil {
		return false
	}
	var g1, h1, g1r, h1r, A, B bn254.G1Affine
	if err := g1.Unmarshal(prevCRS[0]); err != nil {
		return false
	}
	if err := h1.Unmarshal(prevCRS[1]); err != nil {
		return false
	}
	if err := g1r.Unmarshal(newCRS[0]); err != nil {
		return false
	}
	if err := h1r.Unmarshal(newCRS[1]); err != nil {
		return false
	}
	if err := A.Unmarshal(proof.A); err != nil {
		return false
	}
	if err := B.Unmarshal(proof.B); err != nil {
		return false
	}

	cHash := sha256.New()
	g1Raw := g1.Bytes()
	h1Raw := h1.Bytes()
	g1rRaw := g1r.Bytes()
	h1rRaw := h1r.Bytes()
	cHash.Write(g1Raw[:])
	cHash.Write(h1Raw[:])
	cHash.Write(g1rRaw[:])
	cHash.Write(h1rRaw[:])
	cHash.Write(proof.A)
	cHash.Write(proof.B)
	cBytes := cHash.Sum(nil)
	c := new(big.Int).SetBytes(cBytes)
	c.Mod(c, fr.Modulus())

	if c.Cmp(proof.C) != 0 {
		return false
	}

	var sJac, cJac, checkA bn254.G1Jac
	g1Jac := bn254.G1Jac{}
	g1Jac.FromAffine(&g1)
	sJac.ScalarMultiplication(&g1Jac, proof.S)
	g1rJac := bn254.G1Jac{}
	g1rJac.FromAffine(&g1r)
	negC := new(big.Int).Neg(c)
	negC.Mod(negC, fr.Modulus())
	cJac.ScalarMultiplication(&g1rJac, negC)
	checkA.Set(&sJac)
	checkA.AddAssign(&cJac)
	var checkAAffine bn254.G1Affine
	checkAAffine.FromJacobian(&checkA)
	checkARaw := checkAAffine.Bytes()
	if !bytes.Equal(checkARaw[:], proof.A) {
		return false
	}

	var sJacB, cJacB, checkB bn254.G1Jac
	h1Jac := bn254.G1Jac{}
	h1Jac.FromAffine(&h1)
	sJacB.ScalarMultiplication(&h1Jac, proof.S)
	h1rJac := bn254.G1Jac{}
	h1rJac.FromAffine(&h1r)
	cJacB.ScalarMultiplication(&h1rJac, negC)
	checkB.Set(&sJacB)
	checkB.AddAssign(&cJacB)
	var checkBAffine bn254.G1Affine
	checkBAffine.FromJacobian(&checkB)
	checkBRaw := checkBAffine.Bytes()
	if !bytes.Equal(checkBRaw[:], proof.B) {
		return false
	}

	return true
}
