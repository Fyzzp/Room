.PHONY: all build build-master build-web clean dev run test

all: build

build: build-web build-master

build-master:
	@echo "Building Room master..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/room ./cmd

build-web:
	@echo "Building frontend..."
	cd web && npm install && npm run build

dev:
	@echo "Starting master in dev mode..."
	DB_HOST=localhost go run ./cmd

test:
	go test ./...

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

clean:
	rm -rf bin/
	rm -rf web/dist/

release:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/room-linux-amd64 ./cmd
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/room-linux-arm64 ./cmd
	cd web && npm install && npm run build
	tar czf bin/room-web-dist.tar.gz -C web dist
	@echo "Release ready in bin/"
