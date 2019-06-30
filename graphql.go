package gqlclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	auth0grant "github.com/ereyes01/go-auth0-grant"
	"github.com/pkg/errors"
)

type GraphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphqlError  `json:"errors"`
}

func DoGraphql(url string, gql *GraphqlRequest, grant *auth0grant.Grant) (*GraphqlResponse, error) {
	var (
		gqlResp     GraphqlResponse
		accessToken string
	)

	if grant != nil {
		var err error

		accessToken, err = grant.GetAccessToken()
		if err != nil {
			return nil, errors.Wrap(err, "get API access token")
		}
	}

	payload, err := json.Marshal(gql)
	if err != nil {
		return nil, errors.Wrap(err, "json-marshal graphql request payload")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.Wrap(err, "create new request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("response status: %s, body: %s", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Wrapf(err, "decode gql, response body: %s", string(body))
	}

	if len(gqlResp.Errors) > 0 {
		return nil, GraphqlErrors(gqlResp.Errors)
	}

	return &gqlResp, nil
}
