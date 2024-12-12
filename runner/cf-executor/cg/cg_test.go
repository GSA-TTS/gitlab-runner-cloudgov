package cg

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type adapterStub struct {
	CloudI

	StCreds *Creds
	StURL   string
	StApps  []*App

	FailGetApps bool
	FailConnect bool
}

func (a *adapterStub) getApps() (apps []*App, err error) {
	if a.FailGetApps {
		return nil, errors.New("fail")
	}
	return a.StApps, nil
}

func (a *adapterStub) connect(url string, creds *Creds) (_ error) {
	if a.FailConnect {
		return errors.New("fail")
	}
	a.StURL = url
	a.StCreds = creds
	return nil
}

type credIStub struct {
	U    string
	P    string
	Fail bool
}

func (c credIStub) getCreds() (*Creds, error) {
	if c.Fail {
		return nil, errors.New("fail")
	}
	return &Creds{c.U, c.P}, nil
}

func TestNew(t *testing.T) {
	optsStub := &Opts{CredI: credIStub{"a", "b", false}}
	cgStub := &CG{&adapterStub{
		StURL:   apiRootURLDefault,
		StCreds: &Creds{"a", "b"},
	}, optsStub}

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
			got, err := New(&adapterStub{}, tt.opts)
			if (err == nil) != (tt.wantErr == nil) {
				t.Errorf("GetCfClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !errors.As(err, tt.wantErr) {
				t.Errorf("GetCredentials() bad error type: got %T, want %T", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestCG_apiRootURL(t *testing.T) {
	type fields struct {
		CloudI
		Opts *Opts
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
				CloudI: &adapterStub{},
				Opts:   &Opts{CredI: credIStub{}},
			},
		},
		{
			name: "updates root API URL",
			want: "foo",
			fields: fields{
				CloudI: &adapterStub{},
				Opts:   &Opts{CredI: credIStub{}, APIRootURL: "foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				CloudI: tt.fields.CloudI,
				Opts:   tt.fields.Opts,
			}
			got := c.apiRootURL()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestCG_creds(t *testing.T) {
	type fields struct {
		CloudI
		Opts *Opts
	}

	tests := []struct {
		fields  fields
		want    *Creds
		name    string
		wantErr bool
	}{
		{
			name: "returns creds when they already exist",
			want: &Creds{"a", "b"},
			fields: fields{
				CloudI: &adapterStub{},
				Opts:   &Opts{Creds: &Creds{"a", "b"}},
			},
		},
		{
			name: "returns creds from getter when not supplied",
			want: &Creds{"foo", "bar"},
			fields: fields{
				CloudI: &adapterStub{},
				Opts:   &Opts{CredI: credIStub{U: "foo", P: "bar"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				CloudI: tt.fields.CloudI,
				Opts:   tt.fields.Opts,
			}
			got, err := c.creds()
			if (err != nil) != tt.wantErr {
				t.Errorf("CG.creds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestCG_Connect(t *testing.T) {
	type fields struct {
		CloudI
		Opts *Opts
	}

	tests := []struct {
		fields  fields
		want    *CG
		cmpGet  func(c *CG) any
		name    string
		checks  []string
		wantErr bool
	}{
		{
			name:    "fails with creds() err",
			wantErr: true,
			fields: fields{
				CloudI: &adapterStub{},
				Opts:   &Opts{CredI: &credIStub{Fail: true}},
			},
		},
		{
			name: "connect sets URL & creds",
			want: &CG{
				CloudI: &adapterStub{
					StURL:   "butter",
					StCreds: &Creds{Username: "corn", Password: "cob"},
				},
			},
			fields: fields{
				CloudI: &adapterStub{},
				Opts: &Opts{
					APIRootURL: "butter",
					Creds:      &Creds{Username: "corn", Password: "cob"},
				},
			},
			cmpGet: func(c *CG) any {
				return c.CloudI
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				CloudI: tt.fields.CloudI,
				Opts:   tt.fields.Opts,
			}

			got, err := c.Connect()

			if (err != nil) != tt.wantErr {
				t.Errorf("CG.Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.cmpGet == nil {
				tt.cmpGet = func(c *CG) any {
					return c
				}
			}

			if diff := cmp.Diff(tt.cmpGet(got), tt.cmpGet(tt.want)); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestCG_GetApps(t *testing.T) {
	testApps := []*App{{Id: "1", Name: "foo"}}

	type fields struct {
		CloudI
		Opts *Opts
	}

	tests := []struct {
		name    string
		fields  fields
		want    []*App
		wantErr bool
	}{
		{
			name:    "reports errors",
			wantErr: true,
			fields: fields{
				CloudI: &adapterStub{StApps: testApps, FailGetApps: true},
				Opts:   &Opts{CredI: &credIStub{}},
			},
		},
		{
			name: "returns available apps list",
			want: testApps,
			fields: fields{
				CloudI: &adapterStub{StApps: testApps},
				Opts:   &Opts{CredI: &credIStub{}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CG{
				CloudI: tt.fields.CloudI,
				Opts:   tt.fields.Opts,
			}
			got, err := c.GetApps()
			if (err != nil) != tt.wantErr {
				t.Errorf("CG.GetApps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
