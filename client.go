package client

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	httpdialer "github.com/mwitkow/go-http-dialer"
	tls "github.com/refraction-networking/utls"
)

var (
	DefaultBaseURL = ""
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Option func(c *Client)

type Proxy struct {
	host     string
	port     string
	username string
	password string
}

type BypassJA3Transport struct {
	tr1         http.Transport
	tr2         http2.Transport
	mu          sync.RWMutex
	clientHello tls.ClientHelloID
	conn        net.Conn
}

var proxies []Proxy

func NewBypassJA3Transport(helloID tls.ClientHelloID) *BypassJA3Transport {
	return &BypassJA3Transport{clientHello: helloID}
}

func (b *BypassJA3Transport) getTLSConfig(req *http.Request) *tls.Config {
	return &tls.Config{
		ServerName:         req.URL.Host,
		InsecureSkipVerify: true,
	}
}

func (b *BypassJA3Transport) tlsConnect(conn net.Conn, req *http.Request) (*tls.UConn, error) {
	b.mu.RLock()
	tlsConn := tls.UClient(conn, b.getTLSConfig(req), b.clientHello)
	b.mu.RUnlock()

	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("tls handshake fail: %w", err)
	}
	return tlsConn, nil
}

func (b *BypassJA3Transport) SetClientHello(hello tls.ClientHelloID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clientHello = hello
}

func (b *BypassJA3Transport) getConn(req *http.Request, port string) error {
	if len(proxies) > 0 {
		for _, proxy := range proxies {
			uri, err := url.Parse(fmt.Sprintf("http://%s:%s", proxy.host, proxy.port))
			if err != nil {
				return fmt.Errorf("%v", err)
			}

			tun := httpdialer.New(uri, httpdialer.WithConnectionTimeout(time.Second*10))
			conn, err := tun.Dial("tcp", fmt.Sprintf("%s:%s", req.URL.Host, port))
			if err != nil {
				fmt.Println("invalid proxy")
				return fmt.Errorf("tcp net dial fail: %w", err)
			}
			if err != nil {
				return err
			}
			b.conn = conn
			return nil
		}
		return errors.New("no valid proxies")
	} else {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", req.URL.Host, port))
		if err != nil {
			return err
		}
		b.conn = conn
		return nil
	}
}

func (b *BypassJA3Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	switch r.URL.Scheme {
	case "https":
		if err := b.getConn(r, "443"); err != nil {
			return nil, err
		}
		return b.httpsRoundTrip(r)
	case "http":
		if err := b.getConn(r, "80"); err != nil {
			return nil, err
		}
		return b.httpRoundTrip(r)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", r.URL.Scheme)
	}
}

func (b *BypassJA3Transport) httpRoundTrip(req *http.Request) (*http.Response, error) {
	if err := req.Write(b.conn); err != nil {
		panic(err)
	}
	return http.ReadResponse(bufio.NewReader(b.conn), req)
}

func (b *BypassJA3Transport) httpsRoundTrip(req *http.Request) (*http.Response, error) {
	tlsConn, err := b.tlsConnect(b.conn, req)
	if err != nil {
		return nil, fmt.Errorf("tls connect fail: %w", err)
	}
	defer tlsConn.Close()

	httpVersion := tlsConn.ConnectionState().NegotiatedProtocol
	switch httpVersion {
	case "h2":
		conn, err := b.tr2.NewClientConn(tlsConn)
		if err != nil {
			return nil, fmt.Errorf("create http2 client with connection fail: %w", err)
		}
		return conn.RoundTrip(req)
	case "http/1.1", "":
		err := req.Write(tlsConn)
		if err != nil {
			return nil, fmt.Errorf("write http1 tls connection fail: %w", err)
		}
		return http.ReadResponse(bufio.NewReader(tlsConn), req)
	default:
		return nil, fmt.Errorf("unsuported http version: %s", httpVersion)
	}
}

func NewClient(opts ...Option) *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
	}

	client := &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: NewBypassJA3Transport(tls.HelloChrome_102),
			Jar:       jar,
		},
	}

	for _, o := range opts {
		o(client)
	}
	return client
}

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithProxy(p []string) Option {
	return func(c *Client) {
		proxies = parseProxies(p)
	}
}
