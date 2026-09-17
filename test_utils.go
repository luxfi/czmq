package czmq

import (
	"fmt"
	"strings"
	"testing"
)

func assertEqual(t *testing.T, expected any, value any, message ...string) {
	if expected == value {
		return
	}
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("'%v' != '%v'. Expected '%v' but got '%v'", expected, value, expected, value))
	for _, m := range message {
		msg.WriteString(fmt.Sprintf(". %s", m))
	}
	t.Fatal(msg.String())
}
