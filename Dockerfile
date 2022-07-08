FROM golang:1.17.5-alpine AS build

WORKDIR /app

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o /out .


FROM alpine

RUN apk --no-cache add ca-certificates

COPY --from=build /out .

COPY webserver.conf .

CMD ["./out", "server"]





