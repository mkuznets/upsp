FROM golang:1.19.3-alpine3.16 as build
WORKDIR /build
# Cache go modules
ADD go.sum go.mod /build/
RUN go mod download
ENV CGO_ENABLED=0
ADD . /build
RUN go build -ldflags="-s -w" -o upsp mkuznets.com/go/upsp/cmd/upsp

FROM scratch as upsp
COPY --from=build /build/upsp /upsp

