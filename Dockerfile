FROM golang:1.16-alpine as builder

COPY . .

ENV GOPATH=""
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go get -v -t -d ./...
RUN go build -trimpath -a -o copy-playlists -ldflags="-w -s"

ENV USER user
ENV UID 12345

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

FROM scratch
COPY --from=builder /go/copy-playlists /copy-playlists
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER user
EXPOSE 8080

ENTRYPOINT ["/copy-playlists"]