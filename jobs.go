package main

import (
	"encoding/json"
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
	ID          uuid.UUID       `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Status      JobStatus       `json:"status"`
	Attempts    int             `json:"attempts"`
	MaxRetries  int             `json:"maxRetries"`
	LeasedUntil *time.Time      `json:"leasedUntil"`
	RunAt       time.Time       `json:"runAt"`
	LastError   *string         `json:"lastError"`
	CreatedAt   time.Time       `json:"createdAt"`
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
}

func (h *Handler) handleListJobs(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT * FROM jobs
	`
	rows, err := h.db.Query(
		r.Context(),
		query,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var jobs []Job

	for rows.Next() {
		var j Job
		err := rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &j.MaxRetries, &j.LeasedUntil, &j.RunAt, &j.LastError, &j.CreatedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jobs = append(jobs, j)
	}

	if err = json.NewEncoder(w).Encode(jobs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

func handleJobAck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
