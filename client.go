package gqlclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type GraphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors GraphqlErrors   `json:"errors"`
}

func (r *GraphqlResponse) Decode(dest interface{}) error {
	return json.Unmarshal(r.Data, dest)
}

type GraphqlConn struct {
	url    string
	client *http.Client
}

func NewGraphqlConn(url string, client *http.Client) *GraphqlConn {
	return &GraphqlConn{
		url:    url,
		client: client,
	}
}

func (c *GraphqlConn) Do(gql GraphqlRequest, headers map[string]string) (*GraphqlResponse, error) {
	var gqlResp GraphqlResponse

	payload, err := json.Marshal(gql)
	if err != nil {
		return nil, errors.Wrap(err, "json-marshal graphql request payload")
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.Wrap(err, "create new request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	client := c.client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do http request")
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Wrapf(err, "decode gql, response body: %s", string(body))
	}

	return &gqlResp, gqlResp.Errors
}
