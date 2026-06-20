package starliner

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"starliner.app/runner/internal/domain/port"
)

type Client struct {
	config       port.ConfigStore
	registration port.RegistrationStore
	workload     port.WorkloadTracker
	http         *http.Client
}

func NewClient(
	config port.ConfigStore,
	registration port.RegistrationStore,
	workload port.WorkloadTracker,
) *Client {
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	return &Client{
		config:       config,
		registration: registration,
		workload:     workload,
		http:         httpClient,
	}
}

type RegisterRunnerRequest struct {
	Token             string   `json:"token"`
	Name              string   `json:"name"`
	Labels            []string `json:"labels"`
	MaxConcurrentJobs int      `json:"maxConcurrentJobs"`
}

func (c *Client) RegisterRunner(token, name string, labels []string, maxConcurrentJobs int, insecureSkipTLSVerify bool) error {
	baseURL, err := c.config.BaseURL()
	if err != nil {
		return err
	}

	var body bytes.Buffer

	if err := json.NewEncoder(&body).Encode(&RegisterRunnerRequest{
		Token:             token,
		Name:              name,
		Labels:            labels,
		MaxConcurrentJobs: maxConcurrentJobs,
	}); err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		strings.TrimRight(baseURL, "/")+"/api/runners/register",
		&body,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	httpClient := c.http
	if insecureSkipTLSVerify {
		httpClient = &http.Client{
			Timeout: c.http.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	resp, err := httpClient.Do(req)
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

func (c *Client) maxConcurrentJobs() (int32, error) {
	maxJobs, err := c.registration.MaxConcurrentJobs()
	if err != nil {
		return 0, err
	}

	return int32(maxJobs), nil
}

func (c *Client) activeJobs() int32 {
	if c.workload == nil {
		return 0
	}

	return int32(c.workload.ActiveJobs())
}
