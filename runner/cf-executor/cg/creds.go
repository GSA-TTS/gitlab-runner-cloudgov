package cg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Creds struct {
	Username string
	Password string
}

func (cr *Creds) isEmpty() bool {
	return cr.Username == "" && cr.Password == ""
}

type vcapData struct {
	CGSrvAct []struct{ Creds Creds } `json:"cloud-gov-service-account"`
}

func getCreds() (*Creds, error) {
	creds := &Creds{}

	// Check if credentials are supplied through environment
	creds.Username = os.Getenv("CF_USERNAME")
	creds.Password = os.Getenv("CF_PASSWORD")
	if !creds.isEmpty() {
		return creds, nil
	}

	// Check for credentials in VCAP_SERVICES JSON
	var vcd vcapData
	vSrv := os.Getenv("VCAP_SERVICES")
	if err := json.Unmarshal([]byte(vSrv), &vcd); err != nil {
		return nil, fmt.Errorf("error unmarshaling VCAP_SERVICES: %w", err)
	}

	// If creds are still empty we fail
	creds = &vcd.CGSrvAct[0].Creds
	if creds.isEmpty() {
		return nil, errors.New("could not establish credentials")
	}

	return creds, nil
}
