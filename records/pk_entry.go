package records

import (
	"crypto/rsa"
	"bytes"
	"encoding/gob"
	"fmt"
)

type PKEntry struct {
	Pk rsa.PublicKey
	IP string
	Port string
	Time int64
}

var KeyRepo map[string]*PKEntry // map[id:string] entry:PKEntry

func GetKeyByID(id string) *PKEntry {
	return KeyRepo[id]
}

func InsertEntry(id string, pk rsa.PublicKey, recTime int64, ip string, port string) {
	e := new(PKEntry)
	e.Pk = pk
	e.Time = recTime
	e.IP = ip
	e.Port = port
	KeyRepo[id] = e
}

/*
	read key repo blockchain file from disk and load entries into KeyRepo
 */
func GenerateKeyRepo(regfilename string) {
	KeyRepo = make(map[string]*PKEntry)
}

/*
	god encode PKEntry to bytes
 */
func (e PKEntry) Bytes() []byte {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(PKEntry{e.Pk, e.IP, e.Port, e.Time})
	if err != nil {
		fmt.Println(err)
	}
	return b.Bytes()
}

/*
	decode bytes into PKEntry
 */
func BytesToPKEntry(data []byte) *PKEntry {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	e := new(PKEntry)
	err := dec.Decode(e)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return e
}