FROM golang:alpine as builder

WORKDIR /build

ADD backend/go.mod .

COPY backend .

RUN go build -o auth cmd/auth/main.go

FROM alpine

WORKDIR /build

COPY --from=builder /build/auth /build/auth
COPY backend/config/local.yaml /app/backend/config/local.yaml

ENV JWT_SECRET "super-super-secret"

ENV CONFIG_PATH /app/backend/config/local.yaml

CMD ["./auth"]