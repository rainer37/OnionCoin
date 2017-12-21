package bank

type FreeCoinList struct {
	FreeCoins []uint64
}

func (fcl *FreeCoinList) insert(cid uint64) {
	fcl.FreeCoins = append(fcl.FreeCoins, cid)
	println("Inserting Coin :",cid)
}