package testdata

import "embed"

//go:embed *.json
var TestData embed.FS
