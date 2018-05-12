package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/coin"
	"os"
	"github.com/rainer37/OnionCoin/util"
	"io/ioutil"
)

func TestDepositCoins(t *testing.T) {
	vault := coin.InitVault()
	defer os.RemoveAll(coin.COINDIR)

	if vault.GetBalance() != 1 { t.Error("wrong balace") }

	c1 := coin.NewCoin("c1", []byte("contents"), []string{"r1", "r2"})
	c2 := coin.NewCoin("c2", []byte("contents"), []string{"r1", "r2"})
	c3 := coin.NewCoin("c3", []byte("contents"), []string{"r1", "r2"})
	c4 := coin.NewCoin("c1", []byte("contents"), []string{"r1", "r2"})

	vault.Deposit(c1)
	vault.Deposit(c2)
	vault.Deposit(c3)

	if vault.GetBalance() != 4 { t.Error("wrong balace") }
	if !vault.Contains(c1.GetRID()) || !vault.Contains(c2.GetRID()) || !vault.Contains(c3.GetRID()) {
		t.Error("wrong coin contained")
	}
	if vault.GetNumCoins(c1.GetRID()) != 1 { t.Error("wrong get num of coins") }

	vault = coin.InitVault()

	if vault.GetBalance() != 4 { t.Error("wrong balace") }
	if !vault.Contains(c1.GetRID()) || !vault.Contains(c2.GetRID()) || !vault.Contains(c3.GetRID()) {
		t.Error("wrong coin contained")
	}
	if vault.GetNumCoins(c1.GetRID()) != 1 { t.Error("wrong get num of coins") }

	vault.Deposit(c4)
	if vault.GetNumCoins(c1.GetRID()) != 2 { t.Error("wrong get num of coins") }
}

func TestWithdrawCoins(t *testing.T) {
	vault := coin.InitVault()
	defer os.RemoveAll(coin.COINDIR)

	c1 := coin.NewCoin("c1", []byte("contents"), []string{"r1", "r2"})
	c2 := coin.NewCoin("c2", []byte("contents"), []string{"r1", "r2"})
	c3 := coin.NewCoin("c3", []byte("contents"), []string{"r1", "r2"})
	c4 := coin.NewCoin("c1", []byte("contents"), []string{"r1", "r2"})

	vault.Deposit(c1)
	vault.Deposit(c2)
	vault.Deposit(c3)
	vault.Deposit(c4)

	if GetNumCoinFiles() != 4 { t.Error("wrong num of coins file")}

	nc1 := vault.Withdraw(c2.GetRID())
	if nc1.GetRID() != "c2" { t.Error("WRONG ID") }
	if util.GetID(nc1.GetContent()) != "contents" { t.Error("WRONG CONTENTS") }
	if len(nc1.Signers) != 2 { t.Error("WRONG NUM SIGNERS") }
	if nc1.Signers[0] != "r1" || nc1.Signers[1] != "r2" { t.Error("WRONG SIGNES" )}

	if GetNumCoinFiles() != 3 { t.Error("wrong num of coins file") }
	if vault.Contains(c2.GetRID()) { t.Error("wrong contained coins") }

	nc2 := vault.Withdraw(c1.GetRID())
	if nc2.GetRID() != "c1" { t.Error("WRONG ID") }
	if util.GetID(nc1.GetContent()) != "contents" { t.Error("WRONG CONTENTS") }
	if len(nc2.Signers) != 2 { t.Error("WRONG NUM SIGNERS") }
	if nc2.Signers[0] != "r1" || nc2.Signers[1] != "r2" { t.Error("WRONG SIGNES" )}

	if GetNumCoinFiles() != 2 { t.Error("wrong num of coins file") }
	if vault.GetNumCoins(c1.GetRID()) != 1 { t.Error("wrong contained coins") }

	vault.Withdraw(c3.GetRID())
	vault.Withdraw(c1.GetRID())
	if GetNumCoinFiles() != 0 { t.Error("wrong num of coins file") }

	nn := vault.Withdraw(c1.GetRID())
	if nn != nil { t.Error("error getting coin when empty") }
}

func GetNumCoinFiles() int {
	files, err := ioutil.ReadDir(coin.COINDIR)
	util.CheckErr(err)
	return len(files)
}