package config

import "embed"

// FontsFS embeds the DANFE font files so the binary is self-contained
// (same pattern as model/armazenda_database migrations/embedded assets).
// No runtime filesystem lookup: font paths baked in at build time would
// point to the build machine (see getFontPath regression).
//
//go:embed fonts/*.ttf
var FontsFS embed.FS
