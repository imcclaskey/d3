package ports

import (
	"io/fs"
	"os"
	"path/filepath"
)

// FileSystem defines an interface for file system operations.
// This allows for abstraction of the os package for testing or other purposes.
//
//go:generate mockgen -destination=mocks/mock_filesystem.go -package=mocks github.com/imcclaskey/d3/internal/core/ports FileSystem
type FileSystem interface {
	Stat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	ReadDir(name string) ([]fs.DirEntry, error)
	Create(name string) (*os.File, error)
	Remove(name string) error
	RemoveAll(path string) error
	Exists(name string) (bool, error)
	Rename(oldpath, newpath string) error
	Glob(pattern string) ([]string, error)
}

// RealFileSystem is a concrete implementation of FileSystem using the os package.
type RealFileSystem struct{}

// Stat retrieves file information.
func (rfs RealFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// ReadFile reads the entire content of a file.
func (rfs RealFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile writes data to a file.
func (rfs RealFileSystem) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// MkdirAll creates a directory path along with any necessary parents.
func (rfs RealFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// ReadDir reads the directory named by dirname and returns a list of directory entries.
func (rfs RealFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

// Create creates or truncates the named file.
func (rfs RealFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

// Remove removes the named file or (empty) directory.
func (rfs RealFileSystem) Remove(name string) error {
	return os.Remove(name)
}

// RemoveAll removes path and any children it contains.
func (rfs RealFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Exists checks if a file or directory exists.
func (rfs RealFileSystem) Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Rename calls os.Rename.
func (rfs RealFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Glob calls filepath.Glob.
func (rfs RealFileSystem) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}
