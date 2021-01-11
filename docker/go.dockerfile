FROM golang:1.15 as builder
LABEL maintainer="Daniel Ramirez"

RUN apt update && apt install -y build-essential automake libevent-dev libssl-dev zlib1g-dev

RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN GOOS=linux go build -a -o main .

FROM gcr.io/distroless/base-debian10 as prod
COPY --from=builder /build/main /app/
EXPOSE 8080
ENTRYPOINT [ "/app/main" ]
