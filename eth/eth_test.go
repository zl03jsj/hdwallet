package eth

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"gitlab.forceup.in/hdwallet"
	"gitlab.forceup.in/hdwallet/utils"
	"strings"
	"testing"
)

func Test_eth_private_key_and_address(t *testing.T) {
	hex_addr := strings.ToLower("0xE754B487Ce0c268b4413642Cf6Bc684759F95211")
	hex_ := strings.ToLower("0x34854C80F72EE70ADACAE600F8AE5B4EE9AF7448F5034F03B9391AC08E95C44A")
	hex_data := common.FromHex(hex_)

	privatekey, err := crypto.ToECDSA(hex_data)

	utils.Fatal_error(err)

	private_hex := strings.ToLower(hexutil.Encode(crypto.FromECDSA(privatekey)))

	fmt.Printf("private x:=%s\n", hex_)
	fmt.Printf("private y:=%s\n", private_hex)
	if private_hex != hex_ {
		t.Errorf("failed!!!!!!\n")
	} else {
		t.Logf("success")
	}

	address := crypto.PubkeyToAddress(privatekey.PublicKey)
	address_x := strings.ToLower(address.String())
	fmt.Printf("address x:%s\n", address_x)
	if hex_addr != address_x {
		t.Errorf("failed!!!!\n")
	} else {
		t.Logf("success!!!!\n")
	}
}

func Test_eth_extendkey(t *testing.T) {
	// seed, _ := hdkeychain.GenerateSeed(hdkeychain.RecommendedSeedLen)
	// master, _ := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	master, _ := utils.NewExtMaster()

	private_ext, _ := new_from_extkey(master)
	neuter, _ := master.Neuter()
	public__ext, _ := new_from_extkey(neuter)

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

func Test_eth_hdk(t *testing.T) {
	hdk, err := NewHdkey()
	utils.Fatal_error(err)
	hdk_testing(hdk, t)
}

func hdk_testing(hdk *hdwallet.HdKey, t *testing.T) {
	start := uint32(1000000)
	end := start + 30
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
