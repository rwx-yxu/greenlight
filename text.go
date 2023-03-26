package greenlight

import _ "embed"

// Keep all documentation in this file as embeds in order to easily
// support compiles to other target languages by simply changing the
// language identifier before compilation.

//go:embed text/en/greenlight.md
var _greenlight string

//go:embed text/en/start.md
var _start string
