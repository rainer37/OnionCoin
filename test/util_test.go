package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/util"
	"bytes"
)

func TestSplitBytes(t *testing.T) {
	var tests = []struct {
		msg string
		n int
	} {
		{"string", 2},
		{"abc", 1},
		{"abcd", 3},
		{"dasfadsfadsfadsfadfaf", 4},
		{"fadsfadsfadsfadfafadf", 3},
		{"abcdddaerearnnfdifidifdsifi", 10},
		{"abcdddaerearnnfdifidifdsifi", 9},
		{"abcdddaerearnnfdifidifdsifi", 8},
		{"abcdddaerearnnfdifidifdsifi", 7},
	}
	for _, test := range tests {
		if expected := util.SplitBytes([]byte(test.msg), test.n);
		string(bytes.Join(expected, []byte{})) != string(test.msg) {
			t.Errorf("SplitBytes(%q, %d) = %v", test.msg, test.n, expected)
		}
	}
}
