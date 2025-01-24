package cloudgov

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type stubClientAPI struct {
	ClientAPI

	StURL   string
	StCreds *Creds
	StApps  []*App

	FailAppsList bool
	FailConnect  bool
}

func (a *stubClientAPI) appsList() (apps []*App, err error) {
	if a.FailAppsList {
		return nil, errors.New("fail")
	}
	return a.StApps, nil
}

func (a *stubClientAPI) connect(url string, creds *Creds) (_ error) {
	if a.FailConnect {
		return errors.New("fail")
	}
	a.StURL = url
	a.StCreds = creds
	return nil
}

type stubCredsGetter struct {
	U    string
	P    string
	Fail bool
}

func (c stubCredsGetter) getCreds() (*Creds, error) {
	if c.Fail {
		return nil, errors.New("fail")
	}
	return &Creds{c.U, c.P}, nil
}

func TestNew(t *testing.T) {
	optsStub := &Opts{CredsGetter: stubCredsGetter{"a", "b", false}}
	cgStub := &Client{&stubClientAPI{
		StURL:   apiRootURLDefault,
		StCreds: &Creds{"a", "b"},
	}, optsStub}

	tests := []struct {
		want    *Client
		opts    *Opts
		wantErr interface{}
		name    string
	}{
		{name: "fails using default credential getter", wantErr: &syntaxError},
		{name: "returns adapted Client struct", want: cgStub, opts: optsStub},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(&stubClientAPI{}, tt.opts)
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

func TestClient_apiRootURL(t *testing.T) {
	type fields struct {
		ClientAPI
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
				ClientAPI: &stubClientAPI{},
				Opts:      &Opts{CredsGetter: stubCredsGetter{}},
			},
		},
		{
			name: "updates root API URL",
			want: "foo",
			fields: fields{
				ClientAPI: &stubClientAPI{},
				Opts:      &Opts{CredsGetter: stubCredsGetter{}, APIRootURL: "foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				ClientAPI: tt.fields.ClientAPI,
				Opts:      tt.fields.Opts,
			}
			got := c.apiRootURL()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestClient_creds(t *testing.T) {
	type fields struct {
		ClientAPI
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
				ClientAPI: &stubClientAPI{},
				Opts:      &Opts{Creds: &Creds{"a", "b"}},
			},
		},
		{
			name: "returns creds from getter when not supplied",
			want: &Creds{"foo", "bar"},
			fields: fields{
				ClientAPI: &stubClientAPI{},
				Opts:      &Opts{CredsGetter: stubCredsGetter{U: "foo", P: "bar"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				ClientAPI: tt.fields.ClientAPI,
				Opts:      tt.fields.Opts,
			}
			got, err := c.creds()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.creds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestClient_Connect(t *testing.T) {
	type fields struct {
		ClientAPI
		Opts *Opts
	}

	tests := []struct {
		fields  fields
		want    *Client
		cmpGet  func(c *Client) any
		name    string
		checks  []string
		wantErr bool
	}{
		{
			name:    "fails with creds() err",
			wantErr: true,
			fields: fields{
				ClientAPI: &stubClientAPI{},
				Opts:      &Opts{CredsGetter: &stubCredsGetter{Fail: true}},
			},
		},
		{
			name: "connect sets URL & creds",
			want: &Client{
				ClientAPI: &stubClientAPI{
					StURL:   "butter",
					StCreds: &Creds{Username: "corn", Password: "cob"},
				},
			},
			fields: fields{
				ClientAPI: &stubClientAPI{},
				Opts: &Opts{
					APIRootURL: "butter",
					Creds:      &Creds{Username: "corn", Password: "cob"},
				},
			},
			cmpGet: func(c *Client) any {
				return c.ClientAPI
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				ClientAPI: tt.fields.ClientAPI,
				Opts:      tt.fields.Opts,
			}

			got, err := c.Connect()

			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.cmpGet == nil {
				tt.cmpGet = func(c *Client) any {
					return c
				}
			}

			if diff := cmp.Diff(tt.cmpGet(got), tt.cmpGet(tt.want)); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}

func TestClient_AppsList(t *testing.T) {
	testApps := []*App{{Id: "1", Name: "foo"}}

	type fields struct {
		ClientAPI
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
				ClientAPI: &stubClientAPI{StApps: testApps, FailAppsList: true},
				Opts:      &Opts{CredsGetter: &stubCredsGetter{}},
			},
		},
		{
			name: "returns available apps list",
			want: testApps,
			fields: fields{
				ClientAPI: &stubClientAPI{StApps: testApps},
				Opts:      &Opts{CredsGetter: &stubCredsGetter{}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				ClientAPI: tt.fields.ClientAPI,
				Opts:      tt.fields.Opts,
			}
			got, err := c.AppsList()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.AppsList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
