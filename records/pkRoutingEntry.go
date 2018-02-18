package records

import (
	"crypto/rsa"
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"strings"
)

type PKEntry struct {
	Pk rsa.PublicKey
	IP string
	Port string
	Time int64
}

const KEYDIR = "keys/"
var KeyRepo map[string]*PKEntry // map[id:string] entry:PKEntry

func GetKeyByID(id string) *PKEntry {
	return KeyRepo[id]
}

func InsertEntry(id string, pk rsa.PublicKey, recTime int64, ip string, port string) {
	if pe := KeyRepo[id]; pe != nil {
		if pe.Time < recTime {
			pe.IP = ip
			pe.Port = port
			pe.Time = recTime
			writePE(pe, id)
		}
	} else {
		e := &PKEntry{pk, ip, port, recTime}
		KeyRepo[id] = e
		writePE(e, id)
	}
}

/*
	read key repo blockchain file from disk and load entries into KeyRepo
 */
func GenerateKeyRepo() {
	KeyRepo = make(map[string]*PKEntry)
	if yes, _ := exists(KEYDIR); !yes {
		os.Mkdir(KEYDIR, 0600)
	}
	populatePKEntry()
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
		//fmt.Println(err)
		return nil
	}
	return e
}

/*
	Write a PKEntry to disk.
 */
func writePE(pe *PKEntry, id string) {
	path := KEYDIR+id
	os.Remove(path)
	os.Create(path)
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	defer file.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Fprintf(file, "%s", pe.Bytes())
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

/*
	Read all saved PKEntry from disk.
 */
func populatePKEntry() {
	dir := KEYDIR
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", dir, err)
			return err
		}

		dat, err := ioutil.ReadFile(path)
		id := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
		pe := BytesToPKEntry(dat)
		if pe != nil {
			KeyRepo[id] = pe
		}
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dir, err)
	}
}