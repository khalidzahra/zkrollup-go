package state

import (
	"crypto/sha256"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// MerkleTree represents a sparse Merkle tree for state management
type MerkleTree struct {
	depth     int
	zeroHashes [][32]byte
	leaves    map[[32]byte][32]byte
}

// NewMerkleTree creates a new sparse Merkle tree with the given depth
func NewMerkleTree(depth int) *MerkleTree {
	tree := &MerkleTree{
		depth:     depth,
		zeroHashes: make([][32]byte, depth+1),
		leaves:    make(map[[32]byte][32]byte),
	}

	// Compute zero hashes for empty branches
	tree.zeroHashes[0] = [32]byte{}
	for i := 1; i <= depth; i++ {
		hash := sha256.New()
		hash.Write(tree.zeroHashes[i-1][:])
		hash.Write(tree.zeroHashes[i-1][:])
		copy(tree.zeroHashes[i][:], hash.Sum(nil))
	}

	return tree
}

// Update updates the leaf value for a given key
func (t *MerkleTree) Update(key [32]byte, value [32]byte) {
	t.leaves[key] = value
}

// Get retrieves the leaf value for a given key
func (t *MerkleTree) Get(key [32]byte) [32]byte {
	if value, exists := t.leaves[key]; exists {
		return value
	}
	return [32]byte{}
}

// GetRoot computes and returns the current root of the Merkle tree
func (t *MerkleTree) GetRoot() [32]byte {
	var root [32]byte
	if len(t.leaves) == 0 {
		return t.zeroHashes[t.depth]
	}

	// Convert leaves to field elements for circuit compatibility
	nodes := make(map[[32]byte]*fr.Element)
	for key, value := range t.leaves {
		var fe fr.Element
		fe.SetBytes(value[:])
		nodes[key] = &fe
	}

	// Compute root using field elements
	for i := t.depth - 1; i >= 0; i-- {
		newNodes := make(map[[32]byte]*fr.Element)
		for key, value := range nodes {
			siblingKey := computeSiblingKey(key, i)
			var siblingValue *fr.Element
			if sibling, exists := nodes[siblingKey]; exists {
				siblingValue = sibling
			} else {
				siblingValue = new(fr.Element)
				siblingValue.SetBytes(t.zeroHashes[i][:])
			}

			parentKey := computeParentKey(key, i)
			parentValue := hashPair(value, siblingValue)
			newNodes[parentKey] = parentValue
		}
		nodes = newNodes
	}

	// Convert root field element back to bytes
	if len(nodes) > 0 {
		for _, value := range nodes {
			rootBytes := value.Bytes()
			copy(root[:], rootBytes[:])
			break
		}
	}

	return root
}

// GenerateProof generates a Merkle proof for a given key
func (t *MerkleTree) GenerateProof(key [32]byte) ([][32]byte, error) {
	if _, exists := t.leaves[key]; !exists {
		return nil, fmt.Errorf("key not found")
	}

	proof := make([][32]byte, t.depth)
	currentKey := key

	for i := 0; i < t.depth; i++ {
		siblingKey := computeSiblingKey(currentKey, i)
		if sibling, exists := t.leaves[siblingKey]; exists {
			proof[i] = sibling
		} else {
			proof[i] = t.zeroHashes[i]
		}
		currentKey = computeParentKey(currentKey, i)
	}

	return proof, nil
}

// VerifyProof verifies a Merkle proof for a given key and value
func (t *MerkleTree) VerifyProof(key [32]byte, value [32]byte, proof [][32]byte, root [32]byte) bool {
	if len(proof) != t.depth {
		return false
	}

	currentHash := value
	currentKey := key

	for i := 0; i < t.depth; i++ {
		siblingHash := proof[i]
		parentHash := hashPairBytes(currentHash, siblingHash)
		currentHash = parentHash
		currentKey = computeParentKey(currentKey, i)
	}

	return currentHash == root
}

// Helper functions
func computeSiblingKey(key [32]byte, level int) [32]byte {
	var result [32]byte
	copy(result[:], key[:])
	result[level/8] ^= 1 << uint(level%8)
	return result
}

func computeParentKey(key [32]byte, level int) [32]byte {
	var result [32]byte
	copy(result[:], key[:])
	result[level/8] &= ^(1 << uint(level%8))
	return result
}

func hashPair(left, right *fr.Element) *fr.Element {
	result := new(fr.Element)
	leftBytes := left.Bytes()
	rightBytes := right.Bytes()

	hash := sha256.New()
	hash.Write(leftBytes[:])
	hash.Write(rightBytes[:])
	
	result.SetBytes(hash.Sum(nil))
	return result
}

func hashPairBytes(left, right [32]byte) [32]byte {
	var result [32]byte
	hash := sha256.New()
	hash.Write(left[:])
	hash.Write(right[:])
	copy(result[:], hash.Sum(nil))
	return result
}
