# Stage 1: compile inside throwaway container
# FROM golang:latest as gobuild

# COPY /bin/scheduler /go/src/app
# WORKDIR /go/src/app

# RUN apk --no-cache add -U git
# RUN go get -u github.com/golang/dep/cmd/dep && dep ensure
# RUN CGO_ENABLED=0 GOOS=linux go build -o main /go/src/app/cmd/scheduler

# CMD ["./main"]

FROM golang:alpine

COPY /bin/scheduler /
COPY config.yml /
WORKDIR /
ENV REDIS_URL redis:6379
EXPOSE 9000
EXPOSE 9001

CMD ["./scheduler"]
