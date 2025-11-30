package services

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"link-availability-checker/internal/utils/closer"
)

type AvailabilityService interface {
	CheckDomainAvailability(ctx context.Context, domain string) (bool, error)
}

type AvailabilityServiceImpl struct {
	httpClient  *http.Client
	dnsResolver *net.Resolver
}

func NewAvailabilityService() AvailabilityService {
	return &AvailabilityServiceImpl{httpClient: &http.Client{
		Timeout: 4 * time.Second,
		Transport: &http.Transport{
			DialContext:         (&net.Dialer{Timeout: 4 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 4 * time.Second,
		},
	}, dnsResolver: &net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: time.Second * 4}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}}
}

func (svc *AvailabilityServiceImpl) CheckDomainAvailability(ctx context.Context, domain string) (bool, error) {
	// Try to resolve DNS first and skip HTTP request if domain does not exist
	_, err := svc.dnsResolver.LookupIPAddr(ctx, domain)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			return false, nil // Domain does not exist
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return false, err
		}
	}

	// Send HEAD request to check domain availability without downloading body
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fmt.Sprintf("https://%s", domain), nil)
	if err != nil {
		return false, nil
	}

	resp, err := svc.httpClient.Do(req)
	if err != nil {
		//var dnsErr *net.DNSError
		//var urlErr *url.Error
		//var opErr *net.OpError
		//switch {
		//case errors.As(err, &dnsErr):
		//	return false, nil // Domain does not exist
		//case errors.As(err, &urlErr):
		//	return false, nil // Probably certificate issues
		//case errors.As(err, &opErr):
		//	return false, nil
		//default:
		//	return false, fmt.Errorf("failed to send GET request: %w", err)
		//} // Treating all errors as domain not available for now
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return false, err
		}
		return false, nil
	}
	defer closer.Close(resp.Body)

	// Some servers don't support HEAD requests, fallback to GET
	if resp.StatusCode == http.StatusMethodNotAllowed {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s", domain), nil)
		if err != nil {
			return false, nil
		}
		resp, err = svc.httpClient.Do(req)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return false, err
			}
			return false, nil
		}
	}

	if resp.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}
