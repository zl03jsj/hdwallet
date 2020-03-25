package hdwallet

import (
	"gitlab.forceup.in/hdwallet/eth"
	"gitlab.forceup.in/hdwallet/utils"
	"testing"
)

func for_eth() {
	var (
		master_public string
		master_private string
	)
	master, err := utils.NewExtMaster()
	utils.Fatal_error(err)
	eth.NewInst(master)
}

func TestEth(t *testing.T) {

}

func for_filcoin() {

}
