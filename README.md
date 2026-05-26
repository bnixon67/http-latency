# http-latency

`http-latency` is a small Go CLI that performs a single `GET` request and prints a simple latency breakdown for the request lifecycle.

It measures:

- DNS lookup time
- TCP connect time
- TLS handshake time for HTTPS requests
- Time to first response byte
- Total round-trip time, including reading the full response body

It also prints:

- The request method and URL
- The HTTP status
- The response size in bytes
- The `Location` header when the server returns a redirect response

## Behavior

- Only `http://` and `https://` URLs are accepted.
- The request timeout defaults to 5 seconds.
- A custom timeout can be provided with `-timeout`, for example `-timeout 750ms`.
- The timeout must be greater than zero.
- Redirects are not followed. If the server responds with a redirect, the tool reports that response directly.
- The response body is discarded after being read so the full transfer time can be included in the total duration.

## Usage

```bash
go run . https://example.com
```

Or build a binary first:

```bash
go build -o http-latency .
./http-latency https://example.com
```

Set a custom timeout:

```bash
go run . -timeout 2s https://example.com
```

## Example Output

```text
GET https://example.com
Status: 200 OK
Read: 513 bytes

HTTP Latency Breakdown
----------------------
DNS Lookup     : 4.1ms
TCP Connect    : 12.3ms
TLS Handshake  : 18.7ms
First Byte     : 46.2ms
Total Roundtrip: 49.8ms
```

## Implementation Notes

The program uses `net/http/httptrace` hooks to collect request timing information and a request-scoped context timeout to bound total execution time.
