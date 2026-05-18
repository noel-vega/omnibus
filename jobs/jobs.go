package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	db         *pgxpool.Pool
	repository *Repository
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{
		db:         db,
		repository: NewRepository(db),
	}
}

type EnqueueJobRequest struct {
	ID      uuid.UUID
	Type    string
	Payload json.RawMessage
}

func (h *Handler) HandleEnqueueJob(w http.ResponseWriter, r *http.Request) {
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

	if err := h.repository.InsertJob(r.Context(), InsertJobParams(body)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.repository.ListJobs(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(jobs); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	job, err := h.repository.GetJob(r.Context(), jobID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(job); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type DequeueJobRequest struct {
	Type string
}

func (h *Handler) HandleDequeueJob(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Dequeue Job...")
	var data DequeueJobRequest

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := `
	UPDATE jobs
	SET status = 'running'
	WHERE id = (
		SELECT id FROM jobs
		WHERE type = $1 AND status = 'queued' 
		ORDER BY run_at 
	  LIMIT 1
	  FOR UPDATE SKIP LOCKED
	)
	RETURNING *;
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var j Job
	for {
		row := h.db.QueryRow(ctx, query, data.Type)

		if err := row.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.Attempts, &j.MaxRetries, &j.LeasedUntil, &j.RunAt, &j.LastError, &j.CreatedAt); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				slog.Error("sql", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			select {
			case <-ticker.C:
				fmt.Println("Tick")
				continue
			case <-ctx.Done():
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		break
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(j); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// func HandleJobAck(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusNotImplemented)
// }
