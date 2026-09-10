# Sample image with Go and pdftk-java installed.
# Builds the module and runs the tests, including those that need pdftk.
#
#   docker build -t go-pdftools .
FROM golang:1.26-trixie

RUN apt-get update \
    && apt-get install -y --no-install-recommends pdftk-java \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY . .

RUN go vet ./... && go test ./...
