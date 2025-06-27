package cloudgov

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type stubClientAPI struct {
	ClientAPI

	StURL   string
	StCreds *Creds
	StApps  []*App

	FailConnect   bool
	FailAppsList  bool
	FailAppPush   bool
	FailAppFound  bool
	FailAppDelete bool
}

type testErr struct {
	message string
}

func (e *testErr) Error() string {
	return e.message
}

func (e *testErr) Is(err error) bool {
	return e.Error() == err.Error()
}

func (a *stubClientAPI) connect(url string, creds *Creds) (_ error) {
	if a.FailConnect {
		return &testErr{"fail"}
	}
	a.StURL = url
	a.StCreds = creds
	return nil
}

func (a *stubClientAPI) appGet(id string) (*App, error) {
	if id == "" || a.FailAppFound {
		return nil, nil
	}
	return &App{Name: id}, nil
}

func (a *stubClientAPI) appDelete(id string) error {
	if id == "" {
		return &testErr{"Need an App ID to delete"}
	}
	if a.FailAppDelete {
		return &testErr{"FailAppDelete"}
	}
	return nil
}

func (a *stubClientAPI) appPush(m *AppManifest) (*App, error) {
	if a.FailAppPush {
		return nil, &testErr{"FailAppPush"}
	}
	if m.Name == "" {
		return nil, &testErr{"appPush: malformed manifest"}
	}
	return &App{Name: m.Name, State: "TEST"}, nil
}

func (a *stubClientAPI) appsList() (apps []*App, err error) {
	if a.FailAppsList {
		return nil, &testErr{"FailAppsList"}
	}
	return a.StApps, nil
}

type stubCredsGetter struct {
	U    string
	P    string
	Fail bool
}

func (c stubCredsGetter) getCreds() (*Creds, error) {
	if c.Fail {
		return nil, &testErr{"fail"}
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
	testApps := []*App{{Name: "foo"}}

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

func TestClient_Push(t *testing.T) {
	optsStub := &Opts{CredsGetter: stubCredsGetter{"a", "b", false}}
	cgStub := &Client{&stubClientAPI{
		StURL:   apiRootURLDefault,
		StCreds: &Creds{"a", "b"},
	}, optsStub}

	type fields struct {
		ClientAPI ClientAPI
		Opts      *Opts
	}
	type args struct {
		manifest *AppManifest
	}

	tests := map[string]struct {
		fields  fields
		args    args
		want    *App
		wantErr error
	}{
		"Fails without name": {
			fields:  fields{ClientAPI: cgStub, Opts: optsStub},
			args:    args{manifest: &AppManifest{}},
			wantErr: CloudGovClientError{"Push: AppManifest.Name must be defined"},
		},
		"Fails without org": {
			fields:  fields{ClientAPI: cgStub, Opts: optsStub},
			args:    args{manifest: &AppManifest{Name: "Some App"}},
			wantErr: CloudGovClientError{"Push: AppManifest must have Org and Space names"},
		},
		"Passes with all fields": {
			fields: fields{ClientAPI: cgStub, Opts: optsStub},
			args:   args{manifest: &AppManifest{Name: "Some App", OrgName: "Some", SpaceName: "Space"}},
			want:   &App{Name: "Some App", State: "TEST"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Client{
				ClientAPI: tt.fields.ClientAPI,
				Opts:      tt.fields.Opts,
			}
			got, err := c.Push(tt.args.manifest)
			if err != nil || tt.wantErr != nil {
				if tt.wantErr == nil {
					t.Errorf("Client.AppsList() error = %v", err)
				} else if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
					t.Errorf("Client.Push() error mismatch (-want +got):\n%s", diff)
				}
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
