package node

import "github.com/rainer37/OnionCoin/records"

func (n *Node) CoinExchange(dstID string) {
	fentry := records.GetKeyByID(dstID)

	if fentry == nil {
		print("ERR finding id provided", dstID)
		return
	}

	fo := records.MarshalOMsg(COINEXCHANGE,[]byte("Spare a Coin?"),n.ID,n.sk,fentry.Pk)
	n.sendActive(fo, fentry.Port)
}
