package starliner

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	v1 "starliner.app/runner/internal/infrastructure/grpc/proto/v1"
)

const (
	runnerTokenMetadataKey = "authorization"

	defaultHeartbeatInterval = 5 * time.Second
	minReconnectDelay        = time.Second
	maxReconnectDelay        = 30 * time.Second
)

func (c *Client) ServeHeartbeats(ctx context.Context, insecureSkipTLSVerify bool) error {
	token, err := c.credentials.Token()
	if err != nil {
		return err
	}
	if token == "" {
		return fmt.Errorf("runner not registered; run `runner register` first")
	}

	baseURL, err := c.config.BaseURL()
	if err != nil {
		return err
	}

	target, err := grpcTarget(baseURL)
	if err != nil {
		return err
	}

	conn, err := dialScheduler(target, baseURL, insecureSkipTLSVerify)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	client := v1.NewRunnerSchedulerServiceClient(conn)

	var sequence uint64
	interval := defaultHeartbeatInterval
	reconnectDelay := minReconnectDelay

	for {
		if ctx.Err() != nil {
			return nil
		}

		nextInterval, err := runHeartbeatSession(ctx, client, token, &sequence, interval)
		if err == nil {
			reconnectDelay = minReconnectDelay
			if nextInterval > 0 {
				interval = nextInterval
			}

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(interval):
			}
			continue
		}

		if ctx.Err() != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(reconnectDelay):
		}

		if reconnectDelay < maxReconnectDelay {
			reconnectDelay *= 2
			if reconnectDelay > maxReconnectDelay {
				reconnectDelay = maxReconnectDelay
			}
		}
	}
}

func runHeartbeatSession(
	ctx context.Context,
	client v1.RunnerSchedulerServiceClient,
	token string,
	sequence *uint64,
	interval time.Duration,
) (time.Duration, error) {
	streamCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(runnerTokenMetadataKey, token))
	stream, err := client.Connect(streamCtx)
	if err != nil {
		return 0, fmt.Errorf("connect to scheduler: %w", err)
	}

	sendHeartbeat := func() error {
		*sequence++
		return stream.Send(&v1.RunnerMessage{
			Payload: &v1.RunnerMessage_Heartbeat{
				Heartbeat: &v1.RunnerHeartbeat{
					Sequence: *sequence,
				},
			},
		})
	}

	if err := sendHeartbeat(); err != nil {
		return 0, fmt.Errorf("send heartbeat: %w", err)
	}

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return interval, nil
			}
			return 0, fmt.Errorf("receive scheduler message: %w", err)
		}

		ack, ok := msg.GetPayload().(*v1.SchedulerMessage_HeartbeatAck)
		if !ok {
			return 0, fmt.Errorf("unsupported scheduler message type %T", msg.GetPayload())
		}

		if ack.HeartbeatAck.GetSequence() != *sequence {
			return 0, fmt.Errorf(
				"heartbeat ack sequence mismatch: got %d want %d",
				ack.HeartbeatAck.GetSequence(),
				*sequence,
			)
		}

		nextInterval := interval
		if ttl := ack.HeartbeatAck.GetLeaseTtl(); ttl != nil {
			if ttlDuration := ttl.AsDuration(); ttlDuration > 0 {
				nextInterval = ttlDuration
			}
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(nextInterval):
		}

		if err := sendHeartbeat(); err != nil {
			return 0, fmt.Errorf("send heartbeat: %w", err)
		}

		interval = nextInterval
	}
}

func dialScheduler(
	target, baseURL string,
	insecureSkipTLSVerify bool,
) (*grpc.ClientConn, error) {
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
