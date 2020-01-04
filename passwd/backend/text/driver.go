// Package text implements an authentication backend based on /etc/passwd.
// Files managed by this backend store user information in one line per
// user. Each line is of the following form.
//
// 		<account>:<password>
//
// Passwords are stored as base64-encoded bcrypt hashes.
package text

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/maybetheresloop/charter-go/passwd"
)

type connector map[string]string

type driver struct{}

var drv passwd.Driver = driver{}

// Errors that can be returned by the passwd file parser.
var (
	ErrMalformedRecord = errors.New("malformed record")
)

func init() {
	passwd.Register("text", drv)
}

// reader returns a *csv.Reader set up specifically for reading passwd files.
func reader(rd io.Reader) *csv.Reader {
	r := csv.NewReader(rd)
	r.Comma = ':'
	r.Comment = '#'
	return r
}

// openReader parses user information lines into a map and returns it in
// the form of a passwd.Connector.
func openReader(rd io.Reader) (passwd.Connector, error) {
	r := reader(rd)
	c := make(connector)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) != 2 {
			return nil, ErrMalformedRecord
		}

		c[record[0]] = record[1]
	}

	return c, nil
}

// OpenConnector opens the passwd file, parses it, and returns a handle to it
// in the form of a passwd.Connector.
func (drv driver) OpenConnector(dataSourceName string) (passwd.Connector, error) {
	f, err := os.Open(dataSourceName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return openReader(f)
}

// GetPassword retrieves the password of the specified user.
func (c connector) GetPassword(user string) (string, error) {
	pw, ok := c[user]
	if !ok {
		return "", passwd.ErrNotExist
	}
	return pw, nil
}

// CheckUserPassword verifies that the specified password matches that of the
// user. Returns true if the password is correct, false if not.
func (c connector) CheckUserPassword(user string, pass string) error {
	guessHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if base64.StdEncoding.EncodeToString(guessHash) != c[user] {
		return passwd.ErrIncorrectPassword
	}

	return nil
}