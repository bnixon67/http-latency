package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

type timings struct {
	DNSLookup      time.Duration
	TCPConnect     time.Duration
	TLSHandshake   time.Duration
	FirstByte      time.Duration
	TotalRoundtrip time.Duration
}

func measure(urlStr string, timeoutDur time.Duration) (*http.Response, int64, timings, error) {
	t := timings{}
	var start time.Time
	var dnsStartedAt time.Time
	var connectStartedAt time.Time
	var tlsStartedAt time.Time
	var firstByteTime time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStartedAt = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			if !dnsStartedAt.IsZero() {
				t.DNSLookup = time.Since(dnsStartedAt)
			}
		},
		ConnectStart: func(_, _ string) {
			connectStartedAt = time.Now()
		},
		ConnectDone: func(_, _ string, err error) {
			if err == nil && !connectStartedAt.IsZero() {
				t.TCPConnect = time.Since(connectStartedAt)
			}
		},
		TLSHandshakeStart: func() {
			tlsStartedAt = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err == nil && !tlsStartedAt.IsZero() {
				t.TLSHandshake = time.Since(tlsStartedAt)
			}
		},
		GotFirstResponseByte: func() {
			firstByteTime = time.Now()
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDur)
	defer cancel()

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, 0, timings{}, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start = time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, timings{}, err
	}

	written, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, 0, timings{}, fmt.Errorf("read response body: %w", err)
	}

	if !firstByteTime.IsZero() && !start.IsZero() {
		t.FirstByte = firstByteTime.Sub(start)
	}
	t.TotalRoundtrip = time.Since(start)

	return resp, written, t, nil
}

func (t timings) firstByteString() string {
	if t.FirstByte == 0 {
		return "n/a"
	}

	return t.FirstByte.String()
}
