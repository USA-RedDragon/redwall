FROM golang:1.20.1 as build-env

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o redwall ./cmd/redwall/main.go

FROM scratch
COPY --from=build-env /src/redwall /redwall
ENTRYPOINT ["/redwall"]
