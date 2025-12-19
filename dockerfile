FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /metric-agent ./cmd/agent

FROM scratch
COPY --from=build /metric-agent /metric-agent
EXPOSE 9100
ENTRYPOINT ["/metric-agent"]
