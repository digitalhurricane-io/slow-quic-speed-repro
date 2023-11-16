#!/bin/bash

# can be included in gradle deps like:
# implementation files("/path/to/mylib.aar")

# must add -androidapi 21 flag to make gomobile compatible with newer NDKs
# https://github.com/golang/go/issues/52470

# for real android devices
gomobile bind -ldflags="-s -w" -androidapi 21 -target=android/arm,android/arm64 -o mylib.aar .

# includes emulator as build target
# gomobile bind -ldflags="-s -w" -androidapi 21 -target=android -o mylib.aar .