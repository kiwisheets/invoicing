FROM golang:alpine AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o main .

FROM debian:buster-slim

COPY mq.yaml ./

COPY --from=build /build/main /

CMD [ "/main" ]
