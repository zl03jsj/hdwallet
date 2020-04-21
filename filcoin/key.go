package filcoin

import (
	"fmt"
	"github.com/filecoin-project/go-address"
	crypto2 "github.com/filecoin-project/go-crypto"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"golang.org/x/xerrors"
)

type Key struct {
	types.KeyInfo
	PublicKey []byte
	Address   address.Address
}

const KTBLS = "bls"
const KTSecp256k1 = "secp256k1"

func ActSigType(typ string) crypto.SigType {
	switch typ {
	case KTBLS:
		return crypto.SigTypeBLS
	case KTSecp256k1:
		return crypto.SigTypeSecp256k1
	default:
		return 0
	}
}

func NewKey(keyinfo types.KeyInfo) (*Key, error) {
	k := &Key{
		KeyInfo: keyinfo,
	}

	var err error

	if ActSigType(keyinfo.Type) != crypto.SigTypeSecp256k1 {
		return nil, fmt.Errorf("unkown sig type:%s", keyinfo.Type)
	}
	k.PublicKey = crypto2.PublicKey(k.PrivateKey)

	switch k.Type {
	case KTSecp256k1:
		k.Address, err = address.NewSecp256k1Address(k.PublicKey)
		if err != nil {
			return nil, xerrors.Errorf("converting Secp256k1 to address: %w", err)
		}
	case KTBLS:
		k.Address, err = address.NewBLSAddress(k.PublicKey)
		if err != nil {
			return nil, xerrors.Errorf("converting BLS to address: %w", err)
		}
	default:
		return nil, xerrors.Errorf("unknown key type")
	}
	return k, nil

}
