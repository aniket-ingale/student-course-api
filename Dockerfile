# syntax=docker/dockerfile:1

# --- build stage ---
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Static build so it runs on a minimal base image.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# --- runtime stage ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server

# Run as the non-root user provided by the distroless image.
USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/app/server"]
