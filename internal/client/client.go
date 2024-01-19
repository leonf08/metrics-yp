package client

import (
	"context"
	"encoding/hex"
	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/rs/zerolog"
	"go.uber.org/ratelimit"
	"runtime"
	"time"
)

const (
	retries  = 3
	delay    = 1 * time.Second
	maxDelay = 5 * time.Second
)

type Client struct {
	client *resty.Client
	agent  services.Agent
	signer services.Signer
	log    zerolog.Logger
	config agentconf.Config
}

func NewClient(cl *resty.Client, a services.Agent, s services.Signer,
	l zerolog.Logger, config agentconf.Config) *Client {
	return &Client{
		client: cl,
		agent:  a,
		signer: s,
		log:    l,
		config: config,
	}
}

func (c *Client) Start(ctx context.Context) {
	// Gather metrics
	go c.poll(ctx)

	// Report metrics
	go c.report(ctx)

	<-ctx.Done()
}

func (c *Client) poll(ctx context.Context) {
	t := time.NewTicker(time.Duration(c.config.PollInt) * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.log.Info().Msg("client - Start - Gathering metrics")
			if err := c.agent.GatherMetrics(ctx); err != nil {
				c.log.Error().Err(err).Msg("GatherMetrics")
			}
		}
	}
}

func (c *Client) report(ctx context.Context) {
	if c.config.Mode == "batch" {
		c.client.SetBaseURL("http://" + c.config.Addr + "/updates")
	} else {
		c.client.SetBaseURL("http://" + c.config.Addr + "/update")
	}

	c.client.OnBeforeRequest(func(cl *resty.Client, r *resty.Request) error {
		c.log.Info().Str("method", r.Method).Str("url", r.URL).Msg("sending request")
		return nil
	})

	c.client.OnAfterResponse(func(cl *resty.Client, r *resty.Response) error {
		c.log.Info().Str("status", r.Status()).Str("body", string(r.Body())).Msg("received response")
		return nil
	})

	if c.config.Mode == "query" {
		c.client.SetHeader("Content-Type", "text/plain")
	} else {
		c.client.SetHeaders(map[string]string{
			"Content-Type":     "application/json",
			"Accept":           "application/json",
			"Content-Encoding": "gzip",
		})
	}

	c.client.
		SetRetryCount(retries).
		SetRetryWaitTime(delay).
		SetRetryMaxWaitTime(maxDelay)

	rateLimiter := ratelimit.New(c.config.RateLim)

	t := time.NewTicker(time.Duration(c.config.ReportInt) * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			payload, err := c.agent.ReportMetrics(ctx)
			if err != nil {
				c.log.Error().Err(err).Msg("ReportMetrics")
				return
			}

			if c.config.Mode == "batch" {
				_, err = c.client.R().SetBody(payload[0]).SetContext(ctx).Post("")
				if err != nil {
					c.log.Error().Err(err).Msg("client - Start - Send batch request")
				}
			} else {
				tasks := make([]task, 0, len(payload))
				var fn task
				for _, p := range payload {
					p := p
					if c.config.Mode == "json" {
						if c.signer != nil {
							hash, err := c.signer.CalcHash([]byte(p))
							if err != nil {
								c.log.Error().Err(err).Msg("client - Start - CalcHash")
								return
							}

							c.client.SetHeader("HashSHA256", hex.EncodeToString(hash))
						}

						fn = func() error {
							_, err := c.client.R().
								SetBody(p).
								SetContext(ctx).
								Post("")
							return err
						}
					} else {
						fn = func() error {
							_, err := c.client.R().
								SetPathParam("path", p).
								SetContext(ctx).
								Post("/{path}")
							return err
						}
					}

					tasks = append(tasks, fn)
				}

				pool := newWorkerPool(tasks, runtime.NumCPU(), rateLimiter)
				result := pool.run()
				for err := range result {
					if err != nil {
						c.log.Error().Err(err).Msg("client - Start - Send request")
					}
				}
			}
		}
	}
}
