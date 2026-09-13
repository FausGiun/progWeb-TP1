.PHONY: test

test:
	docker compose down -v
	sqlc generate
	go build ./...
	docker compose up -d
	@until [ "$$(docker inspect --format='{{.State.Health.Status}}' postgres-db)" = "healthy" ]; do sleep 1; done
	go test -v ./...
	docker compose down -v