APP_NAME=pr-reviewer-assignment-service

.PHONY: run build test lint docker up down logs e2e test-unit test-e2e test-integration

## Запуск приложения
run:
	go run ./cmd/main.go

## Сборка бинарника
build:
	go build -o bin/$(APP_NAME) ./cmd/main.go

## Тесты:
test-unit:
	go test ./internal/... -count=1

test-e2e:
	go test ./test/e2e -count=1 -v

test-integration:
	go test ./test/integration -count=1 -v

## Все тесты одной командой.
test: test-unit test-e2e test-integration

## Docker image
docker:
	docker build -t $(APP_NAME):latest .

## docker compose up/down
up:
	docker-compose up -d

up-build:
	docker-compose up --build -d

down:
	docker-compose down

## Сбрасывает volume (для сброса БД)
down-v:
	docker-compose down -v

logs:
	docker-compose logs -f $(APP_NAME)
