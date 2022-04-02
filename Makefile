start-db:
	 docker run --name posgres12 -d \
		-e POSTGRES_PASSWORD=mypassword \
		-e POSTGRES_USER=root \
		-p 5432:5432 \
		postgres:12-alpine

init-db:
	docker exec -it posgres12 createdb --username=root --owner=root simple_bank
drop-db:
	docker exec -it posgres12 dropdb simple_bank

migrate-new:
	migrate create -ext sql -dir db/migrations -seq
migrate-up:
	migrate -path db/migrations -database "postgresql://root:mypassword@localhost:5432/simple_bank?sslmode=disable" --verbose up
migrate-down:
	migrate -path db/migrations -database "postgresql://root:mypassword@localhost:5432/simple_bank?sslmode=disable" --verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./... | sed ''/PASS/s//$$(printf "\033[32mPASS\033[0m")/'' | sed ''/FAIL/s//$$(printf "\033[31mFAIL\033[0m")/''

skaffold:
	make start-db && make init-db && make migrate-up

.PHONY: 
	start-db init-db drop-db migrate-new migrate-up migrate-down sqlc test