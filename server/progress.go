package server

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/youtube/v3"
	"sync"
	"time"
)

// copyingProgress contains information about status of copying.
type copyingProgress struct {
	DestPlaylist *youtube.Playlist  `json:"dest_playlist"`
	Count        int                `json:"current"`
	End          int                `json:"count"`
	Cancel       context.CancelFunc `json:"cancel"`
	Expire       time.Time          `json:"expire"`
}

const expireLong = time.Hour

var (
	progressMap = sync.Map{}
	// anonymous function for unit testing
	timeNow = func() time.Time {
		return time.Now()
	}
)

// getCopyingProgress returns copying progress information for sessionID.
func getCopyingProgress(sessionID string) (*copyingProgress, error) {
	progressInterface, ok := progressMap.Load(sessionID)
	if !ok || progressInterface == nil {
		return nil, fmt.Errorf("%w progress for %s session", ErrNotFound, sessionID)
	}
	progress, ok := progressInterface.(*copyingProgress)
	if !ok {
		return nil, fmt.Errorf("%w of copying progress for %s session", ErrInvalidValue, sessionID)
	}
	return progress, nil
}

// setCopyingProgress sets copying progress information for sessionID.
func setCopyingProgress(sessionID string, progress *copyingProgress) error {
	if sessionID == "" || progress == nil {
		return fmt.Errorf("%w for setting copyingProgress: sessionID=%s, progress=%v",
			ErrInvalidValue, sessionID, progress)
	}
	if progress.Expire == (time.Time{}) {
		progress.Expire = timeNow().Add(expireLong)
	}
	progressMap.Store(sessionID, progress)
	return nil
}

// setCopyingProgressEnd sets a pick value of the progress for sessionID.
func setCopyingProgressEnd(sessionID string, newEnd int) error {
	progress, err := getCopyingProgress(sessionID)
	if err != nil {
		return err
	}
	progress.End = newEnd
	return setCopyingProgress(sessionID, progress)
}

func incrementCopyingProgress(sessionID string, inc int) error {
	progress, err := getCopyingProgress(sessionID)
	if err != nil {
		return err
	}
	progress.Count += inc
	return setCopyingProgress(sessionID, progress)
}

func deleteCopyingProgress(sessionID string) error {
	progress, err := getCopyingProgress(sessionID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil
		}
		return err
	}
	if progress.Cancel != nil {
		progress.Cancel()
	}
	progressMap.Delete(sessionID)
	return nil
}
