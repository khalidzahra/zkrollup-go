package crsutils

import (
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

// TransformCRSWithRandomScalar applies BN254 scalar multiplication to each compressed G1 point in the CRS
// using a random scalar, returning the new CRS and the scalar used.
func TransformCRSWithRandomScalar(currentCRS []byte) ([]byte, *big.Int, error) {
	g1PointSize := bn254.SizeOfG1AffineCompressed
	numPoints := len(currentCRS) / g1PointSize
	if len(currentCRS)%g1PointSize != 0 || numPoints == 0 {
		return nil, nil, errors.New("invalid CRS size for BN254 compressed points")
	}
	// Generate random scalar
	scalarBytes := make([]byte, fr.Bytes)
	if _, err := rand.Read(scalarBytes); err != nil {
		return nil, nil, err
	}
	scalar := new(big.Int).SetBytes(scalarBytes)
	var frScalar fr.Element
	frScalar.SetBigInt(scalar)
	bigScalar := frScalar.ToBigIntRegular(new(big.Int))
	// Transform each CRS G1 point
	myCRS := make([]byte, 0, len(currentCRS))
	for i := 0; i < numPoints; i++ {
		var pt bn254.G1Affine
		if err := pt.Unmarshal(currentCRS[i*g1PointSize : (i+1)*g1PointSize]); err != nil {
			return nil, nil, err
		}
		var newPt bn254.G1Affine
		newPt.ScalarMultiplication(&pt, bigScalar)
		b := newPt.Bytes()
		myCRS = append(myCRS, b[:]...)
	}
	return myCRS, scalar, nil
}

// GenerateRandomCRS returns a CRS consisting of n random BN254 G1 points (compressed)
func GenerateRandomCRS(numPoints int) ([]byte, error) {
	if numPoints <= 0 {
		return nil, errors.New("numPoints must be positive")
	}
	g1PointSize := bn254.SizeOfG1AffineCompressed
	crs := make([]byte, 0, numPoints*g1PointSize)
	for i := 0; i < numPoints; i++ {
		scalarBytes := make([]byte, fr.Bytes)
		if _, err := rand.Read(scalarBytes); err != nil {
			return nil, err
		}
		scalar := new(big.Int).SetBytes(scalarBytes)
		var frScalar fr.Element
		frScalar.SetBigInt(scalar)
		bigScalar := frScalar.ToBigIntRegular(new(big.Int))
		var pt bn254.G1Affine
		pt.ScalarMultiplication(&bn254.G1Affine{}, bigScalar)
		b := pt.Bytes()
		crs = append(crs, b[:]...)
	}
	return crs, nil
}
