package relay

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc"
	"github.com/Zhang-986/Stardew-Sage/echofarm-core/kitex_gen/echofarmrpc/echofarmgateway"
	"github.com/cloudwego/gopkg/bufiox"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans/gonet"
	"github.com/cloudwego/kitex/transport"
)

type kitexGatewayClient struct {
	client echofarmgateway.Client
}

type pinnedTLSDialer struct {
	fingerprint [sha256.Size]byte
	now         func() time.Time
}

// gonetTLSConn supplies the buffered reader and writer expected by Kitex's
// gonet transport while retaining the authenticated TLS connection underneath.
type gonetTLSConn struct {
	net.Conn
	reader    *bufiox.DefaultReader
	writer    *bufiox.DefaultWriter
	closeOnce sync.Once
	closeErr  error
}

func newGonetTLSConn(connection net.Conn) *gonetTLSConn {
	return &gonetTLSConn{
		Conn:   connection,
		reader: bufiox.NewDefaultReader(connection),
		writer: bufiox.NewDefaultWriter(connection),
	}
}

func (c *gonetTLSConn) Reader() *bufiox.DefaultReader { return c.reader }
func (c *gonetTLSConn) Writer() *bufiox.DefaultWriter { return c.writer }
func (c *gonetTLSConn) Read(body []byte) (int, error) { return c.reader.Read(body) }

func (c *gonetTLSConn) Close() error {
	c.closeOnce.Do(func() {
		_ = c.reader.Release(nil)
		c.closeErr = c.Conn.Close()
	})
	return c.closeErr
}

func NewKitexClient(config Config) (GatewayClient, error) {
	dialer := &pinnedTLSDialer{fingerprint: config.CertSHA256, now: time.Now}
	generated, err := echofarmgateway.NewClient(
		"EchoFarmGateway",
		client.WithHostPorts(config.UpstreamAddress),
		client.WithTransportProtocol(transport.TTHeaderFramed),
		client.WithDialer(remote.Dialer(dialer)),
		client.WithTransHandlerFactory(gonet.NewCliTransHandlerFactory()),
		client.WithConnectTimeout(5*time.Second),
		client.WithRPCTimeout(config.RequestTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("create Kitex gateway client: %w", err)
	}
	return &kitexGatewayClient{client: generated}, nil
}

func (c *kitexGatewayClient) Call(ctx context.Context, operation Operation, request *echofarmrpc.RelayRequest) (*echofarmrpc.RelayResponse, error) {
	switch operation {
	case OperationHealth:
		return c.client.Health(ctx, request)
	case OperationLearn:
		return c.client.Learn(ctx, request)
	case OperationNextAction:
		return c.client.NextAction(ctx, request)
	case OperationActionResult:
		return c.client.ActionResult_(ctx, request)
	case OperationCorrection:
		return c.client.Correction(ctx, request)
	case OperationPlayerModel:
		return c.client.PlayerModel(ctx, request)
	case OperationSkill:
		return c.client.Skill(ctx, request)
	case OperationMemory:
		return c.client.Memory(ctx, request)
	case OperationModelUsage:
		return c.client.ModelUsage(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported relay operation %q", operation)
	}
}

func (d *pinnedTLSDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	raw, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	connection := tls.Client(raw, &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, // The exact leaf certificate is verified below.
		VerifyConnection: func(state tls.ConnectionState) error {
			return verifyPinnedCertificate(state, d.fingerprint, d.now())
		},
	})
	if err := connection.SetDeadline(time.Now().Add(timeout)); err != nil {
		_ = raw.Close()
		return nil, err
	}
	if err := connection.Handshake(); err != nil {
		_ = raw.Close()
		return nil, err
	}
	if err := connection.SetDeadline(time.Time{}); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return newGonetTLSConn(connection), nil
}

func verifyPinnedCertificate(state tls.ConnectionState, expected [32]byte, now time.Time) error {
	if len(state.PeerCertificates) == 0 {
		return errors.New("upstream certificate is missing")
	}
	certificate := state.PeerCertificates[0]
	if now.Before(certificate.NotBefore) || now.After(certificate.NotAfter) {
		return errors.New("upstream certificate is outside its validity window")
	}
	presented := sha256.Sum256(certificate.Raw)
	if subtle.ConstantTimeCompare(expected[:], presented[:]) != 1 {
		return errors.New("upstream certificate fingerprint mismatch")
	}
	return nil
}
