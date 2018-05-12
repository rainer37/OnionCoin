package records

import (
	"crypto/rsa"
	"fmt"
	"os"
	"io/ioutil"
	"math/rand"
	"sync"
	"github.com/rainer37/OnionCoin/util"
	"encoding/json"
)

type PKEntry struct {
	Pk rsa.PublicKey
	IP string
	Port string
	Time int64
}

const KEYDIR = "keys/"
const RECORDPREFIX = "[RCOD]"

var keyRepo map[string]*PKEntry // map[id:string] entry:PKEntry
var mete = sync.RWMutex{}

/*
	read key repo blockchain file from disk and load entries into keyRepo
 */
func GenerateKeyRepo() {
	InitKeyRepo()
	if yes, _ := util.Exists(KEYDIR); !yes {
		os.Mkdir(KEYDIR, 0777)
	}
	PopulatePKEntry()
}

func InitKeyRepo() {
	keyRepo = make(map[string]*PKEntry)
}

/*
	Read all saved PKEntry from disk.
 */
func PopulatePKEntry() {
	files, err := ioutil.ReadDir(KEYDIR)
	util.CheckErr(err)
	for _, f := range files {
		err := getPEonDisk(f.Name())
		util.CheckErr(err)
	}
}

/*
	Read a PE from disk by its file name/pe_id.
 */
func getPEonDisk(id string) error {
	dat, err := ioutil.ReadFile(KEYDIR + id)
	util.CheckErr(err)
	pe := BytesToPKEntry(dat)
	insertPE(id, pe)
	return err
}

/*
	check if there is PKEntry associated with id in memory and on disk.
 */
func GetKeyByID(id string) *PKEntry {
	pe := getPE(id)
	if pe != nil { return pe }
	if yes , _ := util.Exists(KEYDIR+id); yes {
		err := getPEonDisk(id)
		util.CheckErr(err)
		return getPE(id)
	}
	return nil
}

func insertPE(id string, pe *PKEntry) {
	mete.Lock()
	defer mete.Unlock()
	keyRepo[id] = pe
}

func getPE(id string) *PKEntry {
	mete.Lock()
	defer mete.Unlock()
	return keyRepo[id]
}

func InsertEntry(id string, pk rsa.PublicKey, recTime int64, ip string, port string) {
	pe := new(PKEntry)
	pe.Pk = pk
	pe.IP = ip
	pe.Port = port
	pe.Time = recTime

	insertPE(id, pe)
	WritePE(pe, id)
}

/*
	json encode PKEntry to bytes
 */
func (e PKEntry) Bytes() []byte {
	b, err := json.Marshal(e)
	util.CheckErr(err)
	return b
}

/*
	decode json bytes into PKEntry
 */
func BytesToPKEntry(data []byte) *PKEntry {
	pe := new(PKEntry)
	err := json.Unmarshal(data, pe)
	util.CheckErr(err)
	return pe
}

/*
	(Over)Write a PKEntry to disk.
 */
func WritePE(pe *PKEntry, id string) {
	path := KEYDIR + id
	os.Remove(path)
	os.Create(path)
	file, err := os.OpenFile(path, os.O_RDWR, 0777)
	defer file.Close()
	util.CheckErr(err)
	fmt.Fprintf(file, "%s", pe.Bytes())
}

func KeyRepoSize() int { return len(keyRepo) }

/*
	Gather multiple PEntry and format them into json.
 */
func PackPEs(num int) []byte {
	pes := make(map[string]PKEntry)
	i := 0
	for id, v := range keyRepo {
		if i >= num { break }
		pes[id] = *v
		i++
	}
	b, err := json.Marshal(pes)
	util.CheckErr(err)
	return b
}

func UnpackPEs(b []byte) {
	pes := map[string]PKEntry{}
	err := json.Unmarshal(b, &pes)
	util.CheckErr(err)
	for i,v := range pes {
		InsertEntry(i, v.Pk, v.Time, v.IP, v.Port)
	}
}

/*
	Get all PEs except the ones in ids.
 */
func AllPEs(ids []string) []*PKEntry {
	var pes []*PKEntry
	for i, v := range keyRepo {
		if !util.Contains(ids, i) {
			pes = append(pes, v)
		}
	}
	return pes
}

/*
	get all ids in my key repo.
 */
func AllIDs() (ids []string) {
	for i := range keyRepo { ids = append(ids, i) }
	return
}

func RandomPath() (path []string) {
	count := 0
	num := rand.Int() % 2 + 2
	ids := AllIDs()
	for count < num {
		index := rand.Int() % len(keyRepo)
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