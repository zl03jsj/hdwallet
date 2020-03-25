package filcoin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/chain/wallet"
	_ "github.com/filecoin-project/lotus/lib/sigs/secp"
	"github.com/ipsn/go-secp256k1"
	"gitlab.forceup.in/hdwallet"
	"math/big"
)

type filcoin_hdk struct {
	*hdwallet.BaseExtKey
	filcoinkey *wallet.Key
}

func (self *filcoin_hdk) Child(index uint32) (hdwallet.IExtendKey, error) {
	baseext, err := self.BaseExtKey.Child(index)
	if err != nil {
		return nil, err
	}
	return NewInst(baseext.Extkey)
}

func (self *filcoin_hdk) Base() *hdwallet.BaseExtKey {
	return self.BaseExtKey
}

func (self *filcoin_hdk) String() string {
	if buffer, err := json.Marshal(self.filcoinkey.KeyInfo); err != nil {
		return ""
	} else {
		return hex.EncodeToString(buffer)
	}
}

func (self *filcoin_hdk) ExtendKeyStr() string {
	return self.Extkey.String()
}

func (self *filcoin_hdk) Address(param interface{}) (hdwallet.Address, error) {
	// todo : address.Testnet or address.Mainnet ??????
	return self.filcoinkey.Address, nil
}

func (self *filcoin_hdk) init() (*filcoin_hdk, error) {
	var err error

	if self.IsPrivate() {
		var private *ecdsa.PrivateKey
		if private, err = self.Private(); err == nil {
			self.filcoinkey, err = filcoin_key_from_private(private)
		}
	} else {
		var public *ecdsa.PublicKey
		if public, err = self.Public(); err == nil {
			self.filcoinkey, err = filcoin_key_from_public(public)
		}
	}

	return self, nil
}

func filcoin_raw_private(key *ecdsa.PrivateKey) []byte {
	privkey := make([]byte, 32)
	blob := key.D.Bytes()
	copy(privkey[32-len(blob):], blob)
	return privkey
}

func filcoin_key_from_public(public *ecdsa.PublicKey) (*wallet.Key, error) {
	public_data := elliptic.Marshal(public.Curve, public.X, public.Y)
	address, err := address.NewSecp256k1Address(public_data)
	if err != nil {
		return nil, err
	}

	filcoinkey := &wallet.Key{
		KeyInfo:   types.KeyInfo{types.KTSecp256k1, nil},
		PublicKey: public_data,
		Address:   address}
	return filcoinkey, nil
}

func filcoin_key_from_private(key *ecdsa.PrivateKey) (*wallet.Key, error) {
	private_data := filcoin_raw_private(key)
	return wallet.NewKey(types.KeyInfo{types.KTSecp256k1, private_data})
}

func NewInst(ext_key *hdkeychain.ExtendedKey) (hdwallet.IExtendKey, error) {
	return (&filcoin_hdk{BaseExtKey: &hdwallet.BaseExtKey{ext_key}}).init()
}

func NewInsFromExtString(extend_key_str string) (hdwallet.IExtendKey, error) {
	var hdk *hdkeychain.ExtendedKey
	var err error
	if hdk, err = hdkeychain.NewKeyFromString(extend_key_str); err != nil {
		return nil, err
	}
	return NewInst(hdk)
}

// cli command: lotus wallet export 't1gsu7dkdufuygtbcclhmjcao7vg7cap46l6ij2ri',
// get following result:
// "7b2254797065223a22736563703235366b31222c22507269766174654b6579223a22674f447046677043326b424b5a565a4d70682b6d4d6b36514a3473793732437a7041656452386a654f61633d227d"
// use following function to get private key, address from result string
func filcoinkey_from_string(hex_ string) (*wallet.Key, error) {
	hexdata, err := hex.DecodeString(hex_)
	if err != nil {
		return nil, err
	}

	var keyinfo types.KeyInfo
	err = json.Unmarshal(hexdata, &keyinfo)
	if err != nil {
		return nil, err
	}

	return wallet.NewKey(keyinfo)
}

func ecdsakey_from_private_key_data(prikey []byte) (*ecdsa.PrivateKey, error) {
	c := secp256k1.S256()
	k := big.NewInt(0).SetBytes(prikey)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = c
	priv.D = k
	priv.PublicKey.X, priv.PublicKey.Y = c.ScalarBaseMult(k.Bytes())
	return priv, nil
}
