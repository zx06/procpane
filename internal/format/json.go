package format

import (
	"encoding/json"
	"io"
	"time"

	"github.com/zx06/procpane/internal/model"
)

// WriteListenersJSON writes jq-friendly JSON.
func WriteListenersJSON(w io.Writer, listeners []model.Listener) error {
	resp := model.ListResponse{
		GeneratedAt: time.Now().UTC(),
		Listeners:   listeners,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(resp)
}
