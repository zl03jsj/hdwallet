package hdwallet

import (
	"crypto/ecdsa"
	"github.com/btcsuite/btcutil/hdkeychain"
)

type BaseExtKey struct {
	Extkey *hdkeychain.ExtendedKey
}

func (self *BaseExtKey) IsPrivate() bool {
	return self.Extkey.IsPrivate()
}

func (self *BaseExtKey) Private() (*ecdsa.PrivateKey, error) {
	if self.Extkey == nil {
		return nil, InvalidHdKey
	}
	private, err := self.Extkey.ECPrivKey()
	if err != nil {
		return nil, err
	}
	return private.ToECDSA(), nil
}

func (self *BaseExtKey) ExtKeyStr() string {
	return self.Extkey.String()
}

func (self *BaseExtKey) Public() (*ecdsa.PublicKey, error) {
	public, err := self.Extkey.ECPubKey()
	if err != nil {
		return nil, err
	}
	return public.ToECDSA(), nil
}

func (self *BaseExtKey) Child(index uint32) (*BaseExtKey, error) {
	baseext, err := self.Extkey.Child(index)
	if err != nil {
		return nil, err
	}
	return &BaseExtKey{Extkey: baseext}, nil
}

type Address interface {
	String() string
}
