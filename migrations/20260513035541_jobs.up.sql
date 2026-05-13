CREATE TABLE IF NOT EXISTS jobs (
    id           UUID PRIMARY KEY,                       -- UUIDv7, generated client-side
    type         TEXT NOT NULL,
    payload      JSONB NOT NULL DEFAULT '{}'::jsonb,
    status       TEXT NOT NULL DEFAULT 'queued'
                 CHECK (status IN ('queued','running','completed','dead')),
    attempts     SMALLINT NOT NULL DEFAULT 0,
    max_retries  SMALLINT NOT NULL DEFAULT 3,
    leased_until TIMESTAMPTZ,
    run_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_dequeue
    ON jobs (type, run_at ASC)
    WHERE status = 'queued';

CREATE INDEX idx_jobs_lease_expiry
    ON jobs (leased_until)
    WHERE status = 'running';

CREATE INDEX idx_jobs_status_created
    ON jobs (status, created_at DESC);
