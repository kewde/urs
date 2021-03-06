// Copyright 2014 Hein Meling and Haibin Zhang. All rights reserved.
// Additions made by tecnovert (Particl).
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

const numOfKeys = 1000

var (
	DefaultCurve = elliptic.P256()
	keyring      *PublicKeyRing
	testkey      *ecdsa.PrivateKey
	testmsg      []byte
	testsig      *RingSign
)

func TestGenerateKey(t *testing.T) {
	runtime.GOMAXPROCS(4)
	var err error
	testkey, err = GenerateKey(DefaultCurve, crand.Reader)
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
}

func TestNewPublicKeyRing(t *testing.T) {
	keyring = NewPublicKeyRing(1)
	keyring.Add(testkey.PublicKey)
	expectedLen := 1
	if len(keyring.Ring) != expectedLen {
		t.Errorf("len(keyring)=%d, expected %d", len(keyring.Ring), expectedLen)
	}
}

func TestPopulateKeyRing(t *testing.T) {
	keyring = NewPublicKeyRing(numOfKeys)
	rand.Seed(23)
	k := rand.Intn(numOfKeys)
	fmt.Println("Index of my key: ", k)
	for i := 0; i < numOfKeys; i++ {
		key, err := GenerateKey(DefaultCurve, crand.Reader)
		if err != nil {
			fmt.Println(err.Error())
			t.FailNow()
		}
		if i == k { // designate this as my key
			testkey = key
		}
		// add the public key part to the ring
		keyring.Add(key.PublicKey)
	}
	if len(keyring.Ring) != numOfKeys {
		t.Errorf("len(keyring)=%d, expected %d", len(keyring.Ring), numOfKeys)
	}
}

func TestSign(t *testing.T) {
	testmsg = []byte("Hello, world.")
	var err error
	testsig, err = Sign(crand.Reader, testkey, keyring, testmsg)
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}

	fmt.Printf("testsig.hsx %s\n", testsig.X.String())
	fmt.Printf("testsig.hsy %s\n", testsig.Y.String())

	mR := append(testmsg, keyring.Bytes()...)

	c := keyring.Ring[0].Curve
	h := sha256.New()
	h.Write(mR)
	d := h.Sum(nil)

	fmt.Printf("looping through ring of %d\n", keyring.Len())
	for j := 0; j < keyring.Len(); j++ {

		rx, ry := c.ScalarMult(keyring.Ring[j].X, keyring.Ring[j].Y, d)

		//if testsig.X == rx && testsig.Y == ry {
		if testsig.X.String() == rx.String() && testsig.Y.String() == ry.String() {
			fmt.Printf("Found signing key: %d\nx: %s\ny: %s\n", j, rx.String(), ry.String())
		}
	}

}

func TestVerify(t *testing.T) {
	if !Verify(keyring, testmsg, testsig) {
		fmt.Println("urs: signature verification failed")
		t.FailNow()
	}
}

func BenchmarkSign(b *testing.B) {
	runtime.GOMAXPROCS(8)
	var err error
	for i := 0; i < b.N; i++ {
		testsig, err = Sign(crand.Reader, testkey, keyring, testmsg)
		if err != nil {
			fmt.Println(err.Error())
			b.FailNow()
		}
	}
}

func BenchmarkVerify(b *testing.B) {
	runtime.GOMAXPROCS(8)
	for i := 0; i < b.N; i++ {
		if !Verify(keyring, testmsg, testsig) {
			fmt.Println("urs: signature verification failed")
			b.FailNow()
		}
	}
}
