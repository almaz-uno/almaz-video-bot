package loghook

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/rs/zerolog"
)

type GoroutineStack struct {
	GIDName   string
	StackFile string
	StackPath string
}

func (gid *GoroutineStack) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	stack := debug.Stack()

	b := bytes.TrimPrefix(stack, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]

	e.Str(gid.GIDName, string(b))

	// it saves stacktrace in file and output the
	if gid.StackPath != "" && os.MkdirAll(gid.StackPath, 0o755) == nil {
		stack = stack[bytes.IndexByte(stack, '\n')+1:]
		sum := sha512.Sum512(stack)
		stackID := hex.EncodeToString(sum[:])
		stackFile := filepath.Join(gid.StackPath, stackID)

		if _, e1 := os.Stat(stackFile); errors.Is(e1, os.ErrNotExist) {
			if e2 := os.WriteFile(stackFile, stack, 0o644); e2 != nil {
				e.Str("err-stack-write", e2.Error())
			}
		}
		e.Str(gid.StackFile, stackFile)
	}
}
