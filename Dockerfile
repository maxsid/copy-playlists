FROM golang:1.16 as build

COPY . .

ENV GOPATH=""
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go get -v -t -d ./...
RUN go build -trimpath -a -o copy-playlists -ldflags="-w -s"

RUN useradd -u 12345 user

FROM scratch
COPY --from=build /go/copy-playlists /copy-playlists
COPY --from=build /etc/passwd /etc/passwd
USER user

ENTRYPOINT ["/copy-playlists"]