package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/urfave/cli"
)

const (
	AppName        = "charter-pw"
	DefaultBackend = "csv"
	DefaultFile    = "/etc/charterd/passwd.csv"
	Version        = "0.1.0"
)

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Usage = "Manage users for the Charter FTP server"
	app.Description = "Manage users for the Charter FTP server"
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "backend, b",
			Value: DefaultBackend,
		},
		cli.StringFlag{
			Name:  "file, f",
			Value: DefaultFile,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:        "useradd",
			Action:      userAdd,
			Description: "create a new FTP user",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("%s: %v\n", AppName, err)
		os.Exit(1)
	}
}

func tty() (*os.File, error) {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return nil, errors.New("couldn't prompt for password")
	}

	return f, nil
}

func promptPassword(sc *bufio.Scanner) (string, error) {
	fmt.Printf("Password: ")
	pass1, err := terminal.ReadPassword(0)
	if err != nil {
		return "", nil
	}

	fmt.Printf("Please confirm the password: ")
	pass2, err := terminal.ReadPassword(0)
	if err != nil {
		return "", err
	}

	if bytes.Equal(pass1, pass2) {
		return "", errors.New("passwords do not match")
	}

	return string(pass1), nil
}

func userAdd(ctx *cli.Context) error {
	user := ctx.Args().First()
	if user == "" {
		return errors.New("missing login")
	}

	//db, err := passwd.Open(ctx.String("backend"), ctx.String("file"))
	//if err != nil {
	//	return err
	//}
	//defer db.Close()

	// Password prompt. Passwords may only be accepted from a terminal.
	f, err := tty()
	if err != nil {
		return err
	}

	_, err = promptPassword(bufio.NewScanner(f))
	if err != nil {
		return err
	}

	return nil

	//return db.UserAdd(user, pass)
}
