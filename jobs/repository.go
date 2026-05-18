package jobs

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db,
	}
}

func (r *Repository) ListJobs(ctx context.Context) ([]Job, error) {
	rows, err := r.db.Query(ctx, `SELECT * FROM jobs`)
	if err != nil {
		return nil, err
	}

	var jobs []Job

	for rows.Next() {
		var j Job
		err := rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &j.MaxRetries, &j.LeasedUntil, &j.RunAt, &j.LastError, &j.CreatedAt)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (r *Repository) GetJob(ctx context.Context, jobID int) (Job, error) {
	query := `SELECT * FROM jobs WHERE id = $1;`
	row := r.db.QueryRow(ctx, query, jobID)
	var j Job

	if err := row.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &j.MaxRetries, &j.LeasedUntil, &j.RunAt, &j.LastError, &j.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return j, ErrJobNotFound
		}

		return j, err
	}

	return j, nil
}

type InsertJobParams struct {
	ID      uuid.UUID
	Type    string
	Payload json.RawMessage
}

func (r *Repository) InsertJob(ctx context.Context, params InsertJobParams) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO jobs (id, type, payload) values ($1, $2, $3);`,
		params.ID, params.Type, params.Payload,
	)
	return err
}
