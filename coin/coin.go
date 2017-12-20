package coin

const COIN_PREFIX string = "[COIN]"

type Coin struct {
	ID string
	PK []byte
	Sig []byte
}

func print(str string) {
	println(COIN_PREFIX, str)
}

func New_Coin() Coin {
	coin := Coin{"1338", []byte("hello"), []byte("world")}
	print("Coin is an Onion : " + string(coin.ID))
	return coin
}

func (c Coin) Get_PK() []byte {
	return c.PK
}

func (c Coin) Get_Sig() []byte {
	return c.Sig
}

func (c Coin) Get_ID() string {
	return c.ID
}