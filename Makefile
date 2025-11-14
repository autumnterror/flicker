test-jwt:
	go test ./internal/auth/jwt -v
test-psql:
	go test ./internal/auth/psql -v

compose-db-up:
	docker compose -f ./build/deploy/docker-compose.db.yml up -d
compose-db-down:
	docker compose -f ./build/deploy/docker-compose.db.yml down
compose-up:
	docker compose -f ./build/deploy/docker-compose.yml up -d
compose-down:
	docker compose -f ./build/deploy/docker-compose.yml down

docker-build: 
	docker build --platform linux/amd64 -t zitrax78/flicker --file ./build/deploy/dockerfile .


docx:
	swag init --dir ./cmd/flicker,./internal/net,./internal/views --output ./docs
