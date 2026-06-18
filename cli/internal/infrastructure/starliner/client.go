package starliner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"starliner.app/runner/internal/conf"
	"starliner.app/runner/internal/domain/port"
)

type Client struct {
	conf        *conf.Config
	credentials port.CredentialsStore
	http        *http.Client
}

func NewClient(conf *conf.Config, credentials port.CredentialsStore) *Client {
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	return &Client{
		conf:        conf,
		credentials: credentials,
		http:        httpClient,
	}
}

type RegisterRunnerRequest struct {
	Token string `json:"token"`
}

func (c *Client) RegisterRunner(token string) error {
	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(&RegisterRunnerRequest{Token: token}); err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.conf.ServerBaseUrl+"/runners/register",
		&body,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("register runner failed, status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) setAuth(req *http.Request) error {
	token, err := c.credentials.Token()
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return nil
}
