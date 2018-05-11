package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/records"
	"strings"
)

func TestOMsgMashal(t *testing.T) {

	sk := ocrypto.RSAKeyGen()
	payload := []byte("gods upon sky")
	nodeID := "Ella"
	m := records.MarshalOMsg('0', payload, nodeID, sk, sk.PublicKey)

	if m == nil {
		t.Error("cannot MarshallOMsg")
	}

	omsg, ok := records.UnmarshalOMsg(m, sk)

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

	if string(omsg.GetPayload()) != string(payload) {
		t.Error("wrong payload")
	}

	if string(omsg.GetPayloadTsHash()) != string(ocrypto.RSASign(sk, payload)) {
		t.Error("wrong sig")
	}
}
