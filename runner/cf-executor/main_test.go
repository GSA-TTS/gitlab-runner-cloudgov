package main

import (
	"fmt"
	"testing"
)

// func TestGetCfClient(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		want    *client.Client
// 		wantErr bool
// 	}{{}}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := GetCfClient()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetCfClient() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetCfClient() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

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
