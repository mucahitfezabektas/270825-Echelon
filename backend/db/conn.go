package db

import "github.com/jackc/pgx/v5/pgconn"

func PGConn() *pgconn.PgConn {
	return RawPGConn
}
