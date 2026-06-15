package filters

// Shared types for filter chips used across report, entry and departure views.
//
// A Chip is a label/value pair shown above a list to indicate what filter is
// currently applied. Each chip carries the form-field name (`Key`) it
// represents so the UI can remove that specific filter without rebuilding the
// whole form.
//
// FilterChips is the template-facing payload: `Items` is the list of chips
// to render and `OOB` is the render-strategy flag that toggles the
// `hx-swap-oob="true"` attribute on the chips container (needed when chips
// arrive as an out-of-band swap after a filter apply).

type ChipEntry struct {
	Key   string
	Label string
	Value string
}

type FilterChips struct {
	Items []ChipEntry
	OOB   bool
}
