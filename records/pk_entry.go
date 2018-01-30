package records

import (
	"crypto/rsa"
	"time"
	"crypto/sha256"
)

type PKEntry struct {
	pk rsa.PublicKey
	keyHash []byte
	time time.Time
}

var KeyRepo map[string]*PKEntry // map[id:string] entry:PKEntry

func GetKeyByID(id string) rsa.PublicKey {
	return KeyRepo[id].pk
}

func GetKeyHashByID(id string) []byte {
	return KeyRepo[id].keyHash
}

func GetTimeByID(id string) time.Time {
	return KeyRepo[id].time
}

func InsertEntry(id string, pk rsa.PublicKey, recTime time.Time) {
	e := new(PKEntry)
	h := sha256.New()
	h.Write(pk.N.Bytes())
	e.keyHash = h.Sum(nil)
	e.pk = pk
	e.time = recTime
	KeyRepo[id] = e
}

/*
	read key repo blockchain file from disk and load entries into KeyRepo
 */
func GenerateKeyRepo(regfilename string) {
	KeyRepo = make(map[string]*PKEntry)
}
