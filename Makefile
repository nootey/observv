default: http

http:
	go run ./cmd http

migrate:
	go run ./cmd migrate

migrate-fresh:
	go run ./cmd migrate fresh

seed:
	go run ./cmd seed