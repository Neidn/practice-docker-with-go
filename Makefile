#POSTGRES_URL=postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable
POSTGRES_URL=postgresql://root:secret@neidn.com:5432/simple_bank?sslmode=disable

migrateup:
	migrate -path db/migration -database "$(POSTGRES_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(POSTGRES_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(POSTGRES_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(POSTGRES_URL)" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go practice-docker/db/sqlc Store

.PHONY: migratedown migratedown1 migrateup migrateup1 sqlc test server mock
