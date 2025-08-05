// Package postgres implements postgres connection.
package postgres

import (
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
	db, openErr := sqlx.Open("pgx", url)
	if openErr != nil {
		log.Fatalf("sqlx.Open(): %v", openErr)
	}

	err := db.Ping()

	if err != nil {
		log.Fatalf("db.Ping(): %v", err)
	}

	pg := &Postgres{
		DB:      db,
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
