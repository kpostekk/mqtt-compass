package main

import (
	"strings"
)

type TopicTreeLeaf struct {
	Topic    string
	Value    []byte
	Children TopicTree
	Parent   *TopicTreeLeaf
}

type TopicTree map[string]*TopicTreeLeaf

type TopicMap map[string][]byte

func (tm *TopicMap) IntoTree() *TopicTree {
	t := make(TopicTree)

	for k, v := range *tm {
		t.Add(k, v)
	}

	return &t
}

func (t *TopicTree) Add(topic string, value []byte) *TopicTreeLeaf {
	return t.AddRecursive(topic, value, nil)
}

func (t *TopicTree) AddRecursive(topic string, value []byte, parent *TopicTreeLeaf) *TopicTreeLeaf {
	parts := strings.Split(topic, "/")

	if len(parts) == 0 {
		return nil
	}

	// log.Default().Println(parts)

	currentPart := parts[0]
	reminder := parts[1:]

	// if reminder is empty, we are at the end of the topic
	if len(reminder) == 0 {
		(*t)[currentPart] = &TopicTreeLeaf{
			Topic:    currentPart,
			Value:    value,
			Parent:   parent,
			Children: make(TopicTree),
		}

		return (*t)[currentPart]
	}

	// if reminder is not empty, we need to go deeper
	if _, ok := (*t)[currentPart]; !ok {
		(*t)[currentPart] = &TopicTreeLeaf{
			Topic:    currentPart,
			Value:    nil,
			Parent:   parent,
			Children: make(TopicTree),
		}
	}

	(*t)[currentPart].Children.AddRecursive(strings.Join(reminder, "/"), value, (*t)[currentPart])

	return (*t)[currentPart]
}

func (t *TopicTreeLeaf) TopicCanonical() string {
	if t.Parent == nil {
		return t.Topic
	}

	return t.Parent.TopicCanonical() + "/" + t.Topic
}

func truncateString(s string, length int) string {
	isLenExceeded := len(s) > length
	if isLenExceeded {
		return s[:length] + "..."
	}

	return s
}

func (t *TopicTreeLeaf) RenderString(depth int) string {
	const INDENT = "    "
	sb := strings.Builder{}

	for i := 0; i < depth; i++ {
		sb.WriteString(INDENT)
	}

	sb.WriteString(t.Topic)

	if t.Value != nil {
		sb.WriteString(" = ")
		sb.WriteString(truncateString(string(t.Value), 48))
	}

	for _, v := range t.Children {
		sb.WriteString("\n")
		sb.WriteString(INDENT)
		sb.WriteString(v.RenderString(depth + 1))
	}

	return sb.String()
}

func (t *TopicTreeLeaf) String() string {
	return t.RenderString(0)
}

func (t *TopicTree) RenderString() string {
	sb := strings.Builder{}

	for _, v := range *t {
		sb.WriteString(v.RenderString(0))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (t *TopicTree) String() string {
	return t.RenderString()
}

func (t *TopicTreeLeaf) ContainsCanonicalTopic(topic string) bool {
	if t.TopicCanonical() == topic {
		return true
	}

	for _, v := range t.Children {
		if v.ContainsCanonicalTopic(topic) {
			return true
		}
	}

	return false
}