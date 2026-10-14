package togo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync/atomic"
)

// fakeDriver is a minimal database/sql driver so Kernel.SQL can be tested
// without a real database. The DSN selects behaviour: "ping-fail" opens a
// connection whose Ping fails, "open-fail" fails to open a connection at all.
type fakeDriver struct{ open atomic.Int32 }

var testDriver = &fakeDriver{}

func init() { sql.Register("togo-fake", testDriver) }

var (
	errFakeOpen = errors.New("fake: open failed")
	errFakePing = errors.New("fake: ping failed")
)

func (d *fakeDriver) Open(dsn string) (driver.Conn, error) {
	if dsn == "open-fail" {
		return nil, errFakeOpen
	}
	d.open.Add(1)
	return &fakeConn{d: d, pingFail: dsn == "ping-fail"}, nil
}

type fakeConn struct {
	d        *fakeDriver
	pingFail bool
}

func (c *fakeConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("fake: not supported") }
func (c *fakeConn) Begin() (driver.Tx, error)           { return nil, errors.New("fake: not supported") }
func (c *fakeConn) Close() error                        { c.d.open.Add(-1); return nil }

func (c *fakeConn) Ping(context.Context) error {
	if c.pingFail {
		return errFakePing
	}
	return nil
}
