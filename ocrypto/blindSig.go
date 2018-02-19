package ocrypto

import (
	"crypto/rsa"
	"math/big"
	"crypto/rand"
	"io"
)

var bigZero = big.NewInt(0)
var bigOne = big.NewInt(1)

/*
	blind the data with blind factor randomly selected from [0,pub-key's N)
	return blindeddata, blindfactor
 */
func Blind(key *rsa.PublicKey, data []byte) ([]byte, []byte) {
	blinded, bfactor, err := blind(rng, key, new(big.Int).SetBytes(data))
	if err != nil {
		panic(err)
	}
	return blinded.Bytes(), bfactor.Bytes()
}

/*
	removing the blind factor
 */
func Unblind(key *rsa.PublicKey, blindedSig, bfactor []byte) []byte {
	m := new(big.Int).SetBytes(blindedSig)
	bfactorBig := new(big.Int).SetBytes(bfactor)
	m.Mul(m, bfactorBig)
	m.Mod(m, key.N)
	return m.Bytes()
}

/*
	Blind signing, which is a bit different from crypto/rsa.Sign
 */
func BlindSign(key *rsa.PrivateKey, data []byte) []byte {
	c := new(big.Int).SetBytes(data)
	m, err := Decrypt(rand.Reader, key, c)
	checkErr(err)
	return m.Bytes()
}

func VerifyBlindSig(key *rsa.PublicKey, data, sig []byte) bool {
	m := new(big.Int).SetBytes(data)
	bigSig := new(big.Int).SetBytes(sig)
	c := Encrypt(new(big.Int), key, bigSig)
	return m.Cmp(c) == 0
}

/*
	generate blind factor and blinded data based on pub-key's N
 */
func blind(random io.Reader, key *rsa.PublicKey, c *big.Int) (blinded, unblinder *big.Int, err error) {
	var r *big.Int

	for {
		r, err = rand.Int(random, key.N)
		if err != nil {
			return
		}
		if r.Cmp(bigZero) == 0 {
			r = bigOne
		}
		ir, ok := modInverse(r, key.N)
		if ok {
			bigE := big.NewInt(int64(key.E))
			rpowe := new(big.Int).Exp(r, bigE, key.N)
			cCopy := new(big.Int).Set(c)
			cCopy.Mul(cCopy, rpowe)
			cCopy.Mod(cCopy, key.N)
			return cCopy, ir, nil
		}
	}
}

/*
	from go's crypto/rsa, computing inverse of a mode n.
 */
func modInverse(a, n *big.Int) (ia *big.Int, ok bool) {
	g := new(big.Int)
	x := new(big.Int)
	y := new(big.Int)
	g.GCD(x, y, a, n)
	if g.Cmp(bigOne) != 0 {
		return
	}

	if x.Cmp(bigOne) < 0 {
		x.Add(x, n)
	}

	return x, true
}

/*
	from go's crypto/rsa, standard rsa encryption of m	m^e mod pub.N
 */
func Encrypt(cipher *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	e := big.NewInt(int64(pub.E))
	cipher.Exp(m, e, pub.N)
	return cipher
}

func Decrypt(random io.Reader, priv *rsa.PrivateKey, c *big.Int) (m *big.Int, err error) {
	if c.Cmp(priv.N) > 0 {
		err = rsa.ErrDecryption
		return
	}

	var ir *big.Int
	if random != nil {
		var r *big.Int
		for {
			r, err = rand.Int(random, priv.N)
			if err != nil {
				return
			}
			if r.Cmp(bigZero) == 0 {
				r = bigOne
			}
			var ok bool
			ir, ok = modInverse(r, priv.N)
			if ok {
				break
			}
		}
		bigE := big.NewInt(int64(priv.E))
		rpowe := new(big.Int).Exp(r, bigE, priv.N)
		cCopy := new(big.Int).Set(c)
		cCopy.Mul(cCopy, rpowe)
		cCopy.Mod(cCopy, priv.N)
		c = cCopy
	}

	if priv.Precomputed.Dp == nil {
		m = new(big.Int).Exp(c, priv.D, priv.N)
	} else {
		// We have the precalculated values needed for the CRT.
		m = new(big.Int).Exp(c, priv.Precomputed.Dp, priv.Primes[0])
		m2 := new(big.Int).Exp(c, priv.Precomputed.Dq, priv.Primes[1])
		m.Sub(m, m2)
		if m.Sign() < 0 {
			m.Add(m, priv.Primes[0])
		}
		m.Mul(m, priv.Precomputed.Qinv)
		m.Mod(m, priv.Primes[0])
		m.Mul(m, priv.Primes[1])
		m.Add(m, m2)

		for i, values := range priv.Precomputed.CRTValues {
			prime := priv.Primes[2+i]
			m2.Exp(c, values.Exp, prime)
			m2.Sub(m2, m)
			m2.Mul(m2, values.Coeff)
			m2.Mod(m2, prime)
			if m2.Sign() < 0 {
				m2.Add(m2, prime)
			}
			m2.Mul(m2, values.R)
			m.Add(m, m2)
		}
	}

	if ir != nil {
		// Unblind.
		m.Mul(m, ir)
		m.Mod(m, priv.N)
	}

	return
}
