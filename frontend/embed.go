package frontend

import "embed"

// Assets holds the production Vite build (frontend/dist).
//
//go:embed all:dist
var Assets embed.FS
