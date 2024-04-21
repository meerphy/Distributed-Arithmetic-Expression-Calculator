FROM golang:1.22.1-alpine3.19

WORKDIR /app

COPY go.mod go.sum /

RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

CMD [ "./main" ]