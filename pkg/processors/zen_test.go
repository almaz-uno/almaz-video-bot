package processors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_extractM3U8(t *testing.T) {
	url := "https://zen.yandex.ru/video/watch/62c339096ed093503d29c2b8"
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	m3u8s, title, err := extractM3U8(ctx, url)
	require.NoError(t, err)

	assert.Equal(t, "ДРОБНИЦКИЙ | №46: Каким будет миропорядок Запада?", title)
	assert.NotEmpty(t, m3u8s)
}
