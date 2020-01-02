package csv

import (
	"github.com/maybetheresloop/charter-go/passwd"
)

type connector map[string]string
type context struct{}

var (
	ctx passwd.DriverContext = &context{}
)

func init() {
	passwd.Register("csv", ctx)
}

func (ctx context) OpenConnector(dataSourceName string) (passwd.Connector, error) {
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
