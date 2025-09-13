FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY web ./web
COPY lib ./lib
RUN apk add gcc musl-dev
RUN CGO_ENABLED=1 go build -o tanzia-web ./web/.

CMD ["./tanzia-web"]
