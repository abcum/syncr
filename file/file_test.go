// Code generated by https://github.com/abcum/tmpl
// Source file: file/file_test.go.tmpl
// DO NOT EDIT!

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
	"io"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var run = 5000
var sze int
var txt []byte
var out []byte

func init() {

	os.RemoveAll(".test.db")

	txt, _ = ioutil.ReadFile("../data.txt")

}

func TestMain(t *testing.T) {

	defer func() {
		os.RemoveAll(".test.db")
	}()

	var s *Storage

	Convey("Create a new file syncr instance", t, func() {
		h, e := New(".test.db")
		So(e, ShouldBeNil)
		s = h
	})

	Convey("Read data from the file syncr instance", t, func() {
		s.Seek(0, 0)
		b := make([]byte, 128)
		_, e := s.Read(b)
		So(e, ShouldEqual, io.EOF)
	})

	Convey("Write data to the file syncr instance", t, func() {
		s.Seek(0, 0)
		for i := 0; i < run; i++ {
			l, e := s.Write(txt)
			sze += l
			if e != nil {
				if e != io.EOF {
					panic(e)
				}
				break
			}
		}
		So(sze, ShouldEqual, run*len(txt))
	})

	Convey("Sync data to the file syncr instance", t, func() {
		So(s.Sync(), ShouldBeNil)
	})

	Convey("Read data from the file syncr instance", t, func() {
		s.Seek(0, 0)
		for i := 0; i <= run*50; i++ {
			b := make([]byte, 128)
			l, e := s.Read(b)
			out = append(out, b[:l]...)
			if e != nil {
				if e != io.EOF {
					panic(e)
				}
				break
			}
		}
		So(len(out), ShouldEqual, run*len(txt))
		So(out[:len(txt)], ShouldResemble, txt)
		So(out[len(out)-len(txt):], ShouldResemble, txt)
	})

	Convey("Close the file syncr instance", t, func() {
		So(s.Close(), ShouldBeNil)
	})

	Convey("Close again", t, func() {
		So(s.Close(), ShouldBeNil)
	})

}
