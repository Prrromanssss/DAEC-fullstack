run-rabbitmq:
	docker run -d --name my-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

run-postgres:
	docker run --name habr-pg-13.3 -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=daee -d postgres:13.3

generate:
	cd ./backend/internal/protos && protoc -I proto proto/daee/daee.proto \ 
	--go_out=./gen/go --go_opt=paths=source_relative \ 
	--go-grpc_out=./gen/go --go-grpc_opt=paths=source_relative