package node

import (
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/util"
	"time"
	"github.com/rainer37/OnionCoin/blockChain"
)

/*
	CoSigning protocol to get the newBie registered into blockChain.
	pk, id = newBie'pub key and its id.
 */
func (n *Node) registerCoSign(pk rsa.PublicKey, id string) {

	banks := currentBanks
	counter := 1
	newBieInfo := append(ocrypto.EncodePK(pk), []byte(id)...)

	// first sign it by myself.
	pkHash := util.Sha(ocrypto.EncodePK(pk))
	regBytes := n.blindSign(append(pkHash[:], []byte(id)...))

	signers := []string{n.ID}

	for _, b := range banks {
		if counter == util.NUMCOSIGNER {
			break
		}
		if b != n.ID {
			// print("sending REGCOSIGNRQ to", b)
			n.sendOMsgWithID(REGCOSIGNREQUEST, newBieInfo, b)

			var rBytes []byte

			select{
			case reply := <-n.regChan:
				print("cosigned pk received from", b)
				rBytes = reply
			case <-time.After(util.COSIGNTIMEOUT * time.Second):
				print(b, "reg cosign no response, try next bank")
				// counter++
				continue
			}

			regBytes = append(regBytes, rBytes...)
			signers = append(signers, b)
			counter++
		}
	}

	if len(signers) != util.NUMCOSIGNER {
		print("Not Enough Sigs Acquired, abort...")
		return
	}

	print("Enough Signing Received, Register Node", id,
		"by", len(signers), "Signer:", signers)

	txn := blockChain.NewPKRTxn(id, pk, regBytes, signers)
	n.bankProxy.AddTxn(txn)

	// TODO: sync this?
	// go n.broadcastTxn(txn, blockChain.PK)
}

/*
	Upon received register request, sign the pk, and reply it.
 */
func (n *Node) regCoSignRequest(payload []byte, senderID string) {
	pkHash := util.Sha(payload[:PKRQLEN])
	mySig := n.blindSign(append(pkHash[:], payload[PKRQLEN:]...))
	n.sendOMsgWithID(REGCOSIGNREPLY, mySig, senderID)
}