package passwd

import (
	"errors"
	"fmt"
)

type unknownDriverError string

func (err unknownDriverError) Error() string {
	return fmt.Sprintf("unknown driver: %s", err)
}

func errUnknownDriver(driverName string) error {
	return unknownDriverError(driverName)
}

var (
	drivers = make(map[string]Driver)
)

var (
	ErrNotExist          = errors.New("user does not exist")
	ErrIncorrectPassword = errors.New("incorrect password")
)

type Connector interface {
	GetPassword(user string) (string, error)
	CheckUserPassword(user string, pass string) error
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

func (db *DB) UserAdd(user string, pass string) error {
	return nil
}

func (db *DB) Commit() error {
	return nil
}

func (db *DB) Close() error {
	if err := db.Commit(); err != nil {
		return err
	}

	return nil
}

func Open(driverName, dataSourceName string) (*DB, error) {
	driver, ok := drivers[driverName]
	if !ok {
		return nil, errUnknownDriver(driverName)
	}

	conn, err := driver.OpenConnector(dataSourceName)
	if err != nil {
		return nil, err
	}

	return &DB{conn}, nil
}

func Register(driverName string, driver Driver) {
	drivers[driverName] = driver
}
