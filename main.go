package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	env, _ := godotenv.Read()

	db, err := pgxpool.New(context.Background(), env["PG_CONNECTION_STRING"])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not establish connection to db: %v", err)
		os.Exit(1)
	}

	jobHandler := NewHandler(db)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /jobs", jobHandler.handleEnqueueJob)
	mux.HandleFunc("GET /jobs", jobHandler.handleListJobs)
	mux.HandleFunc("POST /jobs/{id}/ack", handleJobAck)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      35 * time.Second, // > worker long-poll duration
		IdleTimeout:       120 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
