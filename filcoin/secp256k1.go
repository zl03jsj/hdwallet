package filcoin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"gitlab.forceup.in/hdwallet"
	"gitlab.forceup.in/hdwallet/utils"
	"math/big"
)

type speck256k1_hdk struct {
	*hdwallet.BaseExtKey
	filcoinkey *Key
}

func (self *speck256k1_hdk) Child(index uint32) (hdwallet.IExtendKey, error) {
	baseext, err := self.BaseExtKey.Child(index)
	if err != nil {
		return nil, err
	}
	return new_from_extkey(baseext.Extkey)
}

func (self *speck256k1_hdk) Base() *hdwallet.BaseExtKey {
	return self.BaseExtKey
}

func (self *speck256k1_hdk) String() string {
	if buffer, err := json.Marshal(self.filcoinkey.KeyInfo); err != nil {
		return ""
	} else {
		return hex.EncodeToString(buffer)
	}
}

func (self *speck256k1_hdk) ExtendKeyStr() string {
	return self.Extkey.String()
}

func (self *speck256k1_hdk) Address(param interface{}) (hdwallet.Address, error) {
	// todo : address.Testnet or address.Mainnet ??????
	return self.filcoinkey.Address, nil
}

func (self *speck256k1_hdk) init() (*speck256k1_hdk, error) {
	var err error

	if self.IsPrivate() {
		var private *ecdsa.PrivateKey
		if private, err = self.Private(); err == nil {
			self.filcoinkey, err = Filcoin_key_from_private(private)
		}
	} else {
		var public *ecdsa.PublicKey
		if public, err = self.Public(); err == nil {
			self.filcoinkey, err = filcoin_key_from_public(public)
		}
	}

	return self, nil
}

func PrivkeyToBytes(key *ecdsa.PrivateKey) []byte {
	privkey := make([]byte, 32)
	blob := key.D.Bytes()
	copy(privkey[32-len(blob):], blob)
	return privkey
}

func filcoin_key_from_public(public *ecdsa.PublicKey) (*Key, error) {
	public_data := elliptic.Marshal(public.Curve, public.X, public.Y)
	address, err := address.NewSecp256k1Address(public_data)
	if err != nil {
		return nil, err
	}

	filcoinkey := &Key{
		KeyInfo:   types.KeyInfo{KTSecp256k1, nil},
		PublicKey: public_data,
		Address:   address}
	return filcoinkey, nil
}

func Filcoin_key_from_private_hex(str string) (*Key, error) {
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	private, err := PrivatkeyFromBytes(data)
	if err != nil {
		return nil, err
	}

	return Filcoin_key_from_private(private)
}

func Filcoin_key_from_private(key *ecdsa.PrivateKey) (*Key, error) {
	private_data := PrivkeyToBytes(key)
	return NewKey(types.KeyInfo{KTSecp256k1, private_data})
}

func new_from_extkey(ext_key *hdkeychain.ExtendedKey) (hdwallet.IExtendKey, error) {
	return (&speck256k1_hdk{BaseExtKey: &hdwallet.BaseExtKey{ext_key}}).init()
}

func new_from_extkey_str(extend_key_str string) (hdwallet.IExtendKey, error) {
	var hdk *hdkeychain.ExtendedKey
	var err error
	if hdk, err = hdkeychain.NewKeyFromString(extend_key_str); err != nil {
		return nil, err
	}
	return new_from_extkey(hdk)
}

// cli command: lotus wallet export 't1gsu7dkdufuygtbcclhmjcao7vg7cap46l6ij2ri',
// get following result:
// "7b2254797065223a22736563703235366b31222c22507269766174654b6579223a22674f447046677043326b424b5a565a4d70682b6d4d6b36514a3473793732437a7041656452386a654f61633d227d"
// use following function to get private key, address from result string
func Filcoinkey_from_string(hex_ string) (*Key, error) {
	hexdata, err := hex.DecodeString(hex_)
	if err != nil {
		return nil, err
	}

	var keyinfo types.KeyInfo
	err = json.Unmarshal(hexdata, &keyinfo)
	if err != nil {
		return nil, err
	}

	return NewKey(keyinfo)
}

func PrivatkeyFromBytes(prikey []byte) (*ecdsa.PrivateKey, error) {
	var priv *ecdsa.PrivateKey

	curve := btcec.S256()
	if true {
		privkkk, _ := btcec.PrivKeyFromBytes(curve, prikey)
		return privkkk.ToECDSA(), nil
	} else {
		k := big.NewInt(0).SetBytes(prikey)
		priv = new(ecdsa.PrivateKey)
		priv.PublicKey.Curve = curve
		priv.D = k
		priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())
	}
	return priv, nil
}

func NewSecp256k1Hdkey() (*hdwallet.HdKey, error) {
	master, err := utils.NewExtMaster()
	utils.Fatal_error(err)
	filcoin_hdk, err := new_from_extkey(master)
	utils.Fatal_error(err)
	return hdwallet.NewFromExtKey(filcoin_hdk, "filcoin", 0)
}

func NewHdkFromExtkeyString(extKeyStr, slat string, first uint32) (*hdwallet.HdKey, error) {
	extkey, err := new_from_extkey_str(extKeyStr)
	if err != nil {
		return nil, err
	}
	return hdwallet.NewFromExtKey(extkey, slat, first)
}
