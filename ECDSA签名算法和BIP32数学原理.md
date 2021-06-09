#  ECDSA签名算法和HDWallet数学原理

## golang对于ecdsa算法的实现

### 简述

椭圆曲线算法, 
就是在椭圆曲线上的一系列的离散的有限的点, 并且定义了一个虚拟的0点(原点), 逆元, 加法和乘法二元运算
并且这些二元运算满足加法交换律和结合律.

这些点形成组成了一个有限域, 称为**[阿贝尔群](https://zh.wikipedia.org/wiki/%E9%98%BF%E8%B4%9D%E5%B0%94%E7%BE%A4)**.

### 私钥生成:randFieldElement

```golang
func randFieldElement(c elliptic.Curve, rand io.Reader) (k *big.Int, err error) {
	params := c.Params()
	b := make([]byte, params.BitSize/8+8)
	_, err = io.ReadFull(rand, b)
	if err != nil {
		return
	}

	k = new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(params.N, one)
	k.Mod(k, n)
	k.Add(k, one)
	return
}
```

randFieldElement作用是使用 *curve paramater* 来生成一个新的私钥k.
参数***c***是***curve paramater***(或者叫 ***domain parameters***)定义了在有限域中的椭圆曲线的阿贝尔群.

> Our elliptic curve algorithms will work in a cyclic subgroup of an elliptic curve over a finite field. Therefore, our algorithms will need the following parameters:
>
> - The **prime p** that specifies the size of the finite field.
> - The **coefficients a and b** of the elliptic curve equation.
> - The **base point G** that generates our subgroup.
> - The **order n** of the subgroup.
> - The **cofactor h** of the subgroup. ($ h = N/n $,其中$ N $ 是椭圆曲线的阶数)
>
> In conclusion, the **domain parameters** for our algorithms are the **sextuple (p,a,b,G,n,h)**.

### 生成签名:signGeneric

```golang
func signGeneric(pk *PrivateKey, csprng *cipher.StreamReader, c elliptic.Curve, hash []byte) (r, s *big.Int, err error) {
	N := c.Params().N
	if N.Sign() == 0 {
		return nil, nil, errZeroParam
	}
	var k, kInv *big.Int
	for {
		for {
			k, err = randFieldElement(c, *csprng)
			if err != nil {
				r = nil
				return
			}
			if in, ok := pk.Curve.(invertible); ok {
				kInv = in.Inverse(k)
			} else {
				kInv = fermatInverse(k, N) // N != 0
			}
			r, _ = pk.Curve.ScalarBaseMult(k.Bytes())
			r.Mod(r, N)
			if r.Sign() != 0 {
				break
			}
		}
		e := hashToInt(hash, c)
		s = new(big.Int).Mul(pk.D, r)
		s.Add(s, e)
		s.Mul(s, kInv)
		s.Mod(s, N) // N != 0
		if s.Sign() != 0 {
			break
		}
	}
	return
}
```

签名函数返回值包含两部分内容, sig = (r, s)
- k为临时生成的私钥
- e为签名数据的hash值的整数形式
- R = k * G, 所以R为临时私钥k的公钥
- 函数 Inverse, fermatInverse为由doman paramater(secp256k1)定义的椭圆曲线有限域定义的计算逆元的代数实现.
- 函数 ScalarBaseMult 为有限域上的乘法的代数实现
- pk 为私钥, P为公钥

代码中的两层for循环,是因为临时生成的私钥不满足条件(**具体原因后面在说**),for循环会重新再次随机生成临时私钥对于,99.99%的情况是,这个for循环只会执行一次.
所以根据代码可以把计算 *r* 和 *s* 的代数表达式简单的写成:

$\begin{aligned}
r = R.x
\end{aligned}$

**即 r 的几何意义为:临时私钥k的公钥R[ = k * G 为椭圆曲线上的一个点]的x坐标.**

$\begin{aligned}
s = (e + r * pk) / k
\end{aligned}$

### 验证签名:verifyGeneric

```go
func verifyGeneric(pub *PublicKey, c elliptic.Curve, hash []byte, r, s *big.Int) bool {
	e := hashToInt(hash, c)
	var w *big.Int
	N := c.Params().N
	if in, ok := c.(invertible); ok {
		w = in.Inverse(s)
	} else {
		w = new(big.Int).ModInverse(s, N)
	}
	u1 := e.Mul(e, w)
	u1.Mod(u1, N)
	u2 := w.Mul(r, w)
	u2.Mod(u2, N)
	// Check if implements S1*g + S2*p
	var x, y *big.Int
	if opt, ok := c.(combinedMult); ok {
		x, y = opt.CombinedMult(pub.X, pub.Y, u1.Bytes(), u2.Bytes())
	} else {
		x1, y1 := c.ScalarBaseMult(u1.Bytes())
		x2, y2 := c.ScalarMult(pub.X, pub.Y, u2.Bytes())
		x, y = c.Add(x1, y1, x2, y2)
	}
	if x.Sign() == 0 && y.Sign() == 0 {
		return false
	}
	x.Mod(x, N)
	return x.Cmp(r) == 0
}
```

- 函数***ScalarBaseMult***是椭圆曲线有限域上n * G的乘法, 其中n为参数
- 函数***ScalarMult*** 是椭圆曲线有限域上定义的n * P的乘法, 第1,2个参数表示P的x和y坐标, 第3个参数为n.

根据函数的实现可以写出验证代数表达式, 并执行推导出如下结果:

$\begin{aligned} 
e*G/s + r*Pub/s &= e*G/s + r*Pk*G/s \\
&= (e+r*Pk)*G/s \\
&= ((e + r * Pk) * G)\ /\  ((e + r * Pk) / k) \\
&= k * G \\
&= R
\end{aligned}$

- e为签名数据hash值的整数形式

- G为Domain parameters的中定义的椭圆曲线的生成点

- k 为在signGeneric函数中生成的临时私钥

- 传入的参数r, 为k的公钥的x坐标

- R(代数表达式最后推出的结果), 就是k的公钥

**函数verifyGeneric最后`x.Cmp(r)==0`就是比较上面的代数表达式推算出的的R.x(k的公钥的x坐标)和signGeneric(签名函数)返回的r(临时私钥的公钥的x坐标)是否相等来判断签名是否验证成功的.**

## 分层确定钱包<sub>Hierarchical Deterministic Wallet</sub>

分层确定钱包的详细描述及相关细节在[\<<master bitcoin-HD Wallets (BIP-32/BIP-44)\>>](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch05.asciidoc#hd-wallets-bip-32bip-44) 和 [\<<BIP32\>>](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki)中已经有非常详细的说明.

这里不再重复这些内容, 其中, 分层确定确定钱包有一个非常重要的特性:
> A very useful characteristic of HD wallets is the ability to derive public child keys from public parent keys, *without* having the private keys.
> HD wallets 一个非常有用的特性是:不需要知道父私钥,就能够通过父公钥派生出子公钥.

这个特性是分层确定钱包最奇妙的地方, **这一章节就是来讲清楚HD wallet这个特性背后的数学原理.**

首先定义代表某些计算的符号如下:

>- point(p): returns the coordinate pair resulting from EC point multiplication (repeated application of the EC group operation) of the secp256k1 base point with the integer p.
>- ser32(i): serialize a 32-bit unsigned integer i as a 4-byte sequence, most significant byte first.
>- ser256(p): serializes the integer p as a 32-byte sequence, most significant byte first.
>- serP(P): serializes the coordinate pair P = (x,y) as a byte sequence using SEC1's compressed form: (0x02 or 0x03) || ser256(x), where the header byte depends on the parity of the omitted y coordinate.
>- parse256(p): interprets a 32-byte sequence as a 256-bit number, most significant byte first.

### 父扩展私钥派生子扩展私钥

> The function CKDpriv((k<sub>par</sub>, c<sub>par</sub>), i) &rarr; (k<sub>i</sub>, c<sub>i</sub>) computes a child extended private key from the parent extended private key:
> - Check whether i ≥ 2<sup>31</sup> (whether the child is a hardened key).
>   -  If so (hardened child): let I = HMAC-SHA512(Key = c<sub>par</sub>, Data = 0x00 || ser<sub>256</sub>(k<sub>par</sub>) || ser<sub>32</sub>(i)). (Note: The 0x00 pads the private key to make it 33 bytes long.)
>   - If not (normal child): let I = HMAC-SHA512(Key = c<sub>par</sub>, Data = ser<sub>P</sub>(point(k<sub>par</sub>)) || ser<sub>32</sub>(i)).
> - Split I into two 32-byte sequences, I<sub>L</sub> and I<sub>R</sub>.
> - The returned child key k<sub>i</sub> is parse<sub>256</sub>(I<sub>L</sub>) + k<sub>par</sub> (mod n).
> - The returned chain code c<sub>i</sub> is I<sub>R</sub>.
> - In case parse<sub>256</sub>(I<sub>L</sub>) ≥ n or k<sub>i</sub> = 0, the resulting key is invalid, and one should proceed with the next value for i. (Note: this has probability lower than 1 in 2<sup>127</sup>.)

**函数CKDpriv为派生子私钥的函数, 参数和返回值解释为:K<sub>par</sub>(父私钥),C<sub>par</sub>(父链码),K<sub>i</sub>子私钥, C<sub>i</sub>(子链码)**

**其中point(k<sub>par</sub>) = k<sub>par</sub> * G = 父公钥, 记为k_pub<sub>par</sub>**

为了突出重点, 这里把上面的计算过程精简一为下面的过程:

1.  HMAC-SHA512(Key = c<sub>par</sub>, Data = ser<sub>P</sub>(point(k<sub>par</sub>)) 得到64个字节的数组
2. 64个字节前32位作为子链码
3. 后32位作为临时私钥k<sub>ephemeral</sub>
4. k<sub>ephemeral</sub> + k<sub>par</sub> 作为子私钥记为:k<sub>child</sub>
5. 根据子私钥可以通过( k<sub>child</sub> * G)计算出子公钥记为**k_pub<sub>child</sub>**

### 父扩展公钥派生子扩展公钥 ###

>The function CKDpub((K<sub>par</sub>, c<sub>par</sub>), i) &rarr; (K<sub>i</sub>, c<sub>i</sub>) computes a child extended public key from the parent extended public key. It is only defined for non-hardened child keys.
>* Check whether i ≥ 2<sup>31</sup> (whether the child is a hardened key).
>** If so (hardened child): return failure
>** If not (normal child): let I = HMAC-SHA512(Key = c<sub>par</sub>, Data = ser<sub>P</sub>(K<sub>par</sub>) || ser<sub>32</sub>(i)).
>* Split I into two 32-byte sequences, I<sub>L</sub> and I<sub>R</sub>.
>* The returned child key K<sub>i</sub> is point(parse<sub>256</sub>(I<sub>L</sub>)) + K<sub>par</sub>.
>* The returned chain code c<sub>i</sub> is I<sub>R</sub>.
>* In case parse<sub>256</sub>(I<sub>L</sub>) ≥ n or K<sub>i</sub> is the point at infinity, the resulting key is invalid, and one should proceed with the next value for i.

同样, 为了突出重点, 把上面描述的过程精简为下面的过程:

1. HMAC-SHA512(Key = c<sub>par</sub>, Data = ser<sub>P</sub>(point(k<sub>par</sub>)) 得到64个字节的数组
2. 64个字节的前32个字节作为子链码
3. 后32位作为临时私钥k<sub>ephemeral</sub>
4. 然后计算k<sub>ephemeral </sub>* G = k_pub<sub>ephemeral</sub> 为临时公钥
5. 然后计算  k_pub<sub>ephemeral</sub> + k_pub<sub>par</sub> 作为子公钥 = k_pub<sub>child</sub>

又由于椭圆曲线上的点是一个阿贝尔群, 满足加法交换律和结合律, 可以有下面的推导过程:
$\begin{aligned}
k\_pub_{ephemeral} + k\_pub_{par} &= k_{ephemeral} * G + k_{par} * G \\
&= (k_{ephemeral} + k_{par}) * G \\
&= k_{child} * G \\
&= k\_pub_{child} \\ 
\end{aligned}$

**这就是为什么HD Wallet只需要暴露扩展公钥就能推测出子私钥地址的原因**.

### 分层确定钱包的风险
分层确定钱包的风险请参考这篇文章:[Private Key Recovery Combination Attacks](https://github.com/zl03jsj/hdwallet/blob/master/Private%20Key%20Recovery%20Combination%20Attacks.pdf)

## 参考引用

[Elliptic Curve Cryptography: a gentle introduction](https://andrea.corbellini.name/2015/05/17/elliptic-curve-cryptography-a-gentle-introduction/)
[ecdsa math](https://github.com/bitcoinbook/bitcoinbook/blob/develop/ch06.asciidoc#creating-a-digital-signature)
[What is modular arithmetic](https://www.khanacademy.org/computing/computer-science/cryptography/modarithmetic/a/what-is-modular-arithmetic)
[模运算](https://blog.sengxian.com/algorithms/mod-world)
[椭圆曲线加密算法](https://zhuanlan.zhihu.com/p/101907402)
[ECC椭圆曲线详解](https://www.cnblogs.com/Kalafinaian/p/7392505.html)
[bitcoin extendedkey 源码](https://github.com/btcsuite/btcutil/blob/faeebcb9abbed8d21aa424d0447af576a72e1b8e/hdkeychain/extendedkey.go#L229)
