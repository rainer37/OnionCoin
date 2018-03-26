package ui

import (
	"bufio"
	"os"
	"fmt"
	"github.com/rainer37/OnionCoin/node"
	"regexp"
	"strings"
)

func print(str ...interface{}) {
	fmt.Print("OCSYS:$ ")
	fmt.Print(str...)
}

func Listen(n *node.Node) {
	for {
		print()
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')

		print("["+text[:len(text)-1]+"]\n")

		matched, _ := regexp.MatchString("ex *", text)
		if matched {
			tokens := strings.Split(text, " ")
			go n.CoinExchange(tokens[1][:len(tokens[1])-1])
		}
		matched, _ = regexp.MatchString("fwd *", text)
		if matched {
			tokens := strings.Split(text[:len(text)-1], " ")
			if len(tokens) >= 3 {
				n.SendOninoMsg(tokens[2:], tokens[1])
			}
		}
		matched, _ = regexp.MatchString("pub *", text)
		if matched {
			tokens := strings.Split(text, " ")
			n.CoinExchange(tokens[1][:len(tokens[1])-1])
		}

		matched, _ = regexp.MatchString("adv *", text)
		if matched {
			tokens := strings.Split(text, " ")
			n.CoinExchange(tokens[1][:len(tokens[1])-1])
		}
	}
}