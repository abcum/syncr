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

type Storage struct {
	file string
	pntr *os.File
	lock sync.Mutex
}

func New(name string) (*Storage, error) {

	var s *Storage
	var e error

	s.file = name
	s.pntr, e = os.OpenFile(s.file, os.O_CREATE|os.O_RDWR, 0666)

	return s, e

}

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

func (s *Storage) Read(p []byte) (int, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pntr.Read(p)

}

func (s *Storage) Write(p []byte) (int, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pntr.Write(p)

}

func (s *Storage) Seek(offset int64, whence int) (int64, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pntr.Seek(offset, whence)

}

func (s *Storage) Sync() error {

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pntr.Sync()

}
