package starliner

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func dialRunnerGRPC(target, baseURL string, insecureSkipTLSVerify bool) (*grpc.ClientConn, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	var transportCreds credentials.TransportCredentials
	if u.Scheme == "http" {
		transportCreds = insecure.NewCredentials()
	} else {
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
		if insecureSkipTLSVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		transportCreds = credentials.NewTLS(tlsConfig)
	}

	return grpc.NewClient(target, grpc.WithTransportCredentials(transportCreds))
}

func grpcTarget(baseURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("base URL must include a host")
	}

	if u.Port() != "" {
		return u.Host, nil
	}

	switch u.Scheme {
	case "https":
		return net.JoinHostPort(u.Hostname(), "443"), nil
	case "http":
		return net.JoinHostPort(u.Hostname(), "80"), nil
	default:
		return "", fmt.Errorf("base URL must use http or https")
	}
}
