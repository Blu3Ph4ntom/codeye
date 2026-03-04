package scanner

import "sort"

// SortLangs sorts a slice of LangStats by the given field.
// If desc is true, larger values come first.
func SortLangs(langs []LangStats, by string, desc bool) {
	sort.Slice(langs, func(i, j int) bool {
		var less bool
		switch by {
		case "files":
			less = langs[i].Files < langs[j].Files
		case "code":
			less = langs[i].Code < langs[j].Code
		case "blank":
			less = langs[i].Blank < langs[j].Blank
		case "comment":
			less = langs[i].Comment < langs[j].Comment
		case "lang", "name":
			less = langs[i].Name < langs[j].Name
		default: // "lines"
			less = langs[i].Lines < langs[j].Lines
		}
		if desc {
			return !less
		}
		return less
	})
}
