//go:build !no_cgo

package main

// #cgo CFLAGS: -Wno-return-local-addr
import "C"

import (
	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
)
