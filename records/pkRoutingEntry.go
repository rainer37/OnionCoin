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
	"math/rand"
	"sync"
	"github.com/rainer37/OnionCoin/util"
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
var mete = sync.RWMutex{}

/*
	check if there is PKEntry associated with id in memory and on disk.
 */
func GetKeyByID(id string) *PKEntry {
	mete.Lock()
	defer mete.Unlock()
	pe := KeyRepo[id]
	if pe != nil { return pe }
	if yes , _ := util.Exists(KEYDIR+id); yes {
		dat, err := ioutil.ReadFile(KEYDIR+id)
		util.CheckErr(err)
		pe = BytesToPKEntry(dat)
		return pe
	}
	return nil
}

func InsertEntry(id string, pk rsa.PublicKey, recTime int64, ip string, port string) {
	mete.Lock()
	defer mete.Unlock()

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
	if yes, _ := util.Exists(KEYDIR); !yes {
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
	util.CheckErr(err)
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
	util.CheckErr(err)
	fmt.Fprintf(file, "%s", pe.Bytes())
}

/*
	Read all saved PKEntry from disk.
 */
func populatePKEntry() {
	dir := KEYDIR
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		util.CheckErr(err)
		dat, err := ioutil.ReadFile(path)
		id := strings.Split(path, "/")[len(strings.Split(path, "/"))-1]
		pe := BytesToPKEntry(dat)
		if pe != nil {
			KeyRepo[id] = pe
		}
		return nil
	})
	util.CheckErr(err)
}

/*
	get all ids in my key repo.
 */
func allIDs() (ids []string) {
	for i := range KeyRepo { ids = append(ids, i) }
	return
}

func RandomPath() (path []string) {
	count := 0
	num := rand.Int() % 2 + 2
	ids := allIDs()
	for count < num {
		index := rand.Int() % len(KeyRepo)
		id := ids[index]
		if !util.Contains(path, id) {
			path = append(path, id)
			count++
		}
	}
	return
}

func print(str ...interface{}) {
	fmt.Print(RECORDPREFIX+" ")
	fmt.Println(str...)
}