FROM golang:1.21-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go get -u github.com/pressly/goose/cmd/goose

COPY . .

COPY .env /app/

RUN bash -c 'source /app/.env'

RUN go build cmd/daee/main.go

COPY sql/schema /app/sql/schema

RUN cd /app/sql/schema && goose postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB}" up

EXPOSE 3000

CMD ["./main"]