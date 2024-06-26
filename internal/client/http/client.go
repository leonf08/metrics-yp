package http

import (
	"context"
	"encoding/hex"
	"net"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/leonf08/metrics-yp.git/internal/client/workerpool"
	"github.com/leonf08/metrics-yp.git/internal/config/agentconf"
	"github.com/leonf08/metrics-yp.git/internal/services"
	"github.com/rs/zerolog"
	"go.uber.org/ratelimit"
)

const (
	retries  = 3
	delay    = 1 * time.Second
	maxDelay = 5 * time.Second
)

// Client is a client for collecting and sending metrics to the server
type Client struct {
	client *resty.Client
	agent  services.Agent
	signer *services.HashSigner
	crypto services.Crypto
	log    zerolog.Logger
	config agentconf.Config
}

// NewClient creates a new client
func NewClient(cl *resty.Client, a services.Agent, s *services.HashSigner, cr services.Crypto,
	l zerolog.Logger, config agentconf.Config) *Client {
	return &Client{
		client: cl,
		agent:  a,
		signer: s,
		crypto: cr,
		log:    l,
		config: config,
	}
}

// Start starts the client
func (c *Client) Start(ctx context.Context) error {
	// Gather metrics
	go c.poll(ctx)

	// Report metrics
	go c.report(ctx)

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

	ip, err := GetIP()
	if err != nil {
		c.log.Error().Err(err).Msg("GetIP")
		return
	}

	c.client.
		SetHeader("X-Real-IP", ip.String()).
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
				_, err = c.client.R().
					SetBody(payload[0]).
					SetContext(ctx).Post("")
				if err != nil {
					c.log.Error().Err(err).Msg("client - Start - Send batch request")
				}
			} else {
				tasks := make([]workerpool.Task, 0, len(payload))
				var fn workerpool.Task
				for _, p := range payload {
					p := p
					r := c.client.R()
					if c.config.Mode == "json" {
						b := []byte(p)
						if c.crypto != nil {
							b, err = c.crypto.Encrypt(b)
							if err != nil {
								c.log.Error().Err(err).Msg("client - Start - Encrypt")
								return
							}
						}
						if c.signer != nil {
							hash, err := c.signer.CalcHash(b)
							if err != nil {
								c.log.Error().Err(err).Msg("client - Start - CalcHash")
								return
							}

							r.SetHeader("HashSHA256", hex.EncodeToString(hash))
						}

						fn = func() error {
							_, err = r.SetBody(b).
								SetContext(ctx).
								Post("")
							return err
						}
					} else {
						fn = func() error {
							_, err = r.SetPathParam("path", p).
								SetContext(ctx).
								Post("/{path}")
							return err
						}
					}

					tasks = append(tasks, fn)
				}

				pool := workerpool.NewWorkerPool(tasks, runtime.NumCPU(), rateLimiter)
				result := pool.Run()
				for err := range result {
					if err != nil {
						c.log.Error().Err(err).Msg("client - Start - Send request")
					}
				}
			}
		}
	}
}

func GetIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
