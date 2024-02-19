FROM golang:1.22

WORKDIR /app

COPY cmd/ cmd/
COPY internal/ internal/

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

RUN go install -v /app/cmd/exporter

CMD ["exporter"]
