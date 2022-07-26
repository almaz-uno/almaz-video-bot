package main

import (
	"io/fs"
	"syscall"
	"time"

	"almaz.uno/dev/almaz-video-bot/pkg/extractors/mediadl"
	"github.com/rs/zerolog/log"
)

type fileInfo struct {
	d        fs.DirEntry
	Path     string
	URL      string
	LinksURL string
}

const timeFormat = "2006-01-02 15:04:05Z07:00"

func (fi fileInfo) Name() string {
	return fi.d.Name()
}

func (fi fileInfo) SizeStr() string {
	return mediadl.FileSizeHumanReadable(fi.Size())
}

func (fi fileInfo) ATimeAgo() time.Duration {
	return time.Duration(time.Since(fi.ATime()).Seconds()) * time.Second
}

func (fi fileInfo) CTimeAgo() time.Duration {
	return time.Duration(time.Since(fi.CTime()).Seconds()) * time.Second
}

func (fi fileInfo) MTimeAgo() time.Duration {
	return time.Duration(time.Since(fi.MTime()).Seconds()) * time.Second
}

func (fi fileInfo) ATimeStr() string {
	return fi.ATime().Format(timeFormat)
}

func (fi fileInfo) CTimeStr() string {
	return fi.CTime().Format(timeFormat)
}

func (fi fileInfo) MTimeStr() string {
	return fi.MTime().Format(timeFormat)
}

func (fi fileInfo) MTime() time.Time {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return time.Time{}
	}
	return i.ModTime()
}

func (fi fileInfo) ATime() time.Time {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return time.Time{}
	}
	stat := i.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
}

func (fi fileInfo) CTime() time.Time {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return time.Time{}
	}
	stat := i.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
}

func (fi fileInfo) Size() int64 {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return 0
	}
	return i.Size()
}
