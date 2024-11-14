package main

import (
	"fmt"
	"testing"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "scoobie"},
	}
	for _, tt := range tests {
		fmt.Println(tt.name)
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
