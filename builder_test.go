// Copyright 2020 Matthew Holt
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

package xcaddy

import (
	"fmt"
	"reflect"
	"testing"
)

func TestReplacementPath_Param(t *testing.T) {
	tests := []struct {
		name string
		r    ReplacementPath
		want string
	}{
		{
			"Empty",
			ReplacementPath(""),
			"",
		},
		{
			"ModulePath",
			ReplacementPath("github.com/x/y"),
			"github.com/x/y",
		},
		{
			"ModulePath Version Pinned",
			ReplacementPath("github.com/x/y v0.0.0-20200101000000-xxxxxxxxxxxx"),
			"github.com/x/y@v0.0.0-20200101000000-xxxxxxxxxxxx",
		},
		{
			"FilePath",
			ReplacementPath("/x/y/z"),
			"/x/y/z",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.r.Param())
			if got := tt.r.Param(); got != tt.want {
				t.Errorf("ReplacementPath.Param() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewReplace(t *testing.T) {
	type args struct {
		old string
		new string
	}
	tests := []struct {
		name string
		args args
		want Replace
	}{
		{
			"Empty",
			args{"", ""},
			Replace{"", ""},
		},
		{
			"Constructor",
			args{"a", "b"},
			Replace{"a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewReplace(tt.args.old, tt.args.new); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEmbedDir(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  EmbedDir
	}{
		{
			"Simple embed",
			"/home/my/cool/files",
			EmbedDir{
				Dir: "/home/my/cool/files",
			},
		},
		{
			"Aliased embed",
			"www:/home/my/cool/files",
			EmbedDir{
				Dir:  "/home/my/cool/files",
				Name: "www",
			},
		},
		{
			"Aliased embed where alias=`all:`",
			"all:/home/my/cool/files",
			EmbedDir{
				Dir:  "/home/my/cool/files",
				Name: "all",
			},
		},
		{
			"Simple embed including hidden files with `all::`",
			"all::/home/my/cool/files",
			EmbedDir{
				Dir:        "/home/my/cool/files",
				IncludeAll: true,
			},
		},
		{
			"Aliased embed including hidden files with `all::`",
			"all:www:/home/my/cool/files",
			EmbedDir{
				Dir:        "/home/my/cool/files",
				Name:       "www",
				IncludeAll: true,
			},
		},
		/*
			{
				"Edge case: Single leading colon", // TODO(camhux): Is this the correct expectation?
				":/home/my/cool/files",
				EmbedDir{
					Dir: "/home/my/cool/files",
				},
			},
			{
				"Edge case: Double leading colon", // TODO(camhux): Is this the correct expectation?
				"::/home/my/cool/files",
				EmbedDir{
					Dir: "/home/my/cool/files",
				},
			},
			{
				"Edge case: Single leading colon with alias", // TODO(camhux): Is this the correct expectation?
				":www:/home/my/cool/files",
				EmbedDir{
					Dir:  "/home/my/cool/files",
					Name: "www",
				},
			},
			{
				"Edge case: Junk prefix with alias", // TODO(camhux): Is this the correct expectation?
				"???:www:/home/my/cool/files",
				EmbedDir{
					Dir:  "/home/my/cool/files",
					Name: "www",
				},
			},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEmbedDir(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEmbedDir() = %v, want %v", got, tt.want)
			}
		})
	}
}
