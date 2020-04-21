package filcoin

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/filecoin-project/lotus/lib/sigs"
	_ "github.com/filecoin-project/lotus/lib/sigs/secp"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"gitlab.forceup.in/hdwallet/utils"
	"golang.org/x/crypto/blake2b"

	"testing"
)

func sign_with_filcoin(private *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	private_data := Filcoin_raw_private(private)
	sign, err := sigs.Sign(crypto.SigTypeSecp256k1, private_data, data)
	if err != nil {
		return nil, err
	}
	return sign.Data, nil
}

func sign_with_btcpkg(private *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	btcec_private := btcec.PrivateKey(*private)
	hash := blake2b.Sum256(data)
	sign, err := btcec_private.Sign(hash[:])
	if err != nil {
		return nil, err
	}
	r := sign.R.Bytes()
	s := sign.S.Bytes()

	r = append(r, s[:]...)
	return r, nil
}

func TestSign(t *testing.T) {
	private_data, _ := hex.DecodeString("6053f1758bb263ba3d98d3778b6e57b659623d1cdd006090ca9abb5cbc985352")
	private, _ := Ecdsakey_from_private_key_data(private_data)
	var btc_sign, ffi_sign, my_sign []byte
	var err error
	for i := 0; i < 100; i++ {
		to_sign_str := utils.RandString(1024)
		fmt.Printf("index=%d-------------------------\n", i)

		if ffi_sign, err = sign_with_filcoin(private, []byte(to_sign_str)); err != nil {
			fmt.Println(err.Error())
		}

		if btc_sign, err = sign_with_btcpkg(private, []byte(to_sign_str)); err != nil {
			fmt.Println(err)
		}

		if my_sign, err = Sign(crypto.SigTypeSecp256k1, private_data, []byte(to_sign_str)); err != nil {
			fmt.Println(err)
		}

		fmt.Printf(`
		to_sign_data=%s...
		ffi_sign  :  %s
		btc_sign  :  %s
		my__sign  :  %s
		`, to_sign_str[:20], hex.EncodeToString(ffi_sign), hex.EncodeToString(btc_sign), hex.EncodeToString(my_sign))
		if hex.EncodeToString(my_sign) != hex.EncodeToString(ffi_sign) {
			t.Errorf("my_sign != ffi_sign!!!!!!!!!!!!!!!!")
		}
		fmt.Println()
	}
}
