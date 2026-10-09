package types

// StreamsCatalog is a catalog in one format: Streams (streams.json) for a legacy catalog, or
// Available + Selected (available_streams.json, selected_streams.json) for a split catalog.
type StreamsCatalog struct {
	Streams   string
	Available string
	Selected  string
}

// IsSplit reports whether the catalog is in the split format.
func (c StreamsCatalog) IsSplit() bool {
	return c.Available != "" && c.Selected != ""
}
