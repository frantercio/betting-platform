FROM golang:1.23 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM scratch
COPY --from=build /bin/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
