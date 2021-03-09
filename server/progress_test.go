package server

import (
	"github.com/go-test/deep"
	"google.golang.org/api/youtube/v3"
	"sync"
	"testing"
	"time"
)

func Test_setCopyingProgress(t *testing.T) {
	wasTimeNow := timeNow
	timeNow = func() time.Time {
		return time.Unix(0, 0)
	}
	defer func() {
		progressMap = sync.Map{}
		timeNow = wasTimeNow
	}()

	type args struct {
		sessionID string
		progress  *copyingProgress
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				sessionID: "123456789",
				progress: &copyingProgress{
					End:          10,
					Count:        0,
					DestPlaylist: &youtube.Playlist{Id: "123456"},
				},
			},
			want: &copyingProgress{
				End:          10,
				Count:        0,
				Expire:       time.Unix(0, 0).Add(expireLong),
				DestPlaylist: &youtube.Playlist{Id: "123456"},
			},
		},
		{
			name: "OK without Expire changing",
			args: args{
				sessionID: "123456789",
				progress: &copyingProgress{
					End:          10,
					Count:        0,
					Expire:       time.Unix(1000, 1000),
					DestPlaylist: &youtube.Playlist{Id: "123456"},
				},
			},
			want: &copyingProgress{
				End:          10,
				Count:        0,
				Expire:       time.Unix(1000, 1000),
				DestPlaylist: &youtube.Playlist{Id: "123456"},
			},
		},
		{
			name: "Empty session ID",
			args: args{
				sessionID: "",
				progress:  &copyingProgress{End: 10, Count: 0},
			},
			wantErr: true,
		},
		{
			name: "Empty progress params",
			args: args{
				sessionID: "123456789",
				progress:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setCopyingProgress(tt.args.sessionID, tt.args.progress); (err != nil) != tt.wantErr {
				t.Errorf("setCopyingProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got, _ := progressMap.Load(tt.args.sessionID)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_setCopyingProgressEnd(t *testing.T) {
	const sessionID = "123456789"
	defer func() {
		progressMap = sync.Map{}
	}()
	progressMap.Store(sessionID, &copyingProgress{End: 10, Expire: time.Unix(0, 0)})

	type args struct {
		sessionID string
		newEnd    int
	}
	tests := []struct {
		name    string
		args    args
		want    *copyingProgress
		wantErr bool
	}{
		{
			name: "Sample",
			args: args{sessionID: sessionID, newEnd: 20},
			want: &copyingProgress{
				End:    20,
				Expire: time.Unix(0, 0),
			},
		},
		{
			name:    "Not found error",
			args:    args{sessionID: "213f", newEnd: 20},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setCopyingProgressEnd(tt.args.sessionID, tt.args.newEnd); (err != nil) != tt.wantErr {
				t.Errorf("setCopyingProgressEnd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got, _ := progressMap.Load(tt.args.sessionID)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_incrementCopyingProgress(t *testing.T) {
	const sessionID = "123456789"
	defer func() {
		progressMap = sync.Map{}
	}()
	progressMap.Store(sessionID, &copyingProgress{End: 10, Expire: time.Unix(0, 0)})

	type args struct {
		sessionID string
		inc       int
	}
	tests := []struct {
		name    string
		args    args
		want    *copyingProgress
		wantErr bool
	}{
		{
			name: "Sample",
			args: args{sessionID: sessionID, inc: 22},
			want: &copyingProgress{End: 10, Count: 22, Expire: time.Unix(0, 0)},
		},
		{
			name:    "Not found",
			args:    args{sessionID: "dsaasga", inc: 22},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := incrementCopyingProgress(tt.args.sessionID, tt.args.inc); (err != nil) != tt.wantErr {
				t.Errorf("incrementCopyingProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got, _ := progressMap.Load(tt.args.sessionID)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_getCopyingProgress(t *testing.T) {
	const (
		sessionID          = "123456789"
		sessionIDNil       = "6543123"
		sessionIDWrongType = "dd6543123"
	)
	defer func() {
		progressMap = sync.Map{}
	}()
	progressMap.Store(sessionID, &copyingProgress{End: 10, Expire: time.Unix(0, 0)})
	progressMap.Store(sessionIDNil, nil)
	progressMap.Store(sessionIDWrongType, 4215125)

	type args struct {
		sessionID string
	}
	tests := []struct {
		name    string
		args    args
		want    *copyingProgress
		wantErr bool
	}{
		{
			name: "OK",
			args: args{sessionID: sessionID},
			want: &copyingProgress{End: 10, Expire: time.Unix(0, 0)},
		},
		{
			name:    "Not found",
			args:    args{sessionID: "dsafsagasgas"},
			wantErr: true,
		},
		{
			name:    "Nil value",
			args:    args{sessionID: sessionIDNil},
			wantErr: true,
		},
		{
			name:    "Wrong type",
			args:    args{sessionID: sessionIDWrongType},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCopyingProgress(tt.args.sessionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCopyingProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_deleteCopyingProgress(t *testing.T) {
	const (
		sessionID1 = "1234567891"
		sessionID2 = "1234567892"
		sessionID3 = "1234567893"
	)
	cancelCount := 0
	cancel := func() {
		cancelCount++
	}
	defer func() {
		progressMap = sync.Map{}
	}()
	progressMap.Store(sessionID1, &copyingProgress{End: 10, Expire: time.Unix(0, 0), Cancel: cancel})
	progressMap.Store(sessionID2, &copyingProgress{End: 10, Expire: time.Unix(0, 0)})
	progressMap.Store(sessionID3, 231)

	type args struct {
		sessionID string
	}
	tests := []struct {
		name            string
		args            args
		wantCancelCount int
		wantErr         bool
	}{
		{
			name:            "OK",
			args:            args{sessionID: sessionID1},
			wantCancelCount: 1,
		},
		{
			name:            "OK without cancel",
			args:            args{sessionID: sessionID2},
			wantCancelCount: 0,
		},
		{
			name:            "Not found without error",
			args:            args{sessionID: "sdafv"},
			wantCancelCount: 0,
		},
		{
			name:    "Incorrect type",
			args:    args{sessionID: sessionID3},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cancelCount = 0
			if err := deleteCopyingProgress(tt.args.sessionID); (err != nil) != tt.wantErr {
				t.Errorf("deleteCopyingProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if cancelCount != tt.wantCancelCount {
				t.Errorf("cancelCount = %v, want %v", cancelCount, tt.wantCancelCount)
				return
			}
			if _, ok := progressMap.Load(tt.args.sessionID); ok {
				t.Errorf("sessionID %v hasn't deleted", tt.args.sessionID)
			}
		})
	}
}
