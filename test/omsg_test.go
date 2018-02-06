package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/records"
	"fmt"
	"strings"
)

func TestOMsgMashal(t *testing.T) {
	//msg := "the-key-has-to-be-32-bytes-long!"

	sk := ocrypto.RSAKeyGen()

	payload := []byte("rainer is god")
	nodeID := "Ella"
	m := records.MarshalOMsg('0', payload, nodeID, sk)
	omsg, ok := records.UnmarshalOMsg(m, sk)
	fmt.Println(omsg.B)
	if !ok {
		t.Error("cannot UmarshallOMsg")
	}

	if omsg.GetOPCode() != '0' {
		t.Error("wrong opcode")
	}

	if !strings.Contains(string(omsg.GetSenderID()), nodeID) {
		t.Error("wrong sender ID")
	}

	if omsg.GetLenPayload() != 13 {
		t.Error("wrong payload len")
	}

	if string(omsg.GetPayload(omsg.GetLenPayload())) != string(payload) {
		t.Error("wrong payload")
	}

	if string(omsg.GetPayloadTsHash()) != string(ocrypto.RSASign(sk, payload)) {
		t.Error("wrong sig")
	}
}
