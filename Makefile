.PHONY: run test lint docker-up docker-down clean

run:
	docker-compose up -d postgres
	air

test:
	go test -v ./...

lint:
	golangci-lint run

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down -v

clean:
	go clean -cache