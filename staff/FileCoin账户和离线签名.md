# FileCoin账户和离线签名

## 椭圆曲线的选择

FileCoin和大多数主流区块链一样，底层使用secp256k1这条椭圆曲线。secp256k1有多种语言版本。

比特币纯Go版本实现：[btcec](https://github.com/btcsuite/btcd/tree/master/btcec)

以太坊采用的C++版本：[go-ethereum/crypto/secp256k1](https://github.com/ethereum/go-ethereum/tree/master/crypto/secp256k1)

以及FileCoin使用的C版本：[go-secp256k1](github.com/ipsn/go-secp256k1)

💡 注意，FileCoin 支持两种签名，一种是椭圆曲线签名，一种是bls签名。

bls签名库使用了rust版本的实现：[Filecoin Proofs FFI](https://github.com/filecoin-project/filecoin-ffi)

> BLS 签名算法是一种可以实现签名聚合和密钥聚合的算法（即可以将多个密钥聚合成一把密钥，将多个签名聚合成一个签名）。在以太坊未来的 Casper 实现中，有非常多的验证者都要对区块签名，要保证系统的安全性，同时节约存储空间，就需要用到这类签名聚合的算法。
>
> [Boneh-Lynn-Shacham](https://www.iacr.org/archive/asiacrypt2001/22480516.pdf)

我们使用椭圆曲线签名来生成私钥。

## 生成私钥

以比特币`btcec`为例，生成符合FileCoin的私钥：

```go
sk, _ := btcec.NewPrivateKey(btcec.S256())
```

## 生成地址

### 获得公钥

```go
import "crypto/elliptic"

// 序列化椭圆曲线上的坐标点为未压缩格式的公钥
pub := elliptic.Marshal(btcec.S256(), sk.PublicKey.X, sk.PublicKey.Y)
```

### 获得地址

#### 对公钥做哈希运算

FileCoin中有一个通用的哈希函数，使用blake2b快速哈希。如下：

```go
func hash(ingest []byte, cfg *blake2b.Config) []byte {
	hasher, err := blake2b.New(cfg)
	if err != nil {
		// If this happens sth is very wrong.
		panic(fmt.Sprintf("invalid address hash configuration: %v", err)) // ok
	}
	if _, err := hasher.Write(ingest); err != nil {
		// blake2bs Write implementation never returns an error in its current
		// setup. So if this happens sth went very wrong.
		panic(fmt.Sprintf("blake2b is unable to process hashes: %v", err)) // ok
	}
	return hasher.Sum(nil)
}
```

对公钥进行一次哈希运算：

```go
hash(pub, &blake2b.Config{Size: 20})
```
原文注释里是这样说的：

> PayloadHashLength defines the hash length taken over addresses using the Actor and SECP256K1 protocols.

这里的`PayloadHashLength`等于20。

#### 生成地址格式

然后拼接地址格式：

```go
explen := 1 + len(payload)
buf := make([]byte, explen)

buf[0] = protocol
copy(buf[1:], payload)
```

上面的`protocol`值固定为1，`payload`为`hash(pub)`。

#### 对地址编码

FileCoin 内有这样一个函数，将生成编码后我们看到的地址格式如：`t1qecjxje6yjq2yatfgj3noapi5fa3cr7vmrw6xti`

```go
func encode(network Network, addr Address) (string, error) {
	if addr == Undef {
		return UndefAddressString, nil
	}
	var ntwk string
	switch network {
	case Mainnet:
		ntwk = MainnetPrefix
	case Testnet:
		ntwk = TestnetPrefix
	default:
		return UndefAddressString, ErrUnknownNetwork
	}

	var strAddr string
	switch addr.Protocol() {
	case SECP256K1, Actor, BLS:
		cksm := Checksum(append([]byte{addr.Protocol()}, addr.Payload()...))
		strAddr = ntwk + fmt.Sprintf("%d", addr.Protocol()) + AddressEncoding.WithPadding(-1).EncodeToString(append(addr.Payload(), cksm[:]...))
	case ID:
		i, n, err := varint.FromUvarint(addr.Payload())
		if err != nil {
			return UndefAddressString, xerrors.Errorf("could not decode varint: %w", err)
		}
		if n != len(addr.Payload()) {
			return UndefAddressString, xerrors.Errorf("payload contains additional bytes")
		}
		strAddr = fmt.Sprintf("%s%d%d", ntwk, addr.Protocol(), i)
	default:
		return UndefAddressString, ErrUnknownProtocol
	}
	return strAddr, nil
}
```

`Network`看作是枚举，`0`代表`Mainnet`，`1`代表`Testnet`。`TestnetPrefix = “t”`，`MainnetPrefix = “f”`。

对地址格式再次做哈希，作为地址摘要

```go
cksm := hash(buf, &blake2b.Config{Size: 4})
```

> ChecksumHashLength defines the hash length used for calculating address checksums.

`ChecksumHashLength`等于4。

对`hash(pub)`+`cksum`做base32转换。

```go
const encodeStd = "abcdefghijklmnopqrstuvwxyz234567"

// AddressEncoding defines the base32 config used for address encoding and decoding.
var AddressEncoding = base32.NewEncoding(encodeStd)
```

最终地址等于`网络前缀`+`椭圆曲线类型`+`base32(hash(pub)+cksum)`





## 离线签名

```go
// 使用secp256k1签名
// 签名前使用blake2哈希算法对要签名的消息做信息摘要，统一签名消息的长度
func (secpSigner) Sign(pk []byte, msg []byte) ([]byte, error) {
	b2sum := blake2b.Sum256(msg)
	sig, err := crypto.Sign(pk, b2sum[:])
	if err != nil {
		return nil, err
	}

	return sig, nil
}
```

`crypto.Sign`

```go
// Sign signs the given message, which must be 32 bytes long.
func Sign(sk, msg []byte) ([]byte, error) {
  // secp256k1 就是最终的椭圆曲线签名，这个签名可以替换成secp256k1不同语言版本的实现
	return secp256k1.Sign(msg, sk)
}
```


