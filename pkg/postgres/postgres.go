// Package postgres implements postgres connection.
package postgres

import (
	"log"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/jmoiron/sqlx"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres -.
type Postgres struct {
	Builder squirrel.StatementBuilderType
	DB      *sqlx.DB
}

// New -.
func New(url string) (*Postgres, error) {
	_, err := otelsql.Register("pgx", otelsql.WithAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "postgres(schedule-api)"),
	), otelsql.WithTracerProvider(otel.GetTracerProvider()))
	if err != nil {
		return nil, err
	}

	db, err := otelsql.Open("pgx", url, otelsql.WithAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "postgres(schedule-api)"),
	))

	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatalf("db.Ping(): %v", err)
	}

	xdb := sqlx.NewDb(db, "pgx")

	if err := xdb.Ping(); err != nil {
		return nil, err
	}

	pg := &Postgres{
		DB:      xdb,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	return pg, nil
}

// Close -.
func (p *Postgres) Close() {
	if p.DB != nil {
		err := p.DB.Close()
		if err != nil {
			log.Fatalf("db.Close(): %v", err)
		}
	}
}
