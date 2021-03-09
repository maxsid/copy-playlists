package helper

import "testing"

func TestYoutubePlaylistIDFromURL(t *testing.T) {
	type args struct {
		rawurl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Sample 1",
			args: args{rawurl: "https://www.youtube.com/playlist?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-"},
			want: "PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-",
		},
		{
			name: "Trim sample",
			args: args{rawurl: "   https://www.youtube.com/playlist?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-   "},
			want: "PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-",
		},
		{
			name: "Sample 2",
			args: args{rawurl: "   https://youtube.com/playlist?list=PLTVdmvDFrwPMhfnZPXCdUkvy_EH3GnFa-   "},
			want: "PLTVdmvDFrwPMhfnZPXCdUkvy_EH3GnFa-",
		},
		{
			name:    "Not found 1",
			args:    args{rawurl: "https://www.youtube.com/playlist"},
			wantErr: true,
		},
		{
			name:    "Another domain",
			args:    args{rawurl: "https://ru.wikipedia.org/wiki?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-"},
			wantErr: true,
		},
		{
			name:    "Edge 1",
			args:    args{rawurl: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := YoutubePlaylistIDFromURL(tt.args.rawurl)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoutubePlaylistIDFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("YoutubePlaylistIDFromURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYoutubePlaylistIDFromURLBadSymbols(t *testing.T) {
	badSymbols := []rune{0x00, '\n', '\r', '\n', '<', '(', ')', '=', '}', '|', '*', ' '}
	goodUrl := []rune("https://www.youtube.com/playlist?list=PLTVdmvDFrwPMhfnZPXCdUkvy-EH3GnFa-")
	replaceIndex := 50
	for _, bs := range badSymbols {
		goodUrl[replaceIndex] = bs
		if _, err := YoutubePlaylistIDFromURL(string(goodUrl)); err == nil {
			t.Errorf("YoutubePlaylistIDFromURL() not returns error with '%c' bad symbol", bs)
		}
	}
}
