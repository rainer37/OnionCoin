package util

import (
	"os"
	"crypto/sha256"
)

const (
	LOCALHOST = "127.0.0.1"

	IDLEN = 16

	NUMCOSIGNER = 2

	EPOCHLEN = 10
	PROPOSINGDELAY = 5
	PUSHINGDELAY = 3
	PROPOSINGTIME = EPOCHLEN - PROPOSINGDELAY
	PUSHTIME = EPOCHLEN - PUSHINGDELAY

	MAXNUMTXN = 500
	MATUREDIFF = 2

 	RSAKEYLEN = 1024
 	CIPHERKEYLEN = RSAKEYLEN / 8
	SYMKEYLEN = 128
)

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func CheckErr(err error){
	if err != nil { panic(err) }
}

func Contains(ids []string, id string) bool {
	for _, v := range ids { if v == id { return true } }; return false
}

func ShaHash(b []byte) [32]byte {
	return sha256.Sum256(b)
}