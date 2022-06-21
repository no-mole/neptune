package tagger

func NewMatcher(tags []string) Tagger {
	m := &Matcher{tags: map[string]struct{}{}}
	m.Add(tags)
	return m
}

type Matcher struct {
	tags map[string]struct{}
	all  bool
}

func (m *Matcher) Match(tag string) bool {
	if m.all {
		return true
	}
	_, ok := m.tags[tag]
	return ok
}

func (m *Matcher) Add(tags []string) {
	if len(tags) == 0 {
		m.all = true
		return
	}
	for _, tag := range tags {
		if tag == "*" {
			m.all = true
			return
		}
		m.tags[tag] = struct{}{}
	}
}

var _ Tagger = &Matcher{}
