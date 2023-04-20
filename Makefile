postgres:
	docker run -p 5432:5432 --name postgres -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123 -d postgres

createdb:
	docker exec -it postgres createdb --username=root --owner=root bank

dropdb:
	docker exec -it postgres dropdb --username=root bank

migrateup:
	migrate -path db/migration -database "postgresql://root:123@localhost:5432/bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:123@localhost:5432/bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover  ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination api/mock/store.go github.com/homocode/bank_demo/api Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock