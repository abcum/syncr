// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logr

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

var (
	// ErrSeekPositionNotSupported occurs when a syncable type receives a
	// seek destination which it can not fulfil to. These syncable types
	// usually only support seeking to the beginning or to the end, as they
	// write to and read from multiple data files.
	ErrSeekPositionNotSupported = errors.New("Seek position is not supported.")
)

// Options represents syncing configuration options.
type Options struct {
	// MaxSize is used to determine the maximum size that a file can be
	// before the file should be rotated, and a new file used. This is
	// only available on certain types which support rotation of files,
	// and appendable writes. MaxSize should be specified in megabytes.
	MaxSize int
}

// Storage represents a rotating log-file reader and writer.
type Storage struct {
	opts *Options
	pntr *os.File
	lock sync.Mutex
	size struct {
		max int // max file size in bytes
		now int // current file size in bytes
	}
	file struct {
		seek int // are we at beginning (0) or end (2)
		full string
		path string
		base string
		name string
		extn string
	}
}

// New creates a new Syncable storage instance for reading and writing.
func New(name string, opts *Options) (*Storage, error) {

	var s *Storage
	var e error

	s.size.max = opts.MaxSize * 1024 * 1024

	s.opts = opts
	s.file.full = name
	s.file.path = path.Dir(name)
	s.file.base = path.Base(name)
	s.file.extn = path.Ext(name)
	s.file.name = s.file.base[:len(s.file.extn)]

	// Ensure the data directory exists
	if e = os.MkdirAll(s.file.path, 0744); e != nil {
		return nil, e
	}

	return s, e

}

// Close closes the underlying Syncable storage data stream, and prevents
// any further reads of writes to the Storage instance.
func (s *Storage) Close() error {

	s.lock.Lock()
	defer s.lock.Unlock()

	defer func() {
		s.opts = nil
		s.pntr = nil
		s.file.full = ""
		s.file.path = ""
		s.file.base = ""
		s.file.name = ""
		s.file.extn = ""
	}()

	if s.pntr == nil {
		return nil
	}

	return s.pntr.Close()

}

// Read reads up to len(p) bytes into p, reading from the Storage data
// stream, and automatically rotating files when the end of a file is
// reached. It returns the number of bytes read (0 <= n <= len(p)) and
// any error encountered. If the Storage has reached the final log file,
// then an EOF error will be returned.
func (s *Storage) Read(p []byte) (n int, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	n, err = s.pntr.Read(p)

	if err == io.EOF {
		if err = s.next(); err != nil {
			return
		}
	}

	return

}

// Write writes len(p) bytes from p to the underlying Storage data
// stream. It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early. Write
// will always append data to the end of the Storage data stream, ensuring
// data is append-only and never overwritten.
func (s *Storage) Write(p []byte) (n int, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	// No data file is open yet so open one.
	if s.pntr == nil {
		if err = s.last(); err != nil {
			return 0, err
		}
	}

	// We are not at the end of the stream.
	if s.file.seek != 2 {
		if err = s.last(); err != nil {
			return 0, err
		}
	}

	// We need to rotate to a new data file.
	if s.size.max < s.size.now+len(p) {
		if err = s.swap(); err != nil {
			return 0, err
		}
	}

	// Write to the data file.
	n, err = s.pntr.Write(p)

	// Update the data file size.
	s.size.now += n

	return

}

// Seek sets the offset for the next Read or Write to offset, interpreted
// according to whence: SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and SeekEnd means
// relative to the end. Seek returns the new offset relative to the start
// of the file and an error, if any. In some storage types, Seek only
// supports seeking to the beginning and end of the data stream.
func (s *Storage) Seek(offset int64, whence int) (int64, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	if offset != 0 {
		return 0, ErrSeekPositionNotSupported
	}

	if whence != 0 && whence != 2 {
		return 0, ErrSeekPositionNotSupported
	}

	s.file.seek = whence

	return 0, nil

}

// Sync ensures that any buffered data in the stream is committed to stable
// storage. In some Syncable types, Seek does nothing, as the data is written
// and persisted immediately when a Write is made. On Syncable types which
// support BufferWrites, then Sync will ensure the data stored is flushed.
func (s *Storage) Sync() error {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pntr.Sync()

}

// ---------------------------------------------------------------------------

// next opens the next log file for reading, checking all available log
// files and retrieving the next file in-order. If no next log file is
// available, then an EOF error will be returned.
func (s *Storage) next() error {

	files, err := ioutil.ReadDir(s.file.path)
	if err != nil {
		return err
	}

	for _, info := range files {

		// Ignore folders
		if info.IsDir() {
			continue
		}

		// Ignore invisible files
		if info.Name()[:1] == "." {
			continue
		}

		// Last file so open as last
		if info.Name() == s.file.base {
			return s.last()
		}

		// No current file so open first
		if s.pntr == nil {
			s.pntr, err = os.Open(path.Join(s.file.base, info.Name()))
			return err
		}

		// Current file so open next
		if info.Name() == s.pntr.Name() {
			continue
		}

		s.pntr, err = os.Open(path.Join(s.file.base, info.Name()))
		return err

	}

	// Otherwise open the last file as no other files exist

	return s.last()

}

// last ensures that the specified directory exists, and opens the latest
// file for writing, seeking to the end of the file, and storing the size.
func (s *Storage) last() error {

	var err error
	var mta os.FileInfo

	s.pntr, err = os.OpenFile(s.file.base, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	// Get the file metadata
	mta, err = s.pntr.Stat()
	if err != nil {
		return err
	}

	// Seek to the end of the file
	_, err = s.pntr.Seek(0, 2)
	if err != nil {
		return err
	}

	// Mark that we are at the end.
	s.file.seek = 2

	// Get the current data file size.
	s.size.now = int(mta.Size())

	return nil

}

// swap rotates out the current writable log file, and creates a new file
// for writing. The old file is renamed using the current data and time.
func (s *Storage) swap() error {

	var err error

	err = os.Rename(s.file.base, s.name())
	if err != nil {
		return err
	}

	s.pntr, err = os.OpenFile(s.file.base, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	s.size.now = 0

	return nil

}

// name creates a new file archive name for the current data file, using
// the current data and time, and the original file name.
func (s *Storage) name() string {
	return s.file.name + "-" + time.Now().UTC().Format(time.RFC3339) + "." + s.file.extn
}
