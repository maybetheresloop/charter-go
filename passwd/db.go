package passwd

import "errors"

var (
	drivers = make(map[string]Driver)
)

var (
	ErrNotExist = errors.New("user does not exist")
)

type Connector interface {
	GetPassword(user string) (string, error)
}

type Driver interface {
	OpenConnector(dataSourceName string) (Connector, error)
}

type DB struct {
	connector Connector
}

func (db *DB) GetPassword(user string) (string, error) {
	return db.connector.GetPassword(user)
}

func Open(driverName, dataSourceName string) (*DB, error) {
	return nil, nil
}

func Register(driverName string, driver Driver) {
	drivers[driverName] = driver
}
