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
	"log"
)

type PKEntry struct {
	Pk rsa.PublicKey
	IP string
	Port string
	Time int64
}

const KEYDIR = "keys/"
const RECORDPREFIX = "[RCOD]"
var KeyRepo map[string]*PKEntry // map[id:string] entry:PKEntry

/*
	check if there is PKEntry associated with id in memory and on disk.
 */
func GetKeyByID(id string) *PKEntry {
	pe := KeyRepo[id]
	if pe != nil { return pe }
	if yes,_:=exists(KEYDIR+id); yes {
		dat, err := ioutil.ReadFile(KEYDIR+id)
		checkErr(err)
		pe = BytesToPKEntry(dat)
		return pe
	}
	return nil
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
	print("Generating Key Repo")
	KeyRepo = make(map[string]*PKEntry)
	if yes, _ := exists(KEYDIR); !yes {
		os.Mkdir(KEYDIR, 0777)
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
	checkErr(err)
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
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	checkErr(err)
	fmt.Fprintf(file, "%s", pe.Bytes())
}

/*
	Read all saved PKEntry from disk.
 */
func populatePKEntry() {
	dir := KEYDIR
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		checkErr(err)
		dat, err := ioutil.ReadFile(path)
		id := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
		pe := BytesToPKEntry(dat)
		if pe != nil {
			KeyRepo[id] = pe
		}
		return nil
	})
	checkErr(err)
}


func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	fmt.Print(RECORDPREFIX+" ")
	fmt.Println(str...)
}