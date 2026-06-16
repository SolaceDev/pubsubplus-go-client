// solace-messaging-go-client
//
// Copyright 2021-2026 Solace Corporation. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows && amd64
// +build windows,amd64

package ccsmp

// Compiler flags for windows/amd64, applied whether or not the experimental
// support tag is set so that headers always resolve and the unsupported-build
// guard can emit a clear #error instead of a misleading "header not found".
//
// winsock2.h must be force-included before solClient.h: on WIN32 the header
// requires winsock(2).h to have been included first (it checks _WINSOCKAPI_).

/*
#cgo CFLAGS: -I${SRCDIR}/lib/include -DSOLCLIENT_PSPLUS_GO -include winsock2.h
*/
import "C"
