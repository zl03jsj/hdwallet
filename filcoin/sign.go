package filcoin

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"golang.org/x/crypto/blake2b"
)

func Sign(sigType crypto.SigType, priv, msg []byte) ([]byte, error) {
	if sigType != crypto.SigTypeSecp256k1 {
		return nil, fmt.Errorf("not supported sign type")
	}
	var private *ecdsa.PrivateKey
	var err error

	if private, err = Ecdsakey_from_private_key_data(priv); err != nil {
		return nil, err
	}
	hash := blake2b.Sum256(msg)
	return sign(hash[:], private)
}

// github.com/ethereum/go-ethereum/crypto/signature_nocgo.go
func sign(hash []byte, prv *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	if prv.Curve != btcec.S256() {
		return nil, fmt.Errorf("private key curve is not secp256k1")
	}
	sig, err := btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(prv), hash, false)
	if err != nil {
		return nil, err
	}
	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[64] = v
	return sig, nil
}
