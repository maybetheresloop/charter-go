package csv

import (
	"encoding/csv"
	"errors"
	"io"
	"os"

	"github.com/maybetheresloop/charter-go/passwd"
)

type connector map[string]string
type driver struct{}

var (
	drv passwd.Driver = driver{}
)

func init() {
	passwd.Register("csv", drv)
}

func (drv driver) OpenConnector(dataSourceName string) (passwd.Connector, error) {
	f, err := os.Open(dataSourceName)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)

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
			return nil, errors.New("malformed record")
		}

		c[record[0]] = record[1]
	}

	return c, nil
}

func (c connector) GetPassword(user string) (string, error) {
	pw, ok := c[user]
	if !ok {
		return "", passwd.ErrNotExist
	}
	return pw, nil
}
