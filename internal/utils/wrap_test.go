package utils

import (
	"fmt"
	"testing"
)

func TestWrap(t *testing.T) {
	text := "Ut dolor adipisicing laboris labore ut nostrud velit. Ipsum voluptate est sit nostrud laboris et voluptate laborum eu."

	leftWrap := Wrap(text, 20)
	rightWrap := RightWrap(text, 20)

	fmt.Println(leftWrap)
	fmt.Println(rightWrap)

	t.Fail()
}
