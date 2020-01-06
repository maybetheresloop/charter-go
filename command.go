package charter

import (
	"errors"
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
			handler: modeHandler,
		},
		"RETR": {
			argc:    1,
			handler: notImplementedHandler,
		},
		"STOR": {
			argc:    1,
			handler: storHandler,
		},
		"STOU": {
			argc:    0,
			handler: notImplementedHandler,
		},
		"APPE": {
			argc:    1,
			handler: appeHandler,
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
			handler: nlstHandler,
		},
		"NOOP": {
			argc:    0,
			handler: noopHandler,
		},
	}
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
