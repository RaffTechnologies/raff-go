package raff

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"io"
	"math/big"
	"net"
	"strconv"
	"testing"
	"time"
)

// fakePgGateway answers an SSLRequest with 'S' and completes TLS only for the
// one name it knows, like the public gateway on port 5432.
func fakePgGateway(t *testing.T, known string) (string, int) {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), DNSNames: []string{known},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour)}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				req := make([]byte, 8)
				if _, err := io.ReadFull(c, req); err != nil || binary.BigEndian.Uint32(req[4:8]) != 80877103 {
					return
				}
				_, _ = c.Write([]byte{'S'})
				_ = tls.Server(c, &tls.Config{GetConfigForClient: func(h *tls.ClientHelloInfo) (*tls.Config, error) {
					if h.ServerName != known {
						return nil, io.EOF
					}
					return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
				}}).Handshake()
			}(c)
		}
	}()
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	p, _ := strconv.Atoi(port)
	return host, p
}

func TestCheckPublicEndpointPostgresName(t *testing.T) {
	_, port := fakePgGateway(t, "localhost")
	if err := checkPublicEndpoint(context.Background(), "localhost", port, true); err != nil {
		t.Fatalf("known name: %v", err)
	}
	_, port2 := fakePgGateway(t, "other.example")
	if err := checkPublicEndpoint(context.Background(), "localhost", port2, true); err == nil {
		t.Fatal("an unknown name counted as reachable")
	}
}

func TestWaitForPublicEndpointGivesUp(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	_ = ln.Close() // nothing listens: never reachable
	p, _ := strconv.Atoi(port)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	conn := &DatabaseConnection{PublicHost: String("127.0.0.1"), PublicPort: Int(p), PublicConnectionURI: String("mysql://x")}
	if err := WaitForPublicEndpoint(ctx, conn); err == nil {
		t.Fatal("closed port reported reachable")
	}
	if err := WaitForPublicEndpoint(ctx, &DatabaseConnection{}); err != nil {
		t.Fatalf("no public address should return nil: %v", err)
	}
}
