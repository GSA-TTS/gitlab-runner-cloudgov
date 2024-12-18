//go:build !integration

package main

import (
	"fmt"
	"testing"
)

func Test_main(t *testing.T) {
	t.Skip("Test_main is just for experiments right now")

	tests := []struct{ name string }{{name: "run main"}}

	for _, tt := range tests {
		fmt.Println(tt.name)
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
