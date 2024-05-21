FROM golang:1.19 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /lazyop

FROM gcr.io/distroless/base-debian11 AS package

WORKDIR /

COPY --from=build /lazyop /lazyop

USER nonroot:nonroot

ENTRYPOINT ["/lazyop"]