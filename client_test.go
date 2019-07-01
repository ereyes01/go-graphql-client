package gqlclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testConfig struct {
	server  *httptest.Server
	handler http.HandlerFunc
	nCalls  int
}

func (c *testConfig) handlerWrapper(w http.ResponseWriter, r *http.Request) {
	c.nCalls++
	c.handler(w, r)
}

type respData struct {
	X float64 `json:"_x_"`
	Y string  `json:"_y_"`
}

type testCase struct {
	name     string
	request  GraphqlRequest
	response *GraphqlResponse
	expData  respData
}

func (tc testCase) getHandler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()

		// verify headers
		if r.Method != "POST" {
			t.Fatalf("wrong method got: %s expected: POST", r.Method)
		}
		mime := r.Header.Get("Content-Type")
		if mime != "application/json" {
			t.Fatalf("wrong mime type got: %s expected: application/json", mime)
		}
		accept := r.Header.Get("Accept")
		if accept != "application/json" {
			t.Fatalf("wrong accept type got: %s expected: application/json", accept)
		}

		// verify body is as expected
		var recvd GraphqlRequest
		if err := json.NewDecoder(r.Body).Decode(&recvd); err != nil {
			t.Fatal("json unmarshal requets body:", err)
		}
		if !cmp.Equal(tc.request, recvd) {
			t.Fatal("unexpected graphql request:", cmp.Diff(tc.request, recvd))
		}

		// marshal and send response
		respData, err := json.Marshal(tc.response)
		if err != nil {
			t.Fatal("json marshal response:", err)
		}
		w.Write(respData)
	}
}

func TestClient(t *testing.T) {
	setup := func(t *testing.T, h http.HandlerFunc) (*testConfig, func(t *testing.T)) {
		t.Helper()

		c := &testConfig{handler: h}
		c.server = httptest.NewServer(http.HandlerFunc(c.handlerWrapper))

		tearDown := func(t *testing.T) {
			t.Helper()
			if c.nCalls != 1 {
				t.Fatalf("wrong number of requests, got: %d expected: 1", c.nCalls)
			}
		}

		return c, tearDown
	}

	cases := []testCase{
		{
			name: "successful request, no errors",
			request: GraphqlRequest{
				Query:     "query1",
				Variables: map[string]interface{}{"x": 42.0},
			},
			response: &GraphqlResponse{
				Data:   []byte(`{"_x_":1.0,"_y_":"tc1"}`),
				Errors: nil,
			},
			expData: respData{X: 1.0, Y: "tc1"},
		},
		{
			name: "response with errors",
			request: GraphqlRequest{
				Query:     "query2",
				Variables: map[string]interface{}{"x": 43.0},
			},
			response: &GraphqlResponse{
				Data: []byte(`{"_x_":1.0,"_y_":"tc1"}`),
				Errors: GraphqlErrors{
					{
						Status:  412,
						Message: "error-2",
					},
				},
			},
			expData: respData{X: 1.0, Y: "tc1"},
		},
	}

	for _, tc := range cases {
		fixture, tearDown := setup(t, tc.getHandler(t))
		defer tearDown(t)

		conn := NewGraphqlConn(fixture.server.URL, nil)
		resp, err := conn.Do(tc.request, nil)
		if !cmp.Equal(tc.response.Errors, err) {
			t.Fatal("unexpected graphql error:", cmp.Diff(tc.response.Errors, err))
		}
		if !cmp.Equal(tc.response, resp) {
			t.Fatalf("unexpected graphql response data expected: %s got: %s", string(tc.response.Data), string(resp.Data))
		}

		var p respData
		if err := resp.Decode(&p); err != nil {
			t.Fatal("decode graphql response:", err)
		}
		if !cmp.Equal(tc.expData, p) {
			t.Fatal("unexpected graphql response payload:", cmp.Diff(tc.expData, p))
		}
	}
}
