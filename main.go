package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

func main() {
	prog := filepath.Base(os.Args[0])
	log.SetFlags(0)
	log.SetPrefix(prog + ": ")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <url>\n\n", prog)
		fmt.Fprintln(os.Stderr, "Measure HTTP request latency breakdown for a single URL.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}

	var timeoutFlag = flag.String("timeout", "5s", "request timeout duration")
	flag.Parse()
	timeoutDur, err := time.ParseDuration(*timeoutFlag)
	if err != nil {
		log.Fatalf("invalid timeout duration %q: %v", *timeoutFlag, err)
	}
	if timeoutDur <= 0 {
		log.Fatalf("invalid timeout duration %q: must be greater than zero", *timeoutFlag)
	}

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	urlStr := flag.Arg(0)

	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatalf("invalid URL %q: %v", urlStr, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		log.Fatalf("invalid URL %q: must use http or https scheme", urlStr)
	}
	if u.Host == "" {
		log.Fatalf("invalid URL %q: missing host", urlStr)
	}

	fmt.Println(http.MethodGet, urlStr)

	resp, written, t, err := measure(urlStr, timeoutDur)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.Status)

	if location := resp.Header.Get("Location"); location != "" {
		fmt.Printf("Redirect Location: %s\n", location)
	}
	p := message.NewPrinter(language.English)
	p.Printf("Read: %v bytes\n", number.Decimal(written))
	fmt.Println()

	header := "HTTP Latency Breakdown"
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))

	fmt.Printf("DNS Lookup     : %v\n", t.DNSLookup)
	fmt.Printf("TCP Connect    : %v\n", t.TCPConnect)
	if resp.TLS != nil {
		fmt.Printf("TLS Handshake  : %v\n", t.TLSHandshake)
	}
	fmt.Printf("First Byte     : %s\n", t.firstByteString())
	fmt.Printf("Total Roundtrip: %v\n", t.TotalRoundtrip)
}
