package node

func (n *Node) MakeOnionMsg(ids []string, msg string) {

	ids = []string{"FAKEID1338", "FAKEID1339"}
	m := n.WrapABigOnion([]byte(msg), ids)

	npe := n.getPubRoutingInfo(ids[0])
	m = n.prepareOMsg(FWD, m, npe.Pk)
	n.sendActive(m, npe.Port)
}