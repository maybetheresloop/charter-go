package charter

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/maybetheresloop/charter-go/passwd"
)

type Server struct {
	config   *Config
	passwdDb passwd.DB
}

type Config struct {
	Addr          string
	DefaultDir    string
	NoAnonymous   bool
	AnonymousOnly bool
}

type Client struct {
	ctrlConn     net.Conn
	dataConn     net.Conn
	srv          *Server
	response     *bytes.Buffer
	username     string
	rootDir      string
	workingDir   string
	isRegistered bool
}

func (client *Client) handleConn() {
	defer client.ctrlConn.Close()
	r := textproto.NewReader(bufio.NewReader(client.ctrlConn))
	for {
		line, err := r.ReadLine()
		if err != nil {
			break
		}

		cmd, err := ParseLine(line)
		if err == ErrLineIsEmpty {
			continue
		} else if err != nil {
			break
		}

		upperCmd := strings.ToUpper(cmd.Command)
		if !client.isRegistered && upperCmd != "USER" && upperCmd != "PASS" {
			client.sendReply(530, "You aren't logged in.")
			continue
		}

		srvCmd, ok := Commands[strings.ToUpper(cmd.Command)]
		if !ok {
			// Send command unrecognized
			client.sendReply(500, "Unknown command.")
			continue
		}

		// Verify correct arity.
		if len(cmd.Params) < srvCmd.argc {
			client.sendReply(501, "Wrong number of arguments.")
			continue
		}

		isExiting := srvCmd.handler(client, cmd)
		if isExiting {
			break
		}
	}
}

func (client *Client) sendReply(code int, format string, args ...interface{}) error {
	client.response.Reset()

	_, _ = fmt.Fprintf(client.response, "%d ", code)
	_, _ = fmt.Fprintf(client.response, format, args...)
	client.bufferCrlf()

	_, err := client.ctrlConn.Write(client.response.Bytes())
	return err
}

func (client *Client) bufferCrlf() {
	client.response.WriteByte(byte('\r'))
	client.response.WriteByte(byte('\n'))
}

func (client *Client) verifyDir(dir string) error {
	realDir := filepath.Join(client.rootDir, dir)
	return verifyDir(realDir)
}

func verifyDir(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return err.(*os.PathError).Err
	}

	if !stat.IsDir() {
		return ErrNotDir
	}

	return nil
}

func NewServer(config *Config) *Server {
	srv := &Server{
		config: config,
	}

	return srv
}

func (srv *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", srv.config.Addr)
	if err != nil {
		return err
	}

	return srv.Serve(lis)
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
		srv:        srv,
		response:   &bytes.Buffer{},
		rootDir:    srv.config.DefaultDir,
		workingDir: "/",
	}
}
