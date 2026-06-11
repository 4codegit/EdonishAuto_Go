FROM golang:1.22-bookworm AS builder

RUN apt-get update && apt-get install -y \
    pkg-config \
    libgl1-mesa-dev \
    libxcursor-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxi-dev \
    libxxf86vm-dev \
    libglu1-mesa-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=${VERSION:-dev}" -o edonish-auto .

# ── Runtime ─────────────────────────────────────────────────────
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    libgl1 \
    libxcursor1 \
    libxrandr2 \
    libxinerama1 \
    libxi6 \
    libxxf86vm1 \
    libglu1-mesa \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /build/edonish-auto .

ENTRYPOINT ["./edonish-auto"]
