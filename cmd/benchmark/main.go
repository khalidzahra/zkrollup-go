package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	localCrypto "zkrollup/pkg/crypto"
	"zkrollup/pkg/verifier"

	ed "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark-crypto/signature"
	"github.com/consensys/gnark-crypto/signature/eddsa"
	gnarkEddsa "github.com/consensys/gnark/std/signature/eddsa"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/suite"
)

// ProofWrapper for the internal gnark proof
type ProofWrapper struct {
	Ar struct {
		X, Y [4]uint64
	}
	Krs struct {
		X, Y [4]uint64
	}
	Bs struct {
		X, Y struct {
			A0, A1 [4]uint64
		}
	}
	Commitments     []interface{}
	CommitmentProof struct {
		X, Y [4]uint64
	}
}

type ExportSolidityTestSuite struct {
	suite.Suite

	// backend
	backend *simulated.Backend

	// verifier contract
	verifierContract *verifier.Verifier

	// groth16 gnark objects
	vk      groth16.VerifyingKey
	pk      groth16.ProvingKey
	circuit localCrypto.TransactionCircuit
	r1cs    constraint.ConstraintSystem
	prover  *localCrypto.Prover
}

func (t *ExportSolidityTestSuite) SetupTest() {

	const gasLimit uint64 = 10_000_000

	// setup simulated backend
	key, _ := crypto.GenerateKey()
	auth, _ := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	alloc := types.GenesisAlloc{
		auth.From: {Balance: big.NewInt(1000000000000000000)}, // 1 ETH in wei
	}
	t.backend = simulated.NewBackend(alloc)

	auth.GasLimit = gasLimit
	auth.GasPrice = big.NewInt(1000000000)
	contractBackend := t.backend.Client()

	// deploy verifier contract
	_, _, v, err := verifier.DeployVerifier(auth, contractBackend)
	if err != nil {
		log.Fatalf("failed to deploy verifier contract: %v", err)
	}
	t.verifierContract = v
	t.backend.Commit()

	t.pk = groth16.NewProvingKey(ecc.BN254)
	{
		f, _ := os.Open("./generate/zkrollup.pk")
		_, err = t.pk.ReadFrom(f)
		f.Close()
		if err != nil {
			log.Fatalf("failed to read proving key: %v", err)
		}
	}
	t.vk = groth16.NewVerifyingKey(ecc.BN254)
	{
		f, _ := os.Open("./generate/zkrollup.vk")
		_, err = t.vk.ReadFrom(f)
		f.Close()
		if err != nil {
			log.Fatalf("failed to read verifying key: %v", err)
		}
	}

	prover, err := localCrypto.NewProverWithKeys(t.pk, t.vk)
	if err != nil {
		log.Fatalf("failed to create prover: %v", err)
	}
	t.prover = prover
}

func (t *ExportSolidityTestSuite) TestVerifyProof() {

	// create a valid proof
	hFunc := hash.MIMC_BN254.New()

	senderPubKey, senderSigner, err := GenerateKeyPair()
	if err != nil {
		log.Fatalf("failed to generate sender key pair: %v", err)
	}
	receiverPubKey, _, err := GenerateKeyPair()
	if err != nil {
		log.Fatalf("failed to generate receiver key pair: %v", err)
	}

	gSenderPubKey := gnarkEddsa.PublicKey{}
	gReceiverPubKey := gnarkEddsa.PublicKey{}

	gSenderPubKey.Assign(ed.BN254, senderPubKey.Bytes()[:32])
	gReceiverPubKey.Assign(ed.BN254, receiverPubKey.Bytes()[:32])

	amount := big.NewInt(100)
	balance := big.NewInt(200)
	nonce := big.NewInt(200)

	hFunc.Write(getBytes(fmt.Sprint(gReceiverPubKey.A.X)))
	hFunc.Write(getBytes(fmt.Sprint(gReceiverPubKey.A.Y)))
	hFunc.Write(amount.Bytes())
	hFunc.Write(balance.Bytes())
	hFunc.Write(nonce.Bytes())

	nonceHash := hFunc.Sum(nil)

	signature, err := senderSigner.Sign(nonceHash, hFunc)
	if err != nil {
		log.Fatalf("failed to sign nonce: %v", err)
	}

	gSig := gnarkEddsa.Signature{}

	gSig.Assign(ed.BN254, signature)

	witness, err := t.prover.CreateWitness(
		gSenderPubKey,
		gReceiverPubKey,
		amount,
		nonce,
		gSig,
		balance,
	)
	if err != nil {
		log.Fatalf("failed to create witness: %v", err)
	}

	proof, pubWitness, err := t.prover.GenerateProof(witness)
	if err != nil {
		log.Fatalf("failed to generate proof: %v", err)
	}

	// ensure gnark (Go) code verifies it
	err = groth16.Verify(proof, t.vk, pubWitness)
	if err != nil {
		log.Fatalf("failed to verify proof: %v", err)
	}

	proofBytes, err := t.prover.SerializeProof(proof)
	if err != nil {
		log.Fatalf("failed to serialize proof: %v", err)
	}

	pubWitnessBytes, err := t.prover.SerializePublicWitness(pubWitness)
	if err != nil {
		log.Fatalf("failed to serialize public witness: %v", err)
	}

	witnessArray := [6]*big.Int{}
	for i := range 6 {
		witnessArray[i] = new(big.Int).SetBytes(pubWitnessBytes[i*32 : (i+1)*32])
	}

	proofArr := [8]*big.Int{}
	for i := range 8 {
		proofArr[i] = new(big.Int).SetBytes(proofBytes[i*32 : (i+1)*32])
	}

	// Call the contract with the converted arrays
	err = t.verifierContract.VerifyProof(&bind.CallOpts{}, proofArr, witnessArray)
	if err != nil {
		log.Fatalf("failed to verify proof on contract: %v", err)
	}
}

func main() {
	suite := new(ExportSolidityTestSuite)
	suite.SetupTest()
	suite.TestVerifyProof()
	fmt.Println("Test passed")
}

func getBytes(str string) []byte {
	trimmed := strings.Trim(str, "[]")
	parts := strings.Fields(trimmed) // Splits on whitespace
	bytes := make([]byte, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(p)
		if err != nil {
			panic(err)
		}
		bytes[i] = byte(val)
	}
	return bytes
}

func GenerateKeyPair() (signature.PublicKey, signature.Signer, error) {
	privateKey, err := eddsa.New(ed.BN254, rand.Reader)
	if err != nil {
		fmt.Println("failed to create a key pair. error:", err)
		return nil, nil, err
	}
	publicKey := privateKey.Public()

	return publicKey, privateKey, nil
}

// Convert proof bytes to [8]*big.Int
func bytesToProofArray(data []byte) ([8]*big.Int, error) {
	if len(data) < 8*32 {
		return [8]*big.Int{}, fmt.Errorf("insufficient proof data")
	}

	var proof [8]*big.Int
	for i := range 8 {
		start := i * 32
		end := start + 32
		proof[i] = new(big.Int).SetBytes(data[start:end])
	}

	return proof, nil
}

// Convert witness bytes to [6]*big.Int
func bytesToWitnessArray(data []byte) ([6]*big.Int, error) {
	if len(data) < 6*32 {
		return [6]*big.Int{}, fmt.Errorf("insufficient witness data")
	}

	var witness [6]*big.Int
	for i := range 6 {
		start := i * 32
		end := start + 32
		witness[i] = new(big.Int).SetBytes(data[start:end])
	}

	return witness, nil
}

func CastProof(proof groth16.Proof) (*ProofWrapper, error) {
	// Get the underlying concrete value using reflection
	proofValue := reflect.ValueOf(proof)
	if proofValue.Kind() == reflect.Ptr {
		proofValue = proofValue.Elem()
	}

	// Verify the underlying type matches our wrapper
	if proofValue.NumField() != 5 ||
		proofValue.Type().Field(0).Name != "Ar" ||
		proofValue.Type().Field(1).Name != "Krs" ||
		proofValue.Type().Field(2).Name != "Bs" {
		return nil, fmt.Errorf("proof has unexpected structure")
	}

	// Create a new wrapper and copy the data via unsafe pointer
	var wrapper ProofWrapper
	*(*[unsafe.Sizeof(wrapper)]byte)(unsafe.Pointer(&wrapper)) =
		*(*[unsafe.Sizeof(wrapper)]byte)(unsafe.Pointer(proofValue.UnsafeAddr()))

	return &wrapper, nil
}

func limbsToBigInt(limbs [4]uint64) *big.Int {
	// Convert 4 uint64 limbs to 32-byte big-endian representation
	var bytes [32]byte
	binary.BigEndian.PutUint64(bytes[0:8], limbs[3]) // Most significant limb
	binary.BigEndian.PutUint64(bytes[8:16], limbs[2])
	binary.BigEndian.PutUint64(bytes[16:24], limbs[1])
	binary.BigEndian.PutUint64(bytes[24:32], limbs[0]) // Least significant limb
	return new(big.Int).SetBytes(bytes[:])
}
