FROM golang:1.22
ENV GOPATH=/
WORKDIR /microservice
COPY go.mod go.sum /
RUN go mod download
COPY . .
RUN go build -o microservice ./cmd/main.go
CMD ["./microservice"]