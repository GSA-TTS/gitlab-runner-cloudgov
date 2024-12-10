package cg

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type adapterStub struct {
	Adapter
}

func (adapterStub) getApps() (apps []*App, err error) {
	panic("getApps not implemented") // TODO: Implement
}

func (adapterStub) connect(url string, creds *Creds) (_ error) {
	return nil
}

type credIStub struct{}

func (credIStub) getCreds() (*Creds, error) {
	return &Creds{"", ""}, nil
}

var (
	optsStub *Opts = &Opts{CredI: credIStub{}}
	cgStub   *CG   = &CG{adapterStub{}, optsStub}
)

func TestNew(t *testing.T) {
	tests := []struct {
		want    *CG
		opts    *Opts
		wantErr interface{}
		name    string
	}{
		{name: "fails using default credential getter", wantErr: &syntaxError},
		{name: "returns adapted CG struct", want: cgStub, opts: optsStub},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := New(adapterStub{}, tt.opts)

			if (err == nil) != (tt.wantErr == nil) {
				t.Errorf("GetCfClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && !errors.As(err, tt.wantErr) {
				t.Errorf("GetCredentials() bad error type: have %T, want %T", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, have); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCG_apiRootURL(t *testing.T) {
	type fields struct {
		Adapter Adapter
		Opts    *Opts
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "gets default root API URL",
			want: apiRootURLDefault,
			fields: fields{
				Adapter: adapterStub{},
				Opts:    &Opts{CredI: credIStub{}},
			},
		},
		{
			name: "updates root API URL",
			want: "foo",
			fields: fields{
				Adapter: adapterStub{},
				Opts:    &Opts{CredI: credIStub{}, APIRootURL: "foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				Adapter: tt.fields.Adapter,
				Opts:    tt.fields.Opts,
			}
			got := c.apiRootURL()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCG_creds(t *testing.T) {
	type fields struct {
		Adapter Adapter
		Opts    *Opts
	}
	tests := []struct {
		fields  fields
		want    *Creds
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				Adapter: tt.fields.Adapter,
				Opts:    tt.fields.Opts,
			}
			got, err := c.creds()
			if (err != nil) != tt.wantErr {
				t.Errorf("CG.creds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCG_Connect(t *testing.T) {
	type fields struct {
		Adapter Adapter
		Opts    *Opts
	}
	tests := []struct {
		fields  fields
		want    *CG
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				Adapter: tt.fields.Adapter,
				Opts:    tt.fields.Opts,
			}
			got, err := c.Connect()
			if (err != nil) != tt.wantErr {
				t.Errorf("CG.Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCG_GetApps(t *testing.T) {
	type fields struct {
		Adapter Adapter
		Opts    *Opts
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*App
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				Adapter: tt.fields.Adapter,
				Opts:    tt.fields.Opts,
			}
			got, err := c.GetApps()
			if (err != nil) != tt.wantErr {
				t.Errorf("CG.GetApps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Error(diff)
			}
		})
	}
}
