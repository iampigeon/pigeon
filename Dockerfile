# Stage 1: compile inside throwaway container
# FROM golang:latest as gobuild

# COPY /bin/scheduler /go/src/app
# WORKDIR /go/src/app

# RUN apk --no-cache add -U git
# RUN go get -u github.com/golang/dep/cmd/dep && dep ensure
# RUN CGO_ENABLED=0 GOOS=linux go build -o main /go/src/app/cmd/scheduler

# # Stage 2: insise of docker redis
# FROM redis:alpine
# WORKDIR /app
# COPY --from=gobuild /go/src/app/main /app

# # RUN make build .
# CMD ["redis-server"]
# CMD ["./main"]

FROM alpine

COPY /bin/scheduler /
COPY config.yml /
COPY supervisord.conf /
WORKDIR /

RUN mkdir /var/log/supervisor
RUN touch /var/log/supervisor/redis.log
RUN touch /var/log/supervisor/scheduler.log

RUN apk --no-cache add -U supervisor redis

COPY redis.conf /

CMD ["/usr/bin/supervisord", "-c", "/supervisord.conf", "-n"]
# CMD ["redis-server", "redis.conf"]
# CMD ["sleep", "120"]
# CMD ["./scheduler"]
# RUN ["./scheduler"]
