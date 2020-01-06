package charter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	ctrlConn     net.Conn
	dataLis      net.Listener
	dataConn     net.Conn
	dataPort     uint16
	dataType     dataType
	mode         transmissionMode
	server       *Server
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
	formatted := fmt.Sprintf(format, args...)
	lines := strings.Split(formatted, "\n")

	var i int
	for i = 0; i < len(lines)-1; i++ {
		_, _ = fmt.Fprintf(client.response, "%d-%s", code, lines[i])
	}
	_, _ = fmt.Fprintf(client.response, "%d %s", code, lines[i])

	_, err := client.ctrlConn.Write(client.response.Bytes())
	return err
}

func (client *Client) bufferPadding() {
	client.response.WriteByte(byte(' '))
	client.response.WriteByte(byte(' '))
}

func (client *Client) bufferCrlf() {
	client.response.WriteByte(byte('\r'))
	client.response.WriteByte(byte('\n'))
}

func (client *Client) verifyDir(dir string) error {
	realDir := filepath.Join(client.rootDir, dir)
	return verifyDir(realDir)
}

func (client *Client) realPath(path string) string {
	// If the new directory's path is relative, join it with the current working directory. If it's
	// absolute, then use it directly.
	var relDir string
	paramDir := path
	if filepath.IsAbs(paramDir) {
		relDir = paramDir
	} else {
		relDir = filepath.Join(client.workingDir, paramDir)
	}

	// Get the real directory by joining with the client's root directory.
	return filepath.Join(client.rootDir, relDir)
}

func (client *Client) ensureDataConn() bool {
	// If we aren't in passive mode and don't have a data connection listener, abort.
	var err error
	if client.dataLis == nil {
		_ = client.sendReply(425, "No data connection")
		return false
	}

	// Block until we get a data connection.
	client.dataConn, err = client.dataLis.Accept()
	if err != nil {
		_ = client.sendReply(421, "The connection couldn't be accepted")
		return false
	}

	return true
}

func (client *Client) writeFile(filename string, r io.Reader, perm os.FileMode, append bool) error {
	flags := os.O_CREATE | os.O_WRONLY
	if append {
		flags |= os.O_APPEND
	}

	f, err := os.OpenFile(filename, flags, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
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
