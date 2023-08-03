FROM golang:1.20 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o admission-controller .

FROM cgr.dev/chainguard/wolfi-base
# set a label so we can easily identify this image
LABEL base_image="cgr.dev/chainguard/wolfi-base"
COPY --from=build /app/admission-controller /admission-controller
ENTRYPOINT ["/admission-controller"]