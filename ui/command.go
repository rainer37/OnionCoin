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
	fmt.Print("OCSYS:")
	fmt.Print(str...)
}

func NewCommand() {
	println("creating new command.")
}

func Listen(n *node.Node) {
	for {
		print()
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		print(text)

		matched, _ := regexp.MatchString("ex *", text)
		if matched {
			tokens := strings.Split(text, " ")
			n.CoinExchange(tokens[1][:len(tokens[1])-1])
		}
		matched, _ = regexp.MatchString("fwd *", text)
		if matched {
			tokens := strings.Split(text, " ")
			n.CoinExchange(tokens[1][:len(tokens[1])-1])
		}

	}
}