package charter

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode"
)

func noopHandler(client *Client, command FtpCommand) (isExiting bool) {
	return
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
	paramDir := command.Params[0]
	realDir := client.realPath(paramDir)
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
		_ = client.sendReply(230, "OK. Current directory is %s", client.workingDir)
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
	return
}

func modeHandler(client *Client, command FtpCommand) (isExiting bool) {
	mode := unicode.ToLower(rune(command.Params[0][0]))
	if mode == 's' {
		_ = client.sendReply(200, "S OK")
	} else {
		_ = client.sendReply(504, "Please use (S)tream mode")
	}
	return
}

func listHandler(client *Client, command FtpCommand) (isExiting bool) {
	if !client.ensureDataConn() {
		return
	}
	defer client.dataConn.Close()

	// Send over the data connection.
	_, _ = fmt.Fprintf(client.dataConn, "list")
	return
}

func appeHandler(client *Client, command FtpCommand) (isExiting bool) {
	if !client.ensureDataConn() {
		return
	}
	defer client.dataConn.Close()

	paramPath := command.Params[0]
	realPath := client.realPath(paramPath)
	err := client.writeFile(realPath, client.dataConn, 0644, true)
	if err != nil {
		if v, ok := err.(*os.PathError); ok {
			_ = client.sendReply(550, "Can't open %s: %v", paramPath, v.Err)
		} else {
			_ = client.sendReply(550, "Can't append to %s: %v", paramPath, err)
		}
	}

	return
}

func storHandler(client *Client, command FtpCommand) (isExiting bool) {
	// Get source data connection.
	if !client.ensureDataConn() {
		return
	}
	defer client.dataConn.Close()

	// Set up destination file.
	paramPath := command.Params[0]
	realPath := client.realPath(paramPath)
	err := client.writeFile(realPath, client.dataConn, 0644, false)
	if err != nil {
		if v, ok := err.(*os.PathError); ok {
			_ = client.sendReply(550, "Can't open %s: %v", paramPath, v.Err)
		} else {
			_ = client.sendReply(550, "Can't store %s: %v", paramPath, err)
		}
	}

	return
}

func nlstHandler(client *Client, command FtpCommand) (isExiting bool) {
	if !client.ensureDataConn() {
		return
	}
	defer client.dataConn.Close()

	// Send over the data connection.
	_, _ = fmt.Fprintf(client.dataConn, "nlst")
	return
}

func pwdHandler(client *Client, command FtpCommand) (isExiting bool) {
	client.sendReply(257, "%q is your current location", client.workingDir)
	return
}

func quitHandler(client *Client, command FtpCommand) bool {
	return true
}

func typeHandler(client *Client, command FtpCommand) bool {
	// Currently support ASCII and Image types.
	if unicode.ToLower(rune(command.Params[1][0])) == 'a' {
		client.sendReply(200, "TYPE is now ASCII")
	} else if unicode.ToLower(rune(command.Params[1][0])) == 'i' {
		client.sendReply(200, "TYPE is now 8-bit binary")
	}

	return false
}

func notImplementedHandler(client *Client, command FtpCommand) bool {
	_ = client.sendReply(502, "Command not implemented.")
	return false
}
