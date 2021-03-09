package helper

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const regexpPlaylistLinkPattern = `^https:\/\/(?:www\.)?youtube.com\/playlist\?list\=([a-zA-Z0-9\-_]+$)`

var ErrInvalidURL = errors.New("invalid url")

func YoutubePlaylistIDFromURL(rawurl string) (string, error) {
	rawurl = strings.TrimSpace(rawurl)
	comp, err := regexp.Compile(regexpPlaylistLinkPattern)
	if err != nil {
		return "", err
	}
	matches := comp.FindStringSubmatch(rawurl)
	if len(matches) < 2 {
		return "", fmt.Errorf("%w of the playlist: \"%s\" is not a link to playlist", ErrInvalidURL, rawurl)
	}
	return matches[1], nil
}
