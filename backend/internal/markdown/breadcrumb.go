package markdown

import "strings"

type headingEntry struct {
	level int
	title string
}

// HeadingStack tracks heading hierarchy for chunk metadata.
type HeadingStack struct {
	items []headingEntry
}

func NewHeadingStack() *HeadingStack {
	return &HeadingStack{items: make([]headingEntry, 0, 8)}
}

// Update applies markdown heading hierarchy rules.
func (h *HeadingStack) Update(level int, title string) {
	if h == nil || level < 1 {
		return
	}
	title = strings.TrimSpace(title)
	if level == 1 {
		h.items = h.items[:0]
		h.items = append(h.items, headingEntry{level: level, title: title})
		return
	}
	for len(h.items) > 0 && h.items[len(h.items)-1].level >= level {
		h.items = h.items[:len(h.items)-1]
	}
	h.items = append(h.items, headingEntry{level: level, title: title})
}

func (h *HeadingStack) Breadcrumb() string {
	if h == nil || len(h.items) == 0 {
		return ""
	}
	parts := make([]string, 0, len(h.items))
	for _, item := range h.items {
		if item.title != "" {
			parts = append(parts, item.title)
		}
	}
	return strings.Join(parts, " > ")
}

func (h *HeadingStack) SectionTitle() string {
	if h == nil || len(h.items) == 0 {
		return ""
	}
	return h.items[len(h.items)-1].title
}

func (h *HeadingStack) Clone() *HeadingStack {
	if h == nil {
		return NewHeadingStack()
	}
	cloned := NewHeadingStack()
	cloned.items = append(cloned.items, h.items...)
	return cloned
}
