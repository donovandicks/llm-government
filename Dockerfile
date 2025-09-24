FROM golang:1.25 AS setup
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM setup AS build
COPY main.go .
COPY pkg/ ./pkg/
COPY internal/ ./internal/
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o ./dist/main ./main.go

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY simulation.yaml .
COPY --from=build /app/dist/main ./main
EXPOSE 9000
ENTRYPOINT ["/app/main", "-from-file", "simulation.yaml"]

