package util

import (
	"os"
	"crypto/sha256"
	"bytes"
	"encoding/binary"
	"time"
)

const (
	LOCALHOST = "127.0.0.1"

	IDLEN = 16

	NUMCOSIGNER = 2
	COSIGNTIMEOUT = 3

	EPOCHLEN = 10
	PROPOSINGDELAY = 5
	PUSHINGDELAY = 3
	DECISIONDELAY = 2
	PROPOSINGTIME = EPOCHLEN - PROPOSINGDELAY
	PUSHTIME = EPOCHLEN - PUSHINGDELAY

	MAXNUMTXN = 500
	MATUREDIFF = 2
	SIGL = 128

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
func CheckErr(err error){ if err != nil { panic(err) } }
func Contains(ids []string, id string) bool {
	for _, v := range ids { if v == id { return true } }; return false
}
func Sha(b []byte) [32]byte { return sha256.Sum256(b) }
func Strip(b []byte) string { return string(bytes.Trim(b, "\x00")) }
func JoinBytes(bs [][]byte) []byte { return bytes.Join(bs, []byte{}) }
func SortSigs(sigs []byte, verifiers []string) {
	for i:=0; i<len(verifiers) - 1; i++ {
		for j:=0; j<len(verifiers) -i - 1; j++ {
			if verifiers[j] > verifiers[j+1] {
				verifiers[j+1], verifiers[j] = verifiers[j], verifiers[j+1]
				temp := make([]byte, SIGL)
				copy(temp, sigs[(j+1) * SIGL:(j+2) * SIGL])
				copy(sigs[(j+1) * SIGL:(j+2) * SIGL],sigs[j*SIGL:(j+1)*SIGL])
				copy(sigs[j*SIGL:(j+1)*SIGL], temp)
			}
		}
	}
}

func SplitBytes(b []byte, n int) [][]byte {
	if len(b) / n > 40 {
		print("msg too long cannot split")
		return nil
	}
	bs := [][]byte{}
	div := len(b) / n
	for i:=0;i<n;i++ {
		bs = append(bs, b[i*div:(i+1)*div])
	}
	if len(b) % n != 0 { bs = append(bs, b[n*div:]) }
	return bs
}

func CurTSBytes() []byte {
	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))
	return tsBytes
}

func NewBytes(n int, data []byte) []byte {
	nb := make([]byte, n)
	copy(nb, data)
	return nb
}