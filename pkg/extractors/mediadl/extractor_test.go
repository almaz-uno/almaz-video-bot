package mediadl

import (
	"reflect"
	"testing"
)

func Test_FileSizeHumanReadable(t *testing.T) {
	type args struct {
		size int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "less Kb",
			args: args{
				size: 456,
			},
			want: "456b",
		},
		{
			name: "less Mb",
			args: args{
				size: 3*1024 + 78,
			},
			want: "3.1Kb",
		},
		{
			name: "less Gb",
			args: args{
				size: 3*1024*1024 + 600*1024,
			},
			want: "3.6Mb",
		},
		{
			name: "less Tb",
			args: args{
				size: 3*1024*1024*1024 + 345*1024*1024,
			},
			want: "3.3Gb",
		},
		{
			name: "big",
			args: args{
				size: 800*1024*1024*1024*1024 + 345*1024*1024*1024,
			},
			want: "800Tb",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileSizeHumanReadable(tt.args.size); got != tt.want {
				t.Errorf("humanReadableFileSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

const cr = "\r"

func Test_compactWithCR(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "empty",
			args: args{
				src: []byte{},
			},
			want: []byte{},
		},
		{
			name: "the same 3 lines",
			args: args{
				src: []byte(`abc
cde
efg`),
			},
			want: []byte(`abc
cde
efg`),
		},
		{
			name: "with CR 5 lines",
			args: args{
				src: []byte(`abc
eeee` + cr + `aaaaaa` + cr + `cde
aaaaa` + cr + `bbbb` + cr + `efg`),
			},
			want: []byte(`abc
cde
efg`),
		},
		{
			name: "real output",
			args: args{
				src: []byte(`[info] 62d754e20e325c2428b90835: Downloading 1 format(s): 4360-5
[hlsnative] Downloading m3u8 manifest
[hlsnative] Total fragments: 370
[download] Destination: Мир без Америки [62d754e20e325c2428b90835].mp4
` + cr + `[download]  62.1% of ~790.10MiB at   71.79KiB/s ETA 01:11:10 (frag 230/370)` + cr + `[download]  62.1% of ~790.10MiB at  209.70KiB/s ETA 24:21 (frag 230/370)   ` + cr + `[download]  62.1% of ~790.10MiB at  476.49KiB/s ETA 10:43 (frag 230/370)` + cr + `[download]  62.1% of ~790.10MiB at 1000.18KiB/s ETA 05:06 (frag 230/370)`),
			},
			want: []byte(`[info] 62d754e20e325c2428b90835: Downloading 1 format(s): 4360-5
[hlsnative] Downloading m3u8 manifest
[hlsnative] Total fragments: 370
[download] Destination: Мир без Америки [62d754e20e325c2428b90835].mp4
[download]  62.1% of ~790.10MiB at 1000.18KiB/s ETA 05:06 (frag 230/370)`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compactWithCR(tt.args.src); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compactWithCR() = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
