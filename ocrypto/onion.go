package ocrypto

import (
	"strconv"
	"fmt"
)

type Onion struct {
	isHeart	bool
	nextID string
	len int
	content []byte
}

type OnionMaker struct {}

func (o *Onion) toByte() []byte {
	buffer := make([]byte,0)
	buffer = strconv.AppendBool(buffer,o.isHeart)
	buffer = append(buffer, []byte(o.nextID)...)
	buffer = strconv.AppendInt(buffer, int64(o.len), 10)
	buffer = append(buffer, []byte(o.content)...)
	return buffer
}

func (o *OnionMaker) MakeOnion() *Onion { return nil }
func (o *OnionMaker) wrap(pk []byte, nextID string, len int, onion *Onion) (oret *Onion) {
	onionByte := onion.toByte()
	fmt.Println(onionByte)
	oret.len = len
	oret.nextID = nextID
	oret.content = onionByte
	return
}

func (o *OnionMaker) peel (onionBytes []byte) (oret *Onion) {
	oret = nil
	return 
}