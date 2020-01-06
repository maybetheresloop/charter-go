package charter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/maybetheresloop/charter-go/passwd"
)

type transmissionMode int
type dataType int

const (
	ModeStream transmissionMode = iota
	ModeBlock
	ModeCompressed
)

const (
	TypeASCII dataType = iota
	TypeImage
)

type dataConnListener struct {
	active bool
	lis    net.Listener
}

type Server struct {
	auth                []auth
	config              *Config
	passwdDb            passwd.DB
	dataConnListenersMu sync.Mutex
	dataConnListeners   map[uint16]*dataConnListener
}

type auth struct {
	connector passwd.Connector
}

type BackendConf struct {
	Name           string
	DataSourceName string `toml:"data-source-name"`
}

type PassivePortRange struct {
	From uint16
	To   uint16
}

type Config struct {
	Addr             string
	DefaultDir       string
	NoAnonymous      bool `toml:"no-anonymous"`
	AnonymousOnly    bool `toml:"anonymous-only"`
	Backend          []BackendConf
	PassivePortRange PassivePortRange
}

// sendASCII copies from src to dst, translating native line endings in src to
// CRLF line endings in dst.
func sendASCII(dst io.Writer, src io.Reader) error {
	return nil
}

// copyASCII copies from src to dst, translating CRLF line endings in src to
// native line endings in dst.
func storeASCII(dst io.Writer, src io.Reader) error {
	bufDst := bufio.NewWriter(dst)
	bufSrc := bufio.NewReader(src)

	for {
		b, err := bufSrc.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Both "\r" and "\r\n" are transformed into "\n".
		if b == '\r' {
			if err := bufDst.WriteByte('\n'); err != nil {
				return err
			}
			b, err := bufSrc.ReadByte()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if b == '\n' {
				continue
			}
			if err := bufDst.WriteByte(b); err != nil {
				return err
			}
		} else {
			if err := bufDst.WriteByte(b); err != nil {
				return err
			}
		}
	}
	return bufDst.Flush()
}

func NewServer(config *Config) *Server {
	srv := &Server{
		config:            config,
		dataConnListeners: make(map[uint16]*dataConnListener),
	}

	return srv
}

func (srv *Server) ListenAndServe() error {
	// Setup control connection listener.
	lis, err := net.Listen("tcp", srv.config.Addr)
	if err != nil {
		return err
	}

	// Setup data connection listeners.
	if err := srv.addDataConnectionListeners(); err != nil {
		return err
	}

	// Serve control connections.
	return srv.Serve(lis)
}

func (srv *Server) addDataConnectionListeners() error {
	rg := srv.config.PassivePortRange

	// Each listener gets its own acceptor goroutine.
	for i := rg.From; i <= rg.To; i++ {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", i))
		if err != nil {
			return err
		}

		srv.dataConnListeners[i] = &dataConnListener{
			active: false,
			lis:    lis,
		}
	}

	return nil
}

func (srv *Server) reserveDataPort() (uint16, net.Listener) {
	srv.dataConnListenersMu.Lock()
	defer srv.dataConnListenersMu.Unlock()

	for port, lis := range srv.dataConnListeners {
		if !lis.active {
			lis.active = true
		}

		return port, lis.lis
	}

	return uint16(0), nil
}

func (srv *Server) releaseDataPort(port uint16) {
	srv.dataConnListenersMu.Lock()
	defer srv.dataConnListenersMu.Unlock()

	srv.dataConnListeners[port].active = false
}

func (srv *Server) Serve(lis net.Listener) error {
	defer lis.Close()
	for {
		conn, err := lis.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {

			}
		}

		client := srv.newClient(conn)
		go client.handleConn()
	}
}

func (srv *Server) newClient(conn net.Conn) *Client {
	return &Client{
		ctrlConn:   conn,
		server:     srv,
		response:   &bytes.Buffer{},
		rootDir:    srv.config.DefaultDir,
		workingDir: "/",
	}
}
