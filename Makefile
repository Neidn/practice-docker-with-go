POSTGRES_URL=postgresql://root@neidn.com:5432/simple_bank?sslmode=disable&password=secret

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

.PHONY: migratedown migratedown1 migrateup migrateup1 sqlc
