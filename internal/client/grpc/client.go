package grpc

import (
	"context"
	"runtime"
	"time"

	"github.com/leonf08/metrics-yp.git/internal/client/http"
	"github.com/leonf08/metrics-yp.git/internal/client/workerpool"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	proto2 "github.com/leonf08/metrics-yp.git/internal/proto"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/rs/zerolog"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	agent  services.Agent
	log    zerolog.Logger
	config agentconf.Config
}

func NewClient(a services.Agent, l zerolog.Logger, config agentconf.Config) *Client {
	return &Client{
		agent:  a,
		log:    l,
		config: config,
	}
}

func (c *Client) Start(ctx context.Context) error {
	conn, err := grpc.Dial(c.config.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := proto2.NewMetricsClient(conn)

	go c.poll(ctx)

	go c.report(ctx, client)

	<-ctx.Done()
	return ctx.Err()
}

func (c *Client) poll(ctx context.Context) {
	t := time.NewTicker(time.Duration(c.config.PollInt) * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.log.Info().Msg("grpc client - Start - Gathering metrics")
			if err := c.agent.GatherMetrics(ctx); err != nil {
				c.log.Error().Err(err).Msg("GatherMetrics")
			}
		}
	}
}

func (c *Client) report(ctx context.Context, client proto2.MetricsClient) {
	ip, err := http.GetIP()
	if err != nil {
		c.log.Error().Err(err).Msg("GetIP")
		return
	}

	md := metadata.New(map[string]string{"X-Real-IP": ip.String()})
	ctx = metadata.NewOutgoingContext(ctx, md)

	rateLimiter := ratelimit.New(c.config.RateLim)

	t := time.NewTicker(time.Duration(c.config.ReportInt) * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			metrics, err := c.agent.GetMetrics(ctx)
			if err != nil {
				c.log.Error().Err(err).Msg("GetMetrics")
				return
			}

			tasks := make([]workerpool.Task, 0, len(metrics))
			for k, v := range metrics {
				fn := func() error {
					resp, err := client.UpdateMetric(ctx, &proto2.UpdateMetricRequest{
						Metric: &proto2.Metric{
							Id:    k,
							Type:  v.Type,
							Value: v.Val.(float64),
						},
					})

					c.log.Info().Str("response", resp.String()).Msg("grpc client - update metric")
					return err
				}

				tasks = append(tasks, fn)
			}

			pool := workerpool.NewWorkerPool(tasks, runtime.NumCPU(), rateLimiter)
			result := pool.Run()
			for err := range result {
				if err != nil {
					c.log.Error().Err(err).Msg("gRPC client - Start - Send request")
				}
			}
		}
	}
}
