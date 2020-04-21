package main

import (
	"encoding/hex"
	"fmt"
	"github.com/filecoin-project/specs-actors/actors/crypto"
	"gitlab.forceup.in/hdwallet/filcoin"
	utils2 "gitlab.forceup.in/hdwallet/utils"
)

func main() {
	filcoin_sign()
}


func filcoin_sign() {
	private_data, _ := hex.DecodeString("6053f1758bb263ba3d98d3778b6e57b659623d1cdd006090ca9abb5cbc985352")
	for i := 0; i < 100; i++ {
		to_sign_str := utils2.RandString(1024)
		fmt.Printf("index=%d-------------------------\n", i)
		if signed, err := filcoin.Sign(
			crypto.SigTypeSecp256k1,
			private_data, []byte(to_sign_str)); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("btc_sign:", hex.EncodeToString(signed))
		}
	}
}