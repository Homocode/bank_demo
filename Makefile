postgres:
	docker run -p 5432:5432 --name postgresInstance1 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=123 -d postgres

createdb:
	docker exec -it postgresInstance1 createdb --username=root --owner=root bank

dropdb:
	docker exec -it postgresInstance1 dropdb --username=root bank

dbcli:
	docker exec -it postgresInstance1 psql

migrateup:
	migrate -path db/migration -database "postgresql://root:123@localhost:5432/bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:123@localhost:5432/bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

.PHONY: postgres createdb dropdb dbcli migrateup migratedown sqlc