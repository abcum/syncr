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

package syncr

//go:generate go get -u github.com/abcum/tmpl

//go:generate cp test.tmpl file/file_test.go.tmpl
//go:generate tmpl -file=file/file_test.json file/file_test.go.tmpl
//go:generate rm file/file_test.go.tmpl

//go:generate cp test.tmpl gcs/gcs_test.go.tmpl
//go:generate tmpl -file=gcs/gcs_test.json gcs/gcs_test.go.tmpl
//go:generate rm gcs/gcs_test.go.tmpl

//go:generate cp test.tmpl logr/logr_test.go.tmpl
//go:generate tmpl -file=logr/logr_test.json logr/logr_test.go.tmpl
//go:generate rm logr/logr_test.go.tmpl

//go:generate cp test.tmpl s3/s3_test.go.tmpl
//go:generate tmpl -file=s3/s3_test.json s3/s3_test.go.tmpl
//go:generate rm s3/s3_test.go.tmpl
