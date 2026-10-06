package raff

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// publicCheckInterval is how often WaitForPublicEndpoint retries.
const publicCheckInterval = time.Second

// WaitForPublicEndpoint blocks until the database's public address accepts
// connections, or ctx is done. A database reports running before the public
// gateway has picked it up (usually within 10 to 20 seconds); calling this
// before printing or handing out the connection string means the first
// connection works. It never logs in.
//
// PostgreSQL's public port 5432 is routed by the hostname the client sends in
// TLS, so it completes a TLS handshake with that name; other engines have a
// port of their own, which accepts connections only once it is routed.
// A connection without a public address returns nil at once.
func WaitForPublicEndpoint(ctx context.Context, conn *DatabaseConnection) error {
	if conn == nil || StringValue(conn.PublicHost) == "" || IntValue(conn.PublicPort) == 0 {
		return nil
	}
	host, port := StringValue(conn.PublicHost), IntValue(conn.PublicPort)
	postgres := strings.HasPrefix(StringValue(conn.PublicConnectionURI), "postgres")
	var lastErr error
	for {
		if lastErr = checkPublicEndpoint(ctx, host, port, postgres); lastErr == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("public address %s:%d not reachable yet: %w", host, port, lastErr)
		case <-time.After(publicCheckInterval):
		}
	}
}

func checkPublicEndpoint(ctx context.Context, host string, port int, postgres bool) error {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var d net.Dialer
	raw, err := d.DialContext(dialCtx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return err
	}
	defer func() { _ = raw.Close() }()
	if !postgres {
		return nil
	}
	_ = raw.SetDeadline(time.Now().Add(5 * time.Second))
	// SSLRequest: the gateway answers 'S' and then expects TLS naming the
	// database; an unknown name is refused during the handshake.
	req := make([]byte, 8)
	binary.BigEndian.PutUint32(req[0:4], 8)
	binary.BigEndian.PutUint32(req[4:8], 80877103)
	if _, err := raw.Write(req); err != nil {
		return err
	}
	ans := make([]byte, 1)
	if _, err := io.ReadFull(raw, ans); err != nil {
		return err
	}
	if ans[0] != 'S' {
		return fmt.Errorf("server refused TLS")
	}
	// Reachability only: the certificate is checked by the real client.
	tc := tls.Client(raw, &tls.Config{ServerName: host, InsecureSkipVerify: true}) // #nosec G402
	return tc.HandshakeContext(dialCtx)
}
