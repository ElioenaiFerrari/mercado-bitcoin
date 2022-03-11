include .env

.PHONY: up
.PHONY: down


up:
	docker-compose up -d --build

down:
	docker-compose down
