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

package minio

import "errors"

var (
	// ErrSeekPositionNotSupported occurs when a syncable type receives a
	// seek destination which it can not fulfil to. These syncable types
	// usually only support seeking to the beginning or to the end, as they
	// write to and read from multiple data files.
	ErrSeekPositionNotSupported = errors.New("Seek position is not supported.")
)

// Options represents syncing configuration options.
type Options struct {
	// MinSize is used to determine the number of megabytes to write to
	// a file before the file should be rotated, and a new file used. This
	// is only available on certain types which support rotation of files,
	// but do not support appendable writes. The file is overwritten on
	// each write using the buffered data. MinSize should be specified in
	// megabytes.
	MinSize int
	// BufferWrites can be used to determine whether writes are buffered
	// up to a certain number of bytes, before being synced. This is only
	// available on syncable types which do not support appendable writes.
	// BufferWrites should be specified in megabytes.
	BufferWrites int
}

type Storage struct {
	opts *Options
}

func New(path string, opts *Options) (*Storage, error) {

	var s *Storage
	var e error

	s.opts = opts

	return s, e

}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) Read(p []byte) (int, error) {
	return 0, nil
}

func (s *Storage) Write(p []byte) (int, error) {
	return 0, nil
}

func (s *Storage) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (s *Storage) Sync() error {
	return nil
}
