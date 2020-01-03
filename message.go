package charter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrLineIsEmpty = errors.New("line is empty")
	ErrNotDir      = errors.New("not a directory")
)

type commandHandler func(client *Client, command FtpCommand) (isExiting bool)

type Command struct {
	argc    int
	handler commandHandler
}

var Commands map[string]Command

func init() {
	Commands = map[string]Command{
		"USER": {
			argc:    1,
			handler: userHandler,
		},
		"PASS": {
			argc:    1,
			handler: passHandler,
		},
		"ACCT": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"CWD": {
			argc:    1,
			handler: cwdHandler,
		},
		"CDUP": {
			argc:    0,
			handler: cdupHandler,
		},
		"SMNT": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"QUIT": {
			argc:    0,
			handler: quitHandler,
		},
		"REIN": {
			argc:    0,
			handler: notImplementedHandler,
		},
		"PORT": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"PASV": {
			argc:    0,
			handler: pasvHandler,
		},
		"TYPE": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"STRU": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"MODE": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"RETR": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"STOR": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"STOU": {
			argc:    0,
			handler: notImplementedHandler,
		},
		"APPE": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"ABOR": {
			argc:    0,
			handler: notImplementedHandler,
		},
		"DELE": {
			argc:    1,
			handler: deleHandler,
		},
		"RMD": {
			argc:    1,
			handler: rmdHandler,
		},
		"MKD": {
			argc:    1,
			handler: mkdHandler,
		},
		"PWD": {
			argc:    0,
			handler: pwdHandler,
		},
		"LIST": {
			argc:    0,
			handler: listHandler,
		},
		"NLST": {
			argc:    0,
			handler: notImplementedHandler,
		},
	}
}

func rmdHandler(client *Client, command FtpCommand) (isExiting bool) {
	realDir := client.realPath(command.Params[0])
	if err := os.Remove(realDir); err != nil {
		_ = client.sendReply(550, "Can't remove directory: %v", err.(*os.PathError).Err)
	} else {
		_ = client.sendReply(250, "The directory was successfully removed")
	}

	return
}

func deleHandler(client *Client, command FtpCommand) (isExiting bool) {
	paramPath := command.Params[0]
	realPath := client.realPath(paramPath)
	if stat, err := os.Stat(realPath); err != nil {
		_ = client.sendReply(550, "Could not delete %s: %v", paramPath, err.(*os.PathError).Err)
	} else if stat.IsDir() {
		_ = client.sendReply(550, "Could not delete %s: Invalid argument", paramPath)
	}

	if err := os.Remove(realPath); err != nil {
		_ = client.sendReply(550, "Could not delete %s: %v", paramPath, err.(*os.PathError).Err)
	}

	return
}

func mkdHandler(client *Client, command FtpCommand) (isExiting bool) {
	realDir := client.realPath(command.Params[0])
	if err := os.Mkdir(realDir, 0755); err != nil {
		_ = client.sendReply(550, "Can't create directory: %v", err.(*os.PathError).Err)
	} else {
		_ = client.sendReply(257, "%q : The directory was successfully created", paramDir)
	}

	return
}

func userHandler(client *Client, command FtpCommand) (isExiting bool) {
	client.username = command.Params[0]
	_ = client.sendReply(331, "User %s OK. Password required", client.username)
	return
}

func passHandler(client *Client, command FtpCommand) (isExiting bool) {
	if client.username == "" {
		_ = client.sendReply(503, "Login with USER first.")
	} else {
		client.isRegistered = true
		client.sendReply(230, "OK. Current directory is %s", client.workingDir)
	}
	return
}

func cwdHandler(client *Client, command FtpCommand) (isExiting bool) {
	var newDir string
	paramDir := command.Params[0]

	// If the new directory's path is relative, join it with the current working directory. If it's
	// absolute, then use it directly.
	if filepath.IsAbs(paramDir) {
		newDir = paramDir
	} else {
		newDir = filepath.Join(client.workingDir, paramDir)
	}

	// Verify new working directory exists.
	if err := client.verifyDir(newDir); err != nil {
		_ = client.sendReply(550, "Can't change directory to %s: %v", newDir, err)
		return
	}

	// Change the working directory.
	client.workingDir = newDir
	_ = client.sendReply(250, "OK. Current directory is %s", client.workingDir)
	return
}

func cdupHandler(client *Client, command FtpCommand) (isExiting bool) {
	// We don't do any verification of the parent directory; if the parent directory isn't valid,
	// then other commands simply won't work, and the user must change to a valid directory.
	client.workingDir = filepath.Dir(client.workingDir)
	client.sendReply(250, "OK. Current directory is %s", client.workingDir)
	return
}

func pasvHandler(client *Client, command FtpCommand) (isExiting bool) {
	prevPort := client.dataPort
	client.dataPort, client.dataLis = client.server.reserveDataPort()

	// If we were previously listening on a port, release it.
	if prevPort != 0 {
		client.server.releaseDataPort(prevPort)
	}

	if client.dataLis == nil {
		_ = client.sendReply(421, "No ports available for passive mode")
		return true
	}

	_ = client.sendReply(227, "Entering Passive Mode (127,0,0,1,%d,%d)",
		client.dataPort&0xFF00>>8, client.dataPort&0x00FF)
	return false
}

func listHandler(client *Client, command FtpCommand) (isExiting bool) {
	// If we aren't in passive mode and don't have a data connection, abort.
	var err error
	if client.dataLis == nil {
		_ = client.sendReply(425, "No data connection")
		return
	}

	// Block until we get a data connection.
	client.dataConn, err = client.dataLis.Accept()
	if err != nil {
		_ = client.sendReply(421, "The connection couldn't be accepted")
		return
	}

	// Send over the data connection.
	_, _ = fmt.Fprintf(client.dataConn, "list")
	_ = client.dataConn.Close()

	return false
}

func nlstHandler(client *Client, command FtpCommand) (isExiting bool) {
	return
}

func pwdHandler(client *Client, command FtpCommand) (isExiting bool) {
	client.sendReply(257, "%q is your current location", client.workingDir)
	return
}

func quitHandler(client *Client, command FtpCommand) bool {
	//

	return true
}

func notImplementedHandler(client *Client, command FtpCommand) bool {
	_ = client.sendReply(502, "Command not implemented.")
	return false
}

type FtpCommand struct {
	Command string
	Params  []string
}

// ParseLine parses an FTP command from the given FTP line.
func ParseLine(line string) (FtpCommand, error) {
	arr := strings.Fields(line)
	if len(arr) == 0 {
		return FtpCommand{}, ErrLineIsEmpty
	}

	return FtpCommand{
		Command: arr[0],
		Params:  arr[1:],
	}, nil
}
