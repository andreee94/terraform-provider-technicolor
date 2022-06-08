package technicolor

import (
	"bytes"
	"log"
	"math/big"
)

// class
type CryptoUser struct {
	Username      string
	Password      string
	HashAlgorithm int
	NGType        int
	internal      *CryptoUserInternal
	state         *CryptoUserState
}

type CryptoUserInternal struct {
	N       *big.Int
	g       int64
	a       *big.Int
	A       *big.Int
	k       *big.Int
	M       []byte
	HashAMK []byte
}

type CryptoUserState struct {
	Authenticated bool
}

func NewCryptoUser(username string, password string, hashAlgorithm int, ngType int) *CryptoUser {
	a := GenerateRandomBigIntFirstBit(32)
	return NewCryptoUserWithA(username, password, hashAlgorithm, ngType, a)
}

func NewCryptoUserWithA(username string, password string, hashAlgorithm int, ngType int, a *big.Int) *CryptoUser {

	N, g := GetNG(ngType)

	if N == nil || g == 0 {
		log.Fatalf("Invalid N(%v) or g(%v)", N, g)
	}

	// a := GenerateRandomBigIntFirstBit(32)

	// // A = ModPow(g, a, N)
	A := new(big.Int).Exp(big.NewInt(int64(g)), a, N)

	// // k = int("05b9e8ef059c6b32ea59fc1d322d37f04aa30bae5aa9003b8321e21ddb04e300", 16)
	k := bigIntFromStringHex("05b9e8ef059c6b32ea59fc1d322d37f04aa30bae5aa9003b8321e21ddb04e300")

	return &CryptoUser{
		Username:      username,
		Password:      password,
		HashAlgorithm: hashAlgorithm,
		NGType:        ngType,
		internal: &CryptoUserInternal{
			N: N,
			g: g,
			A: A,
			k: k,
			a: a,
		},
		state: &CryptoUserState{
			Authenticated: false,
		},
	}
}

func (c *CryptoUser) StartAuthentication() (string, []byte) {
	return c.Username, LongToBytes(c.internal.A)
}

func (c *CryptoUser) IsAuthenticated() bool {
	return c.state.Authenticated
}

func (c *CryptoUser) ComputeSha(content ...[]byte) []byte {
	return ComputeSha(bytes.Join(content, []byte{}), c.HashAlgorithm)
}

func (c *CryptoUser) ComputeHashAMK(K []byte) []byte {
	return c.ComputeSha(
		LongToBytes(c.internal.A),
		c.internal.M,
		K,
	)
}

func (c *CryptoUser) ComputeM(sChallenge []byte, BChallenge []byte, K []byte) []byte {
	return c.ComputeSha(
		ComputeXorHashNG(c.HashAlgorithm, c.internal.N, c.internal.g),
		c.ComputeSha([]byte(c.Username)),
		sChallenge,
		LongToBytes(c.internal.A),
		BChallenge,
		K,
	)
}

func (c *CryptoUser) ComputeX(sChallenge []byte) []byte {
	return c.ComputeSha(
		sChallenge,
		c.ComputeSha([]byte(c.Username+":"+c.Password)),
	)
}

func (c *CryptoUser) ProcessChallenge(sBytes []byte, BBytes []byte) ([]byte, []byte) {
	// s := BytesToLong(sBytes)
	B := BytesToLong(BBytes)

	// SRP-6a safety check
	if new(big.Int).Mod(B, c.internal.N).Cmp(big.NewInt(0)) == 0 {
		return nil, nil
	}

	u := BytesToLong(c.ComputeSha(
		LongToBytes(c.internal.A),
		BBytes,
	))

	// SRP-6a safety check
	if u.Cmp(big.NewInt(0)) == 0 {
		return nil, nil
	}

	x := BytesToLong(c.ComputeX(sBytes))

	v := new(big.Int).Exp(big.NewInt(c.internal.g), x, c.internal.N)

	// self.S = pow((self.B - k * self.v), (self.a + self.u * self.x), N)
	S := new(big.Int).Exp(
		new(big.Int).Sub(
			B,
			new(big.Int).Mul(
				c.internal.k,
				v,
			),
		),
		new(big.Int).Add(
			c.internal.a,
			new(big.Int).Mul(
				u,
				x,
			)),
		c.internal.N,
	)

	K := c.ComputeSha(LongToBytes(S))

	c.internal.M = c.ComputeM(sBytes, BBytes, K)

	c.internal.HashAMK = c.ComputeHashAMK(K)

	return c.internal.M, c.internal.HashAMK
}

func (c *CryptoUser) ValidateAuthentication(HostHashHMK []byte) bool {
	if bytes.Equal(c.internal.HashAMK, HostHashHMK) {
		c.state.Authenticated = true
	}
	return c.IsAuthenticated()
}
