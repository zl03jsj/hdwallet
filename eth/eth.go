package eth

import (
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"gitlab.forceup.in/hdwallet"
	"gitlab.forceup.in/hdwallet/utils"
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

func new_from_extkey_str(extend_key_str string) (hdwallet.IExtendKey, error) {
	var hdk *hdkeychain.ExtendedKey
	var err error
	if hdk, err = hdkeychain.NewKeyFromString(extend_key_str); err != nil {
		return nil, err
	}
	return new_from_extkey(hdk)
}

func new_from_extkey(hdk *hdkeychain.ExtendedKey) (*eth_hdk, error) {
	return (&eth_hdk{BaseExtKey: &hdwallet.BaseExtKey{hdk}}).init()
}

func NewHdkey() (*hdwallet.HdKey, error) {
	master, err := utils.NewExtMaster()
	utils.Fatal_error(err)
	eth_hdk, err := new_from_extkey(master)
	utils.Fatal_error(err)
	return hdwallet.NewFromExtKey(eth_hdk, "ethereum", 0)
}

func NewHdkFromExtkeyString(extKeyStr, slat string, first uint32) (*hdwallet.HdKey, error) {
	extkey, err := new_from_extkey_str(extKeyStr)
	if err != nil {
		return nil, err
	}

	return hdwallet.NewFromExtKey(extkey, slat, first)
}
