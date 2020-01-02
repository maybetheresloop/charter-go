package charter

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrNotDir = errors.New("not a directory")
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
		"PWD": {
			argc:    0,
			handler: pwdHandler,
		},
	}
}

func userHandler(client *Client, command FtpCommand) (isExiting bool) {
	client.username = command.Params[0]
	client.sendReply(331, "User %s OK. Password required", client.username)
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

func pwdHandler(client *Client, command FtpCommand) (isExiting bool) {
	client.sendReply(257, "%q is your current location", client.workingDir)
	return
}

func quitHandler(client *Client, command FtpCommand) bool {
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

var (
	ErrLineIsEmpty = errors.New("line is empty")
)

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
