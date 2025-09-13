FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY api/ .
COPY lib/ .
RUN apk add gcc musl-dev
RUN CGO_ENABLED=1 go build -o tanzia-api ./api/.

CMD ["./tanzia-api"]
