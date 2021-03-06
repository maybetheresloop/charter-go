// Package text implements an authentication backend based on /etc/passwd.
// Files managed by this backend store user information in one line per
// user. Each line is of the following form.
//
// 		<account>:<password>:<home directory>
//
// Passwords are stored as base64-encoded bcrypt hashes.
package text

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"io"
	"os"

	"github.com/maybetheresloop/charter-go/passwd"
	"golang.org/x/crypto/bcrypt"
)

type userInfo struct {
	pass    string // Base64-encoded bcrypt hash of the user's password.
	homeDir string
}

type connector struct {
	filename string
	users    []string
	userInfo map[string]*userInfo
}

func recordFromUserInfo(user string, info *userInfo) []string {
	var record []string
	record = append(record, user)
	record = append(record, info.pass)

	return record
}

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
func readUsers(rd io.Reader) (*connector, error) {
	r := reader(rd)
	info := make(map[string]*userInfo)
	var users []string

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

		users = append(users, record[0])
		info[record[0]] = &userInfo{
			pass:    record[1],
			homeDir: "",
		}
	}

	return &connector{
		users:    users,
		userInfo: info,
	}, nil
}

// OpenConnector opens the passwd file, parses it, and returns a handle to it
// in the form of a passwd.Connector.
func (drv driver) OpenConnector(dataSourceName string) (passwd.Connector, error) {
	f, err := os.Open(dataSourceName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	c, err := readUsers(f)
	if err != nil {
		return nil, err
	}
	c.filename = dataSourceName

	return c, nil
}

// GetPassword retrieves the password of the specified user.
func (c *connector) GetPassword(user string) (string, error) {
	pw, ok := c.userInfo[user]
	if !ok {
		return "", passwd.ErrNotExist
	}
	return pw.pass, nil
}

// CheckUserPassword verifies that the specified password matches that of the
// user. Returns true if the password is correct, false if not.
func (c *connector) CheckUserPassword(user string, pass string) error {
	guessHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if base64.StdEncoding.EncodeToString(guessHash) != c.userInfo[user].pass {
		return passwd.ErrIncorrectPassword
	}

	return nil
}

// Sync guarantees that the changes made to the connector are persisted to disk.
func (c *connector) Sync() error {
	// Re-open the database file and write the contents of the connector.
	f, err := os.OpenFile(c.filename, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)

	for _, user := range c.users {
		userInfo := c.userInfo[user]
		record := recordFromUserInfo(user, userInfo)
		if err := w.Write(record); err != nil {
			return err
		}
	}

	w.Flush()
	return nil
}
