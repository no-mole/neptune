package tagger

type Tagger interface {
	Match(tag string) bool
	Add(tags []string)
}
