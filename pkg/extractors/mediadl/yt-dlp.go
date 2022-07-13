package mediadl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type (

	// very, Very, VERY thanx https://mholt.github.io/json-to-go/!!!
	YtDlpInfo struct {
		ID      string `json:"id,omitempty"`
		Title   string `json:"title,omitempty"`
		Formats []struct {
			FormatID        string      `json:"format_id,omitempty"`
			ManifestURL     string      `json:"manifest_url,omitempty"`
			Ext             string      `json:"ext,omitempty"`
			Width           interface{} `json:"width,omitempty"`
			Height          interface{} `json:"height,omitempty"`
			Tbr             float64     `json:"tbr,omitempty"`
			Asr             int         `json:"asr,omitempty"`
			Fps             interface{} `json:"fps,omitempty"`
			Language        string      `json:"language,omitempty"`
			FormatNote      string      `json:"format_note,omitempty"`
			Filesize        interface{} `json:"filesize,omitempty"`
			Container       string      `json:"container,omitempty"`
			Vcodec          string      `json:"vcodec,omitempty"`
			Acodec          string      `json:"acodec,omitempty"`
			DynamicRange    interface{} `json:"dynamic_range,omitempty"`
			URL             string      `json:"url,omitempty"`
			FragmentBaseURL string      `json:"fragment_base_url,omitempty"`
			Fragments       []struct {
				Path     string  `json:"path,omitempty"`
				Duration float64 `json:"duration,omitempty"`
			} `json:"fragments,omitempty"`
			Protocol             string  `json:"protocol,omitempty"`
			ManifestStreamNumber int     `json:"manifest_stream_number,omitempty"`
			AudioExt             string  `json:"audio_ext,omitempty"`
			VideoExt             string  `json:"video_ext,omitempty"`
			Abr                  float64 `json:"abr,omitempty"`
			Format               string  `json:"format,omitempty"`
			Resolution           string  `json:"resolution,omitempty"`
			HTTPHeaders          struct {
				UserAgent      string `json:"User-Agent,omitempty"`
				Accept         string `json:"Accept,omitempty"`
				AcceptLanguage string `json:"Accept-Language,omitempty"`
				SecFetchMode   string `json:"Sec-Fetch-Mode,omitempty"`
			} `json:"http_headers,omitempty"`
			Vbr         float64     `json:"vbr,omitempty"`
			FormatIndex interface{} `json:"format_index,omitempty"`
			Preference  interface{} `json:"preference,omitempty"`
			Quality     interface{} `json:"quality,omitempty"`
		} `json:"formats,omitempty"`
		ViewCount          interface{} `json:"view_count,omitempty"`
		Uploader           string      `json:"uploader,omitempty"`
		Description        string      `json:"description,omitempty"`
		Thumbnail          string      `json:"thumbnail,omitempty"`
		WebpageURL         string      `json:"webpage_url,omitempty"`
		OriginalURL        string      `json:"original_url,omitempty"`
		WebpageURLBasename string      `json:"webpage_url_basename,omitempty"`
		WebpageURLDomain   string      `json:"webpage_url_domain,omitempty"`
		Extractor          string      `json:"extractor,omitempty"`
		ExtractorKey       string      `json:"extractor_key,omitempty"`
		Playlist           interface{} `json:"playlist,omitempty"`
		PlaylistIndex      interface{} `json:"playlist_index,omitempty"`
		Thumbnails         []struct {
			URL string `json:"url,omitempty"`
			ID  string `json:"id,omitempty"`
		} `json:"thumbnails,omitempty"`
		DisplayID          string      `json:"display_id,omitempty"`
		Fulltitle          string      `json:"fulltitle,omitempty"`
		RequestedSubtitles interface{} `json:"requested_subtitles,omitempty"`
		HasDrm             interface{} `json:"_has_drm,omitempty"`
		RequestedDownloads []struct {
			FormatID     string  `json:"format_id,omitempty"`
			URL          string  `json:"url,omitempty"`
			ManifestURL  string  `json:"manifest_url,omitempty"`
			Tbr          float64 `json:"tbr,omitempty"`
			Ext          string  `json:"ext,omitempty"`
			Fps          float64 `json:"fps,omitempty"`
			Protocol     string  `json:"protocol,omitempty"`
			Width        int     `json:"width,omitempty"`
			Height       int     `json:"height,omitempty"`
			Vcodec       string  `json:"vcodec,omitempty"`
			Acodec       string  `json:"acodec,omitempty"`
			DynamicRange string  `json:"dynamic_range,omitempty"`
			VideoExt     string  `json:"video_ext,omitempty"`
			AudioExt     string  `json:"audio_ext,omitempty"`
			Vbr          float64 `json:"vbr,omitempty"`
			Abr          float64 `json:"abr,omitempty"`
			Format       string  `json:"format,omitempty"`
			Resolution   string  `json:"resolution,omitempty"`
			HTTPHeaders  struct {
				UserAgent      string `json:"User-Agent,omitempty"`
				Accept         string `json:"Accept,omitempty"`
				AcceptLanguage string `json:"Accept-Language,omitempty"`
				SecFetchMode   string `json:"Sec-Fetch-Mode,omitempty"`
			} `json:"http_headers,omitempty"`
			Epoch                int    `json:"epoch,omitempty"`
			Filename             string `json:"_filename,omitempty"`
			WriteDownloadArchive bool   `json:"__write_download_archive,omitempty"`
		} `json:"requested_downloads,omitempty"`
		FormatID     string      `json:"format_id,omitempty"`
		FormatIndex  interface{} `json:"format_index,omitempty"`
		URL          string      `json:"url,omitempty"`
		ManifestURL  string      `json:"manifest_url,omitempty"`
		Tbr          float64     `json:"tbr,omitempty"`
		Ext          string      `json:"ext,omitempty"`
		Fps          float64     `json:"fps,omitempty"`
		Protocol     string      `json:"protocol,omitempty"`
		Preference   interface{} `json:"preference,omitempty"`
		Quality      interface{} `json:"quality,omitempty"`
		Width        int         `json:"width,omitempty"`
		Height       int         `json:"height,omitempty"`
		Vcodec       string      `json:"vcodec,omitempty"`
		Acodec       string      `json:"acodec,omitempty"`
		DynamicRange string      `json:"dynamic_range,omitempty"`
		VideoExt     string      `json:"video_ext,omitempty"`
		AudioExt     string      `json:"audio_ext,omitempty"`
		Vbr          float64     `json:"vbr,omitempty"`
		Abr          float64     `json:"abr,omitempty"`
		Format       string      `json:"format,omitempty"`
		Resolution   string      `json:"resolution,omitempty"`
		HTTPHeaders  struct {
			UserAgent      string `json:"User-Agent,omitempty"`
			Accept         string `json:"Accept,omitempty"`
			AcceptLanguage string `json:"Accept-Language,omitempty"`
			SecFetchMode   string `json:"Sec-Fetch-Mode,omitempty"`
		} `json:"http_headers,omitempty"`
		Epoch int    `json:"epoch,omitempty"`
		Type  string `json:"_type,omitempty"`
	}

	CommandError struct {
		DownloadInfo
		Cause error
	}

	DownloadInfo struct {
		Path   string
		Args   []string
		Stdout string
		Stderr string
	}
)

func (di DownloadInfo) Command() string {
	return di.Path + " " + strings.Join(di.Args, " ")
}

func (e CommandError) Error() string {
	return e.Command() + ": " + e.Cause.Error()
}

func (e CommandError) Unwrap() error {
	return e.Cause
}

func (info *YtDlpInfo) infoFile(format string) string {
	const tmpl = "%s.%s.json"
	if len(info.RequestedDownloads) > 0 {
		return fmt.Sprintf(tmpl, info.RequestedDownloads[0].Filename, format)
	}
	return fmt.Sprintf(tmpl, info.ID, format)
}

// https://ostechnix.com/youtube-dl-tutorial-with-examples-for-beginners/
const (
	ytDlpExec   = "/usr/bin/yt-dlp"
	titleLength = "64"
)

const (
	VideoFormat     = "best[height<=480]/best"
	VideoBestFormat = "best"
	AudioOnlyFormat = "bestaudio"
)

var commonArgs = []string{
	"--no-colors",
	"--no-simulate",
	"--quiet",
	"--dump-single-json",
	"-o",
	"%(title.:" + titleLength + ")s-%(id)s.%(ext)s",
}

func YtDlp(ctx context.Context, dir, format string, args ...string) (*YtDlpInfo, DownloadInfo, error) {
	aa := append(commonArgs, "--format", format)
	aa = append(aa, args...)

	command := exec.CommandContext(ctx, ytDlpExec, aa...)
	command.Dir = dir
	outBuff := &bytes.Buffer{}
	errBuff := &bytes.Buffer{}
	command.Stdout = outBuff
	command.Stderr = errBuff

	err := command.Run()
	di := DownloadInfo{
		Path:   command.Path,
		Args:   command.Args,
		Stdout: outBuff.String(),
		Stderr: errBuff.String(),
	}
	if err != nil {
		commandErr := &CommandError{
			DownloadInfo: di,
			Cause:        err,
		}
		return nil, di, commandErr
	}

	bb := outBuff.Bytes()

	info := new(YtDlpInfo)
	err = json.Unmarshal(bb, info)

	if err == nil {
		_ = os.WriteFile(filepath.Join(dir, info.infoFile(info.FormatID)), []byte(di.Stdout), 0o644)
		_ = os.WriteFile(filepath.Join(dir, info.infoFile(info.FormatID+".err")), []byte(di.Stderr), 0o644)
	}

	return info, di, err
}
