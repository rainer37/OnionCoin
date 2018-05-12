package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"os"
	"crypto/rsa"
)

func TestPEEncodeDecode(t *testing.T) {
	pk := ocrypto.RSAKeyGen().PublicKey
	pe := records.PKEntry{pk, "192.168.0.33", "3330", 12345}

	pbytes := pe.Bytes()
	npe := records.BytesToPKEntry(pbytes)

	comparePE(t, npe, pk, "192.168.0.33", "3330", 12345)
}

func TestInsertAndGetPE(t *testing.T) {
	records.GenerateKeyRepo()
	defer os.RemoveAll(records.KEYDIR)

	pk := ocrypto.RSAKeyGen().PublicKey
	pk1 := ocrypto.RSAKeyGen().PublicKey

	records.InsertEntry("f", pk, 12345,"192.168.0.33", "3330")
	records.InsertEntry("s", pk1, 999999,"127.0.0.1", "22")

	if records.KeyRepoSize() != 2 { t.Error("wrong after size") }

	pe1 := records.GetKeyByID("f")
	pe2 := records.GetKeyByID("s")

	comparePE(t, pe1, pk,"192.168.0.33", "3330", 12345)
	comparePE(t, pe2, pk1, "127.0.0.1", "22", 999999)

	records.InitKeyRepo()

	if records.KeyRepoSize() != 0 { t.Error("wrong ini size") }

	pe1 = records.GetKeyByID("f")
	comparePE(t, pe1, pk,"192.168.0.33", "3330", 12345)

}

func TestPopulatePKEntry(t *testing.T) {
	os.Mkdir(records.KEYDIR, 0777)
	defer os.RemoveAll(records.KEYDIR)
	records.InitKeyRepo()

	pk := ocrypto.RSAKeyGen().PublicKey
	pe := records.PKEntry{pk, "192.168.0.33", "3330", 12345}
	pk1 := ocrypto.RSAKeyGen().PublicKey
	pe1 := records.PKEntry{pk1, "127.0.0.1", "22", 999999}

	records.WritePE(&pe, "first")
	records.WritePE(&pe1, "second")

	if records.KeyRepoSize() != 0 { t.Error("wrong ini size") }

	records.PopulatePKEntry()

	if records.KeyRepoSize() != 2 { t.Error("wrong after size") }

	comparePE(t, records.GetKeyByID("first"), pk,"192.168.0.33", "3330", 12345)
	comparePE(t, records.GetKeyByID("second"), pk1, "127.0.0.1", "22", 999999)
}

func TestPackPEs(t *testing.T) {
	records.GenerateKeyRepo()
	defer os.RemoveAll(records.KEYDIR)

	pk := ocrypto.RSAKeyGen().PublicKey
	pk1 := ocrypto.RSAKeyGen().PublicKey

	records.InsertEntry("f", pk, 12345,"192.168.0.33", "3330")
	records.InsertEntry("s", pk1, 999999,"127.0.0.1", "22")

	two := records.PackPEs(2)

	records.InitKeyRepo()

	if records.KeyRepoSize() != 0 { t.Error("wrong ini size") }

	records.UnpackPEs(two)

	if records.KeyRepoSize() != 2 { t.Error("wrong after size") }

	pe1 := records.GetKeyByID("f")
	pe2 := records.GetKeyByID("s")

	comparePE(t, pe1, pk,"192.168.0.33", "3330", 12345)
	comparePE(t, pe2, pk1, "127.0.0.1", "22", 999999)
}

func comparePE(t *testing.T, npe *records.PKEntry, pk rsa.PublicKey, ip string, port string, time int64) {
	if npe.IP != ip {
		t.Error("IP not equal")
	}

	if npe.Port != port {
		t.Error("Port not equal")
	}

	if npe.Time != time {
		t.Error("Time not equal")
	}

	if npe.Pk.N.Cmp(pk.N) != 0 || npe.Pk.E != pk.E {
		t.Error("PK not equal")
	}
}