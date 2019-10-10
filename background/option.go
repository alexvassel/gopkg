package background

import (
	db "github.com/severgroup-tt/gopkg-database"
)

type OptionFn func(j *job)

func WithDb(conn db.IClient) OptionFn {
	return func(j *job) {
		j.dbConn = conn
	}
}
