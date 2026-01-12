include .env
export

CONN_STRING=postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable

# DOCKER
.PHONY: build up down logs

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

# migrations
.PHONY: migrate-up migrate-down

migrate-up:
	migrate -path migrations -database "${CONN_STRING}" up

migrate-down:
	migrate -path migrations -database "${CONN_STRING}" down

# psql
.PHONY: psql-conn

psql-conn:
	PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

