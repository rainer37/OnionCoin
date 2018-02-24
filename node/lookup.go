package node

import (
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
)

func (n *Node) LookUpPK(address string) rsa.PublicKey {
	n.sendActive([]byte(REGISTER+string(ocrypto.EncodePK(n.sk.PublicKey))+n.Port), address)
	// waiting for the pk request return.
	enPk := <-n.pkChan
	return ocrypto.DecodePK(enPk)
}

func (n *Node) LookUpIP(id string) {

}