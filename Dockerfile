FROM golang:1.20 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o admission-controller .

FROM alpine:latest
COPY --from=build /app/admission-controller /admission-controller
ENTRYPOINT ["/admission-controller"]
