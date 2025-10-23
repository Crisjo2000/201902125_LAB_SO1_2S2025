FROM rust:1.83 as builder

WORKDIR /usr/src/app

COPY Cargo.toml Cargo.lock ./
COPY src ./src

RUN cargo build --release

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /usr/src/app/target/release/rust-api /app/rust-api

ENV PORT=8000
ENV GO_SERVICE_URL=http://go-grpc-client:8080/tweet

EXPOSE 8000

CMD ["./rust-api"]