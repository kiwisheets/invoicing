FROM golang:alpine AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o main .

WORKDIR /dist

RUN cp /build/main .

FROM debian:buster-slim

COPY --from=build /dist/main /

CMD [ "/main" ]
