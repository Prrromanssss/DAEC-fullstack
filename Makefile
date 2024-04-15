pull-rabbitmq:
	docker pull rabbitmq:3-management

run-rabbitmq:
	docker run -d --name my-rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

run-postgres:
	docker run --name habr-pg-13.3 -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=daee -d postgres:13.3