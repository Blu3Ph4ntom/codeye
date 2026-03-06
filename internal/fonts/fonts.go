// Package fonts embeds the bundled Nerd Font and provides install helpers.
package fonts

import _ "embed"

// JetBrainsMonoRegular contains the JetBrainsMono Nerd Font v3 Regular TTF.
// Embedded at build time via go:embed.
//
//go:embed JetBrainsMonoNerdFont-Regular.ttf
var JetBrainsMonoRegular []byte
