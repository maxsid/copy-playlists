package cli

import "github.com/maxsid/playlists-copy/youtube"

type configCodeRequester interface {
	youtube.ConfigCodeExchanger
	youtube.ConfigCodeURLGenerator
}
