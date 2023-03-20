//go:build tools
// +build tools

package main

import (
	_ "github.com/deepmap/oapi-codegen/cmd/oapi-codegen"
	_ "github.com/go-jet/jet/v2/cmd/jet"
	_ "github.com/leonnicolas/genstrument"
	_ "github.com/pressly/goose/cmd/goose"
)
