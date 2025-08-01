FROM golang:1.24 as builder

WORKDIR /app
COPY . .
RUN go build -o scheduler .

FROM ubuntu:latest

WORKDIR /root
COPY --from=builder /app/scheduler /scheduler
COPY --from=builder /app/web ./web

EXPOSE 7540

ENV TODO_PORT=7540
ENV TODO_DBFILE=./scheduler.db
ENV TODO_PASSWORD=12345

CMD ["/scheduler"]
