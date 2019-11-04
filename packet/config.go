package packet

import (
	"context"
	"crypto/x509"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/packethost/packngo"
	"golang.org/x/sync/semaphore"
)

const (
	consumerToken = "aZ9GmqHTPtxevvFq9SK3Pi2yr9YCbRzduCSXF2SNem5sjB91mDq7Th3ZwTtRqMWZ"
)

type ProviderConfig struct {
	Client          *packngo.Client
	DeviceCreateSem *semaphore.Weighted
}

var redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

func PacketRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, nil
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false, nil
			}
		}

		// The error is likely recoverable so retry.
		return true, nil
	}
	return false, nil
}

func GetProviderConfig(authToken string, maxDevicesCreate int) *ProviderConfig {
	m := ProviderConfig{}

	httpClient := retryablehttp.NewClient()
	httpClient.RetryWaitMin = time.Second
	httpClient.RetryWaitMax = 30 * time.Second
	httpClient.RetryMax = 10
	httpClient.CheckRetry = PacketRetryPolicy
	httpClient.HTTPClient.Transport = logging.NewTransport(
		"Packet",
		httpClient.HTTPClient.Transport)
	m.Client = packngo.NewClientWithAuth(consumerToken, authToken, httpClient)
	m.DeviceCreateSem = semaphore.NewWeighted(int64(maxDevicesCreate))
	return &m
}
