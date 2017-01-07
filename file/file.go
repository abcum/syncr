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

package file

import (
	"os"
	"sync"
)

// Storage represents a file reader and writer.
type Storage struct {
	file string
	pntr *os.File
	lock sync.Mutex
}

// New creates a new Syncable storage instance for reading and writing.
func New(name string) (*Storage, error) {

	var s *Storage
	var e error

	s.file = name
	s.pntr, e = os.OpenFile(s.file, os.O_CREATE|os.O_RDWR, 0666)

	return s, e

}

// Close closes the underlying Syncable storage data stream, and prevents
// any further reads of writes to the Storage instance.
func (s *Storage) Close() error {

	s.lock.Lock()
	defer s.lock.Unlock()

	defer func() {
		s.file = ""
		s.pntr = nil
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
func (s *Storage) Read(p []byte) (int, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pntr.Read(p)

}

// Write writes len(p) bytes from p to the underlying Storage data
// stream. It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early. Write
// will always append data to the end of the Storage data stream, ensuring
// data is append-only and never overwritten.
func (s *Storage) Write(p []byte) (int, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pntr.Write(p)

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

	return s.pntr.Seek(offset, whence)

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
