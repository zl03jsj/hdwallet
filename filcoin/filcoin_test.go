package filcoin

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"gitlab.forceup.in/hdwallet"
	"gitlab.forceup.in/hdwallet/utils"
	"testing"
)

func Test_ecdsaFilCoinAddress(t *testing.T) {
	seed, _ := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	master, _ := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)

	for i := uint32(0); i < 100; i++ {
		extkey, _ := master.Child(i)

		btcpriv, _ := extkey.ECPrivKey()
		ecdspriv := btcpriv.ToECDSA()

		ecdspubl := ecdspriv.PublicKey

		filcoin_priv, err := Filcoin_key_from_private(ecdspriv)
		utils.Fatal_error(err)
		filcoin_publ, err := filcoin_key_from_public(&ecdspubl)
		utils.Fatal_error(err)

		t.Logf("address from privatekey : %s\n", filcoin_priv.Address.String())
		t.Logf("address from publickey  : %s\n", filcoin_publ.Address.String())

		if filcoin_priv.Address.String() == filcoin_publ.Address.String() {
			t.Logf("address testing ok!!!!!!\n")
		} else {
			t.Errorf("address testing failed!!!!!!\n")
		}
	}
}

func Test_filcoin_extkey(t *testing.T) {
	master, err := utils.NewExtMaster()
	utils.Fatal_error(err)

	private_ext, err := new_from_extkey(master)
	utils.Fatal_error(err)

	neuter, err := master.Neuter()
	utils.Fatal_error(err)

	public__ext, err := new_from_extkey(neuter)
	utils.Fatal_error(err)

	for i := uint32(0); i < 100; i++ {

		child_x, err := private_ext.Child(i)
		utils.Fatal_error(err)
		child_y, err := public__ext.Child(i)
		utils.Fatal_error(err)

		addr_x, err := child_x.Address(nil)
		utils.Fatal_error(err)
		addr_y, err := child_y.Address(nil)
		utils.Fatal_error(err)

		t.Logf("address from privatekey : %s\n", addr_x.String())
		t.Logf("address from publickey  : %s\n", addr_y.String())

		if addr_x.String() == addr_y.String() {
			t.Logf("address testing ok!!!!!!\n")
		} else {
			t.Errorf("address testing failed!!!!!!\n")
		}
	}
}

func TestFilcoinkey_from_string(t *testing.T) {
	// lotus wallet export
	addr_str := "t1gsu7dkdufuygtbcclhmjcao7vg7cap46l6ij2ri"
	hex_str := "7b2254797065223a22736563703235366b31222c22507269766174654b6579223a22674f447046677043326b424b5a565a4d70682b6d4d6b36514a3473793732437a7041656452386a654f61633d227d"
	walletk, err := Filcoinkey_from_string(hex_str)
	utils.Fatal_error(err)

	fmt.Printf("address x = %s\n", walletk.Address.String())
	fmt.Printf("address y = %s\n", addr_str)

	if walletk.Address.String() != addr_str {
		t.Errorf("failed, address not matched!!!\n")
	} else {
		t.Logf("success!!!!!!!")
	}

	priv, err := ecdsakey_from_private_key_data(walletk.PrivateKey)
	utils.Fatal_error(err)

	walletk, err = Filcoin_key_from_private(priv)
	utils.Fatal_error(err)

	fmt.Printf("address x = %s\n", walletk.Address.String())
	fmt.Printf("address y = %s\n", addr_str)

	if walletk.Address.String() != addr_str {
		t.Errorf("failed, address not matched!!!\n")
	} else {
		t.Logf("success!!!!!!!")
	}
}

func Test_filcoin_hdk(t *testing.T) {
	hdk, err := NewHdkey()
	utils.Fatal_error(err)
	hdk_testing(hdk, t)
}

func hdk_testing(hdk *hdwallet.HdKey, t *testing.T) {
	var start = uint32(1234567)
	var end = start + 66
	for index := start; index < end; index++ {
		child, err := hdk.Child(index)
		utils.Fatal_error(err)
		addr, err := child.ExtKey.Address(nil)
		utils.Fatal_error(err)
		fmt.Printf("index:%d, key:%s, childkey:%s, address:%s\n", index, child.ExtKey.String(), child.Chiper, addr.String())

		iextkey, err := hdk.ExtKeyFromKey(child.Chiper)
		utils.Fatal_error(err)

		if iextkey.String() != child.ExtKey.String() {
			t.Errorf("address not match, failed")
		} else {
			t.Logf("ok!!!!!!!")
		}
	}
}
