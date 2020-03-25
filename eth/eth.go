package eth

import (
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"gitlab.forceup.in/hdwallet"
	"strings"
)

type eth_hdk struct {
	*hdwallet.BaseExtKey
}

func (self *eth_hdk) Child(index uint32) (hdwallet.IExtendKey, error) {
	baseext, err := self.BaseExtKey.Child(index)
	if err != nil {
		return nil, err
	}
	return &eth_hdk{BaseExtKey: baseext}, nil
}

func (self *eth_hdk) Base() *hdwallet.BaseExtKey {
	return self.BaseExtKey
}

func (self *eth_hdk) String() string {
	if self.IsPrivate() {
		private, err := self.Private()
		if err != nil {
			return ""
		}
		return strings.ToLower(hexutil.Encode(crypto.FromECDSA(private)))
	} else {
		public, err := self.Public()
		if err != nil {
			return ""
		}
		return strings.ToLower(hexutil.Encode(crypto.FromECDSAPub(public)))
	}
}

func (self *eth_hdk) ExtendKeyStr() string {
	return self.Extkey.String()
}

func (self *eth_hdk) Address(param interface{}) (hdwallet.Address, error) {
	public, err := self.Public()
	if err != nil {
		return nil, err
	}
	return crypto.PubkeyToAddress(*public), nil
}

func (self *eth_hdk) init() (*eth_hdk, error) {
	return self, nil
}

func NewInstance(extend_key_str string) (hdwallet.IExtendKey, error) {
	var hdk *hdkeychain.ExtendedKey
	var err error
	if hdk, err = hdkeychain.NewKeyFromString(extend_key_str); err != nil {
		return nil, err
	}
	return NewInst(hdk)
}

func NewInst(hdk *hdkeychain.ExtendedKey) (*eth_hdk, error) {
	return (&eth_hdk{BaseExtKey: &hdwallet.BaseExtKey{hdk}}).init()
}
