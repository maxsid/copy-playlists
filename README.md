# copy-youtube-playlists
A web app that copies several public playlists into one of yours.

## Before start

Firstly you need a JSON credential file, you can create it using the instruction below.

1. You need a Google account with access to Google Cloud Console.
2. [Create a project](https://console.cloud.google.com/projectcreate) and select it. You can give it any name.
3. Enable [YouTube Data API](https://console.cloud.google.com/apis/library/youtube.googleapis.com)
4. Configure [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent/edit).
   Add `https://www.googleapis.com/auth/youtube.force-ssl` scope.
5. Open [API credentials page](https://console.cloud.google.com/apis/credentials) 
   and create credential by clicking `Create Credentials` then `OAuth Client ID`.
7. Select `Web application` type, enter any name and 
   add `Authorized redirect URI` which should be as *http://127.0.0.1:8080/auth*, 
   where *http://127.0.0.1:8080* is a root address of the application.
8. Click `OK` and download *client secret* JSON file of just created credential.

The credential is now created, downloaded and can be used.  

## Build

Build the app by Docker
```shell
docker build -t copy-playlists .
```

Build by Golang compiler
```shell
go build -o copy-playlists .
```

## Usage

Root
```
Usage:
  playlists-copy [command]

Available Commands:
  cli         Run program in CLI mode.
  help        Help about any command
  server      Run web server

Flags:
      --config string       config file (default "/home/user/.config/playlists-copy/config.yaml")
  -c, --credential string   (required) a json credential file from Google Cloud Console
  -h, --help                help for playlists-copy
```

CLI (for single run)
```
Usage:
  playlists-copy cli [flags]

Flags:
  -h, --help   help for cli

Global Flags:
      --config string       config file (default "/home/user/.config/playlists-copy/config.yaml")
  -c, --credential string   (required) a json credential file from Google Cloud Console
```

Server (web server)
```
Usage:
  playlists-copy server [flags]

Flags:
      --addr string   Server listening address (default ":8080")
  -h, --help          help for server

Global Flags:
      --config string       config file (default "/home/user/.config/playlists-copy/config.yaml")
  -c, --credential string   (required) a json credential file from Google Cloud Console
```

## Third-party libraries

* [Cobra](https://github.com/spf13/cobra)
* [Fiber](https://github.com/gofiber/fiber)
* [go-test/deep](https://github.com/go-test/deep)
* [UIkit](https://github.com/uikit/uikit)
* [Google API Go Client](https://github.com/googleapis/google-api-go-client)
