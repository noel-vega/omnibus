package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobStatus string

const (
	StatusQueued    JobStatus = "queued"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusDead      JobStatus = "dead"
)

type Job struct {
	ID          uuid.UUID
	Type        string
	Payload     json.RawMessage
	Status      JobStatus
	Attempts    int
	MaxRetries  int
	LeasedUntil *time.Time
	RunAt       time.Time
	LastError   *string
	CreatedAt   time.Time
}

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		db,
	}
}

type EnqueueJobRequest struct {
	ID      uuid.UUID
	Type    string
	Payload json.RawMessage
}

// CREATE TABLE IF NOT EXISTS jobs (
//     id           UUID PRIMARY KEY,                       -- UUIDv7, generated client-side
//     type         TEXT NOT NULL,
//     payload      JSONB NOT NULL DEFAULT '{}'::jsonb,
//     status       TEXT NOT NULL DEFAULT 'queued'
//                  CHECK (status IN ('queued','running','completed','dead')),
//     attempts     SMALLINT NOT NULL DEFAULT 0,
//     max_retries  SMALLINT NOT NULL DEFAULT 3,
//     leased_until TIMESTAMPTZ,
//     run_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
//     last_error   TEXT,
//     created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
// );

func (h *Handler) handleEnqueueJob(w http.ResponseWriter, r *http.Request) {
	ctype := r.Header.Get("Content-Type")
	if ctype != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var body EnqueueJobRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()

	_, err := h.db.Exec(
		r.Context(),
		`INSERT INTO jobs (id, type, payload) values ($1, $2, $3)`,
		body.ID, body.Type, body.Payload,
	)
	if err != nil {
		slog.Error("db", "insert", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println(body)
}

func handleGetJobs(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func handleJobAck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
