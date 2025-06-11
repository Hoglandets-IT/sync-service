FROM golang:1.24.4-alpine

WORKDIR /app

COPY ./sync-service/go.mod ./
COPY ./sync-service/go.sum ./
RUN go mod download

COPY ./sync-service/* ./

RUN go build -o ./sync-service

EXPOSE 8080

CMD [ "./sync-service" ]
