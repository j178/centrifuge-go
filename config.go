package centrifuge

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Config contains various client options.
type Config struct {
	// Token for a connection authentication.
	Token string
	// Data is an arbitrary data which can be sent to a server in a Connect command.
	// Make sure it's a valid JSON when using JSON protocol client.
	Data []byte
	// Name allows setting client name. You should only use a limited
	// amount of client names throughout your applications – i.e. don't
	// make it unique per user for example, this name semantically represents
	// an environment from which client connects.
	// Zero value means "go".
	Name string
	// Version allows setting client version. This is an application
	// specific information. By default, no version set.
	Version string
	// NetDialContext specifies the dial function for creating TCP connections. If
	// NetDialContext is nil, net.DialContext is used.
	NetDialContext func(ctx context.Context, network, addr string) (net.Conn, error)
	// PrivateChannelPrefix is private channel prefix.
	// Zero value means $.
	PrivateChannelPrefix string
	// ReadTimeout is how long to wait read operations to complete.
	// Zero value means 5 * time.Second.
	ReadTimeout time.Duration
	// WriteTimeout is Websocket write timeout.
	// Zero value means 1 * time.Second.
	WriteTimeout time.Duration
	// HandshakeTimeout specifies the duration for the handshake to complete.
	// Zero value means 1 * time.Second.
	HandshakeTimeout time.Duration
	// MaxServerPingDelay used to set maximum delay of ping from server.
	// Zero value means 10 * time.Second.
	MaxServerPingDelay time.Duration
	// TLSConfig specifies the TLS configuration to use with tls.Client.
	// If nil, the default configuration is used.
	TLSConfig *tls.Config
	// EnableCompression specifies if the client should attempt to negotiate
	// per message compression (RFC 7692). Setting this value to true does not
	// guarantee that compression will be supported. Currently, only "no context
	// takeover" modes are supported.
	EnableCompression bool
	// CookieJar specifies the cookie jar.
	// If CookieJar is nil, cookies are not sent in requests and ignored
	// in responses.
	CookieJar http.CookieJar
	// Header specifies custom HTTP Header to send.
	Header http.Header
	// GetConnectionToken called to get or refresh connection token.
	GetConnectionToken func(ConnectionTokenEvent) (string, error)
	// GetSubscriptionToken called to get or refresh private channel subscription token.
	GetSubscriptionToken func(SubscriptionTokenEvent) (string, error)
}
