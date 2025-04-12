FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk add gcc musl-dev
RUN CGO_ENABLED=1 go build .

CMD ["./tantieme"]
