# Makefile for Database Migration

SHELL := /bin/bash
ENV_FILE := .env
MIGRATIONS_DIR := database/migrations
DATABASE_URL :=  $(shell grep '^DATABASE_URL=' $(ENV_FILE) | cut -d '=' -f 2)

.PHONY:	create migrate rollback

create:
	@echo "creating new migration setup..."
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq events
    
migrate:
	@echo "Applying database migrations..."
	migrate -database $(DATABASE_URL) -path $(MIGRATIONS_DIR) up

rollback:
	@echo "Rolling back the last database migration..."
	migrate -database $(DATABASE_URL) -path $(MIGRATIONS_DIR) down