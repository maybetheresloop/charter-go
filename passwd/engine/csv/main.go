package csv

import (
	"github.com/maybetheresloop/charter-go/passwd"
)

type connector map[string]string
type driver struct{}

var (
	ctx passwd.Driver = driver{}
)

func init() {
	passwd.Register("csv", ctx)
}

func (d driver) OpenConnector(dataSourceName string) (passwd.Connector, error) {
	c := make(connector)
	return c, nil
}

func (c connector) GetPassword(user string) (string, error) {
	pw, ok := c[user]
	if !ok {
		return "", passwd.ErrNotExist
	}
	return pw, nil
}
