package mediadl

import "testing"

func Test_humanReadableFileSize(t *testing.T) {
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
			if got := humanReadableFileSize(tt.args.size); got != tt.want {
				t.Errorf("humanReadableFileSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
