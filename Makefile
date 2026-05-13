include .env
export

migrate-up:
	migrate -database "postgres://$$PG_USER:$$PG_PASSWORD@$$PG_HOST:5432/$$PG_DB?sslmode=disable" -path migrations up

migrate-drop:
	migrate -database "postgres://$$PG_USER:$$PG_PASSWORD@$$PG_HOST:5432/$$PG_DB?sslmode=disable" -path migrations drop 

db-up:
	docker compose up db -d
