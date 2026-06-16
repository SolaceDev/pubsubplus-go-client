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

//go:build windows && amd64 && !windows_shared_experimental
// +build windows,amd64,!windows_shared_experimental

package ccsmp

// Windows support is EXPERIMENTAL and disabled by default. Building this package
// on windows/amd64 without the build tag stops here with a clear message instead
// of an opaque linker error about undefined solClient_* symbols.
//
// To enable it, build with:  go build -tags windows_shared_experimental ./...
//
// Runtime requirements once enabled:
//   - libsolclient.dll must be on PATH (or beside the executable).
//   - For secure (TLS) connections, the OpenSSL 3 libraries libcrypto-3.dll and
//     libssl-3.dll must also be available at runtime. They are NOT redistributed
//     by this module; the application or its installer must provide them.

/*
#error "Solace Go SDK: Windows support is EXPERIMENTAL and disabled by default. Rebuild with -tags windows_shared_experimental. At runtime libsolclient.dll must be on PATH (or beside the executable); for secure (TLS) connections the OpenSSL 3 libraries libcrypto-3.dll and libssl-3.dll must also be present (not redistributed by this module)."
*/
import "C"
