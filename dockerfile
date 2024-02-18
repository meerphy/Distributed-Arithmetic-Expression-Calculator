FROM golang:1.22

ENV GOPATH=/

COPY ./ ./

RUN go mod download
RUN go build -o microservice ./cmd/main.go

CMD [ "./microservice" ]
