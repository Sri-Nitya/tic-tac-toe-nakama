FROM golang:1.23.5 AS builder

WORKDIR /backend
COPY go.mod go.sum ./

RUN go mod tidy
RUN go mod download

COPY . .

RUN go build --buildmode=plugin -trimpath -o ./backend.so


FROM heroiclabs/nakama:3.26.0

COPY --from=builder /backend/backend.so /nakama/data/modules
