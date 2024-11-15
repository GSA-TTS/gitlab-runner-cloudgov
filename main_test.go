package main

import (
	"fmt"
	"testing"
)

func Test_GetCredentials(t *testing.T) {
	username := "tjefferson"
	password := "p4ssw0rdz1"

	json := fmt.Sprintf(
		`{"cloud-gov-service-account":[{"credentials":{"password":"%s","username":"%s"}}]}`,
		password,
		username,
	)

	t.Setenv("VCAP_SERVICES", json)

	t.Run("Get credentials from VCAP_SERVICES", func(t *testing.T) {
		credentials, err := GetCredentials()
		if err != nil {
			t.Error(err)
		}
		if credentials.Username != username {
			t.Errorf("credentials.Username = %s; want %s", credentials.Username, username)
		}
		if credentials.Password != password {
			t.Errorf("credentials.Password = %s; want %s", credentials.Password, password)
		}
	})
}

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
