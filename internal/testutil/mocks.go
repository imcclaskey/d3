package testutil

import (
	"os"
	"time"
)

// MockFileInfo is a minimal implementation of os.FileInfo for shared test usage.
type MockFileInfo struct {
	FName    string
	FIsDir   bool
	FSize    int64
	FMode    os.FileMode
	FModTime time.Time
}

func (mfi MockFileInfo) Name() string       { return mfi.FName }
func (mfi MockFileInfo) Size() int64        { return mfi.FSize }
func (mfi MockFileInfo) Mode() os.FileMode  { return mfi.FMode }
func (mfi MockFileInfo) ModTime() time.Time { return mfi.FModTime }
func (mfi MockFileInfo) IsDir() bool        { return mfi.FIsDir }
func (mfi MockFileInfo) Sys() interface{}   { return nil }
