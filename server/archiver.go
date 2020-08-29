package server

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/mholt/archiver/v3"
	"github.com/pterodactyl/wings/config"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Archiver represents a Server Archiver.
type Archiver struct {
	Server *Server
}

// ArchivePath returns the path to the server's archive.
func (a *Archiver) ArchivePath() string {
	return filepath.Join(config.Get().System.ArchiveDirectory, a.ArchiveName())
}

// ArchiveName returns the name of the server's archive.
func (a *Archiver) ArchiveName() string {
	return a.Server.Id() + ".tar.gz"
}

// Exists returns a boolean based off if the archive exists.
func (a *Archiver) Exists() bool {
	if _, err := os.Stat(a.ArchivePath()); os.IsNotExist(err) {
		return false
	}

	return true
}

// Stat stats the archive file.
func (a *Archiver) Stat() (*Stat, error) {
	return a.Server.Filesystem.unsafeStat(a.ArchivePath())
}

// Archive creates an archive of the server and deletes the previous one.
func (a *Archiver) Archive() error {
	path := a.Server.Filesystem.Path()

	// Get the list of root files and directories to archive.
	var files []string
	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range fileInfo {
		f, err := a.Server.Filesystem.SafeJoin(path, file)
		if err != nil {
			return err
		}

		files = append(files, f)
	}

	stat, err := a.Stat()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if the file exists.
	if stat != nil {
		if err := os.Remove(a.ArchivePath()); err != nil {
			return err
		}
	}

	return archiver.NewTarGz().Archive(files, a.ArchivePath())
}

// DeleteIfExists deletes the archive if it exists.
func (a *Archiver) DeleteIfExists() error {
	stat, err := a.Stat()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if the file exists.
	if stat != nil {
		if err := os.Remove(a.ArchivePath()); err != nil {
			return err
		}
	}

	return nil
}

// Checksum computes a SHA256 checksum of the server's archive.
func (a *Archiver) Checksum() (string, error) {
	file, err := os.Open(a.ArchivePath())
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	buf := make([]byte, 1024*4)
	if _, err := io.CopyBuffer(hash, file, buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
