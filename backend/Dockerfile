FROM golang:1.21.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go get -u github.com/pressly/goose/cmd/goose

COPY . .

RUN go build -o daee .

FROM debian:buster

COPY --from=builder /app/daee /usr/local/bin/daee

EXPOSE 3000

COPY .env /

RUN bash -c 'source /.env'

COPY sql/schema /app/sql/schema

RUN cd /app/sql/schema && goose postgres postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${POSTGRES_DB} up

CMD ["daee"]
