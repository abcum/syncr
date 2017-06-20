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

package gcs

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/abcum/bump"

	"cloud.google.com/go/storage"
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
	// MinSize is used to determine the number of megabytes to write to
	// a file before the file should be rotated, and a new file used. This
	// is only available on certain types which support rotation of files,
	// but do not support appendable writes. The file is overwritten on
	// each write using the buffered data. MinSize should be specified in
	// megabytes.
	MinSize int
}

// Storage represents a Amazon S3 reader and writer.
type Storage struct {
	sync.Mutex
	opts *Options
	lock sync.Mutex
	size struct {
		min int
		now int
	}
	serv struct {
		srv *storage.Client
		ctx context.Context
		bck *storage.BucketHandle
	}
	buff struct {
		bit []byte
		wtr *bump.Writer
		rdr io.ReadCloser
	}
	file struct {
		buck string
		last string
		full string
		path string
		base string
		name string
		extn string
		main string
		poss []string
	}
}

// New creates a new Syncable storage instance for reading and writing.
func New(name string, opts *Options) (s *Storage, e error) {

	p := strings.Index(name, "/")

	s = &Storage{}

	s.opts = opts
	s.file.buck = name[:p]
	s.file.full = name[p+1:]
	s.file.path = path.Dir(s.file.full)
	s.file.base = path.Base(s.file.full)
	s.file.extn = path.Ext(s.file.full)
	s.file.name = s.file.base[:len(s.file.base)-len(s.file.extn)]
	s.file.main = path.Join(s.file.path, s.file.base)

	s.serv.ctx = context.Background()

	s.buff.wtr = bump.NewWriterBytes(&s.buff.bit)

	s.size.min = opts.MinSize * 1024 * 1024

	s.serv.srv, e = storage.NewClient(s.serv.ctx)

	s.serv.bck = s.serv.srv.Bucket(s.file.buck)

	return

}

// Close closes the underlying Syncable storage data stream, and prevents
// any further reads of writes to the Storage instance.
func (s *Storage) Close() error {

	s.lock.Lock()
	defer s.lock.Unlock()

	defer func() {
		s.opts = nil
		s.buff.bit = nil
		s.buff.wtr = nil
		s.buff.rdr = nil
		s.file.last = ""
		s.file.full = ""
		s.file.path = ""
		s.file.base = ""
		s.file.name = ""
		s.file.extn = ""
		s.file.main = ""
		s.file.poss = nil
	}()

	s.serv.srv.Close()

	return s.stop()

}

// Read reads up to len(p) bytes into p, reading from the Storage data
// stream, and automatically rotating files when the end of a file is
// reached. It returns the number of bytes read (0 <= n <= len(p)) and
// any error encountered. If the Storage has reached the final log file,
// then an EOF error will be returned.
func (s *Storage) Read(p []byte) (int, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	// No data file is open yet so open one.
	if s.buff.rdr == nil {
		if err := s.head(); err != nil {
			return 0, err
		}
	}

	n, err := s.buff.rdr.Read(p)
	if err == io.EOF {
		err = s.next()
	}

	return n, err

}

// Write writes len(p) bytes from p to the underlying Storage data
// stream. It returns the number of bytes written from p (0 <= n <= len(p))
// and any error encountered that caused the write to stop early. Write
// will always append data to the end of the Storage data stream, ensuring
// data is append-only and never overwritten.
func (s *Storage) Write(p []byte) (int, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.size.min < s.size.now+len(p) {
		if err := s.sync(); err != nil {
			return 0, err
		}
	}

	s.size.now += len(p)

	return len(p), s.buff.wtr.WriteBytes(p)

}

// Seek sets the offset for the next Read or Write to offset, interpreted
// according to whence: SeekStart means relative to the start of the file,
// SeekCurrent means relative to the current offset, and SeekE8nd means
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

	if whence == 0 {
		s.head()
	}

	return 0, nil

}

// Sync ensures that any buffered data in the stream is committed to stable
// storage. In some Syncable types, Seek does nothing, as the data is written
// and persisted immediately when a Write is made. On Syncable types which
// support BufferWrites, then Sync will ensure the data stored is flushed.
func (s *Storage) Sync() error {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.sync()

}

// ---------------------------------------------------------------------------

func (s *Storage) load(file string) (err error) {

	s.buff.rdr, err = s.serv.bck.Object(file).NewReader(s.serv.ctx)

	s.file.last = file

	return nil

}

func (s *Storage) sync() (err error) {

	if s.size.now == 0 {
		return nil
	}

	wtr := s.serv.bck.Object(s.name()).NewWriter(s.serv.ctx)

	if _, err = wtr.Write(s.buff.bit); err != nil {
		return err
	}

	if err = s.buff.wtr.ResetBytes(&s.buff.bit); err != nil {
		return err
	}

	s.size.now = 0

	return wtr.Close()

}

func (s *Storage) stop() (err error) {

	if s.buff.rdr == nil {
		return nil
	}

	defer func() {
		s.buff.rdr = nil
	}()

	return s.buff.rdr.Close()

}

func (s *Storage) head() (err error) {

	err = s.stop()
	if err != nil {
		return err
	}

	for _, file := range s.look() {
		return s.load(file)
	}

	return io.EOF

}

func (s *Storage) next() (err error) {

	err = s.stop()
	if err != nil {
		return err
	}

	for _, file := range s.look() {
		if file > s.file.last {
			return s.load(file)
		}
	}

	return io.EOF

}

func (s *Storage) look() []string {

	if s.file.poss != nil {
		return s.file.poss
	}

	i := s.serv.bck.Objects(s.serv.ctx, &storage.Query{
		Prefix: s.file.path,
	})

	for {

		o, e := i.Next()
		if e != nil {
			break
		}

		// Ignore folders and invisibles
		if o.Name[:1] == "." {
			continue
		}

		s.file.poss = append(s.file.poss, o.Name)

	}

	return s.file.poss

}

func (s *Storage) name() string {
	return path.Join(s.file.path, s.file.name+"-"+s.time()+s.file.extn)
}

func (s *Storage) time() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05.999999999")
}
