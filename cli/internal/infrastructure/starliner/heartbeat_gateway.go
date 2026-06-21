package starliner

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"starliner.app/runner/internal/domain/port"
	v1 "starliner.app/runner/internal/infrastructure/grpc/proto/v1"
)

const (
	defaultHeartbeatInterval = 5 * time.Second
	minReconnectDelay        = time.Second
	maxReconnectDelay        = 30 * time.Second
)

type HeartbeatGateway struct {
	config       port.ConfigStore
	registration port.RegistrationStore
	workload     port.WorkloadTracker
}

func NewHeartbeatGateway(
	config port.ConfigStore,
	registration port.RegistrationStore,
	workload port.WorkloadTracker,
) port.HeartbeatGateway {
	return &HeartbeatGateway{
		config:       config,
		registration: registration,
		workload:     workload,
	}
}

func (g *HeartbeatGateway) OpenHeartbeatSession(
	ctx context.Context,
	insecureSkipTLSVerify bool,
) (port.HeartbeatSession, error) {
	token, err := g.registration.Token()
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("runner not registered; run `runner register` first")
	}

	baseURL, err := g.config.BaseURL()
	if err != nil {
		return nil, err
	}

	target, err := grpcTarget(baseURL)
	if err != nil {
		return nil, err
	}

	conn, err := dialRunnerGRPC(target, baseURL, insecureSkipTLSVerify)
	if err != nil {
		return nil, err
	}

	client := v1.NewRunnerHeartbeatServiceClient(conn)
	return &heartbeatSession{
		conn:         conn,
		client:       client,
		token:        token,
		registration: g.registration,
		workload:     g.workload,
	}, nil
}

type heartbeatSession struct {
	conn         *grpc.ClientConn
	client       v1.RunnerHeartbeatServiceClient
	token        string
	registration port.RegistrationStore
	workload     port.WorkloadTracker
}

func (s *heartbeatSession) Run(ctx context.Context) error {
	var sequence uint64
	interval := defaultHeartbeatInterval
	reconnectDelay := minReconnectDelay

	for {
		if ctx.Err() != nil {
			return nil
		}

		nextInterval, err := s.runHeartbeatSession(ctx, &sequence, interval)
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

func (s *heartbeatSession) runHeartbeatSession(
	ctx context.Context,
	sequence *uint64,
	interval time.Duration,
) (time.Duration, error) {
	streamCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(runnerTokenMetadataKey, s.token))
	stream, err := s.client.StreamHeartbeats(streamCtx)
	if err != nil {
		return 0, fmt.Errorf("open heartbeat stream: %w", err)
	}

	sendHeartbeat := func() error {
		maxJobs, err := s.registration.MaxConcurrentJobs()
		if err != nil {
			return err
		}

		active := int32(s.workload.ActiveJobs())

		*sequence++
		return stream.Send(&v1.HeartbeatMessage{
			Payload: &v1.HeartbeatMessage_Heartbeat{
				Heartbeat: &v1.RunnerHeartbeat{
					Sequence:          *sequence,
					MaxConcurrentJobs: int32(maxJobs),
					ActiveJobs:        active,
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
			return 0, fmt.Errorf("receive heartbeat ack: %w", err)
		}

		ack, ok := msg.GetPayload().(*v1.HeartbeatAckMessage_HeartbeatAck)
		if !ok {
			return 0, fmt.Errorf("unsupported heartbeat ack message type %T", msg.GetPayload())
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

func (s *heartbeatSession) Close() error {
	return s.conn.Close()
}
