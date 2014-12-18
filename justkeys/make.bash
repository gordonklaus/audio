#!/usr/bin/env bash
# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

mkdir -p jni/armeabi
goandroid build -ldflags="-shared" -o jni/armeabi/libjustkeys.so .
ndk-build NDK_DEBUG=1
ant debug
