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

//go:build windows && amd64 && windows_shared_experimental
// +build windows,amd64,windows_shared_experimental

package ccsmp

// EXPERIMENTAL: Windows support is opt-in and unsupported. It is linked only when
// the build tag "windows_shared_experimental" is set, e.g.:
//
//	go build -tags windows_shared_experimental ./...
//
// This links dynamically against the MSVC (VS2015) shared library via its import
// library (libsolclient.lib); libsolclient.dll must be present at runtime. OpenSSL
// (libcrypto-3.dll / libssl-3.dll) is NOT redistributed by this module and is loaded
// at runtime by CCSMP only when secure (TLS) connections are used -- the application
// or its installer is responsible for providing it (same model as darwin).
//
// The C compiler flags (include path, winsock2.h) live in
// includes_windows_amd64_common.go so they apply on windows/amd64 regardless of the
// experimental tag; only the link step is gated here.

/*
#cgo LDFLAGS: -L${SRCDIR}/lib/windows_amd64 -l:libsolclient.lib -lws2_32 -ladvapi32 -liphlpapi
*/
import "C"
