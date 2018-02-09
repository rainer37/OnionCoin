package node

import (
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/records"
)

func (n *Node) forwardProtocol(payload []byte) {
	sid, prevCoin, iOnion := ocrypto.PeelOnion(n.sk, payload)
	print(sid, string(prevCoin), string(iOnion))

	pe := records.GetKeyByID(sid)

	if pe == nil {
		print("destination reached")
		return
	}

	m := records.MarshalOMsg(FWD,iOnion,n.ID,n.sk,pe.Pk)
	n.sendActive(string(m), pe.Port)

}