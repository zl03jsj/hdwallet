package hdwallet

import (
	"crypto/ecdsa"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcutil/hdkeychain"
	"gitlab.forceup.in/hdwallet/utils"
	"strconv"
	"sync"
)

type HdKey struct {
	slat             string
	next_child_index uint32
	i_ext            IExtendKey
	mutx             *sync.Mutex
}

type HdChildKey struct {
	Chiper string
	ExtKey IExtendKey
}

var (
	InvalidHdKey = fmt.Errorf("Invalid HdKey")
)

func (self *HdKey) IsPrivate() bool {
	return self.IsValid() && self.i_ext.IsPrivate()
}

func (self *HdKey) IsValid() bool {
	return self.i_ext != nil
}

func (self *HdKey) extPubKey() (extkey *hdkeychain.ExtendedKey, err error) {
	if !self.IsValid() {
		err = InvalidHdKey
		return
	}
	return extkey.Neuter()
}

func (self *HdKey) MainKey() IExtendKey {
	return self.i_ext
}

func (self *HdKey) ExtKeyFromKey(key string) (extkey IExtendKey, err error) {
	if !self.IsValid() {
		err = InvalidHdKey
		return
	}

	var index uint32
	index, err = keyToIndex(key, self.slat)
	if err != nil {
		return
	}

	return self.i_ext.Child(index)
}

func (self *HdKey) PrivateEcdsaFromKey(key string) (ecPriKey *ecdsa.PrivateKey, err error) {
	if !self.IsPrivate() {
		err = fmt.Errorf("not a private hdkey")
		return
	}

	var (
		extkey IExtendKey
	)

	extkey, err = self.ExtKeyFromKey(key)
	if err != nil {
		return
	}

	ecPriKey, err = extkey.Base().Private()
	return
}

func (self *HdKey) PublicEcdsaFromKey(key string) (ecPubKey *ecdsa.PublicKey, err error) {
	if !self.IsValid() {
		err = InvalidHdKey
		return
	}

	var extkey IExtendKey
	extkey, err = self.ExtKeyFromKey(key)
	if err != nil {
		return
	}

	return extkey.Base().Public()
}

func (self *HdKey) NextIndex() uint32 {
	var index uint32
	self.mutx.Lock()
	index = self.next_child_index
	self.mutx.Unlock()
	return index
}

func (self *HdKey) NextChilds(nu uint32) (childs []*HdChildKey, err error) {
	if !self.IsValid() {
		err = InvalidHdKey
		return
	}

	if nu == 0 {
		return
	}

	var (
		index uint32
		child *HdChildKey
	)

	self.mutx.Lock()
	defer self.mutx.Unlock()
	for i := uint32(0); i < nu; i++ {
		index = i + self.next_child_index

		child = new(HdChildKey)

		child.ExtKey, err = self.i_ext.Child(index)
		if err != nil {
			return
		}
		child.Chiper, err = indexToKey(index, 64, self.slat)
		if err != nil {
			return
		}

		childs = append(childs, child)
	}

	self.next_child_index = index + 1

	return
}

func (self *HdKey) NextChild() (key string, ecPri *ecdsa.PrivateKey, ecPub *ecdsa.PublicKey, err error) {
	if !self.IsValid() {
		err = InvalidHdKey
		return
	}

	var index uint32
	var extkey IExtendKey

	self.mutx.Lock()
	index = self.next_child_index
	self.next_child_index++
	self.mutx.Unlock()

	key, err = indexToKey(index, 64, self.slat)
	if err != nil {
		return
	}

	extkey, err = self.i_ext.Child(index)
	if err != nil {
		return
	}

	ecPri, err = extkey.Base().Private()
	if err != nil {
		return
	}
	ecPub, err = extkey.Base().Public()
	if err != nil {
		return
	}
	return
}

func NewFromExtKey(extkey IExtendKey, slat string, index uint32) (hdkey *HdKey, err error) {
	if extkey == nil {
		err = fmt.Errorf("Extkey is nil")
		return
	}
	hdkey = &HdKey{slat: slat, i_ext: extkey, next_child_index: index, mutx: new(sync.Mutex)}
	return
}

func indexToKey(index uint32, tolen uint, slat string) (string, error) {
	if tolen < 64 {
		tolen = 64
	}

	index_hex := strconv.FormatUint(uint64(index), 16)

	md5 := utils.MD5(slat + strconv.FormatInt(int64(index), 16))

	index_len := len(index_hex)
	bs := make([]byte, 4)
	// 用32位, 4个字节, 转换成16进制, 形成字符串, 表示index的字符串的位数
	// 字符串需要站8个字符的位置!!
	binary.LittleEndian.PutUint32(bs, uint32(index_len))

	rdlen := tolen - (uint(len(index_hex+md5)) + 8)
	rdstr := utils.RandString(int(rdlen))

	return md5 + fmt.Sprintf("%08x", bs) + rdstr + index_hex, nil
}

func keyToIndex(index_str, slat string) (uint32, error) {
	if len([]byte(index_str)) != 64 {
		return 0, fmt.Errorf("private key length should be 64.")
	}

	index_len_hex := index_str[32:40]

	if index_len_bs, err := hex.DecodeString(index_len_hex); err != nil {
		return 0, err
	} else {
		index_len := binary.LittleEndian.Uint32(index_len_bs)
		real_index_str := string(index_str[64-index_len:])
		if index, err := strconv.ParseUint(real_index_str, 16, 32); err != nil {
			return 0, err
		} else {
			// verify md5, check if data has been changed
			md5 := utils.MD5(slat + strconv.FormatInt(int64(index), 16))
			if md5 == string(index_str[:32]) {
				return uint32(index), nil
			} else {
				return 0, fmt.Errorf("private key hash check faild.")
			}
		}
	}
}
