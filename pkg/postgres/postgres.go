// Package postgres implements postgres connection.
package postgres

import (
	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	// maxPoolSize  int
	// connAttempts int
	//connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	// Pool    *pgxpool.Pool
	DB *sqlx.DB
}

// New -.
func New(url string) (*Postgres, error) {

	//db, openErr := sqlx.Open("pgx", url)
	_, err := otelsql.Register("pgx", otelsql.WithAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "postgres(schedule-api)"),
	), otelsql.WithTracerProvider(otel.GetTracerProvider()))
	if err != nil {
		return nil, err
	}

	// Пример использования otelsql
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

	// Проверяем соединение
	if err := xdb.Ping(); err != nil {
		return nil, err
	}

	pg := &Postgres{
		DB:      xdb,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Custom options
	// for _, opt := range opts {
	//	opt(pg)
	//}

	// poolConfig, err := pgxpool.ParseConfig(url)
	// if err != nil {
	//	return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	//}
	//
	//poolConfig.MaxConns = int32(pg.maxPoolSize)
	//
	//for pg.connAttempts > 0 {
	//	//pgxpool.
	//	pg.Pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	//	if err == nil {
	//		break
	//	}
	//
	//	log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)
	//
	//	time.Sleep(pg.connTimeout)
	//
	//	pg.connAttempts--
	//}

	// if err != nil {
	//	return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	//}

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
