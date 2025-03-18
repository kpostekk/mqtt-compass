package topics

import "github.com/eclipse/paho.golang/packets"
import "strings"

const CONTEXTUAL_PATH_SEPARATOR = "/"

type IncomingPacket packets.Publish

func (packet *IncomingPacket) ContextualPath() ContextualPath {
	return strings.Split(packet.Topic, CONTEXTUAL_PATH_SEPARATOR)
}

type ContextualPath []string

func (path *ContextualPath) String() string {
	if len(*path) == 0 {
		return "{root}"
	}

	return strings.Join(*path, CONTEXTUAL_PATH_SEPARATOR)
}

func (path *ContextualPath) Suffix() string {
	if len(*path) == 0 {
		return path.String()
	}

	return (*path)[len(*path)-1]
}

func (path *ContextualPath) IsRoot() bool {
	return len(*path) == 1
}

func (path *ContextualPath) Equal(other *ContextualPath) bool {
	if len(*path) != len(*other) {
		return false
	}

	for i, part := range *path {
		if part != (*other)[i] {
			return false
		}
	}

	return true
}

func (path *ContextualPath) IsParentOf(other *ContextualPath) bool {
	if len(*path) >= len(*other) {
		return false
	}

	for i, part := range *path {
		if part != (*other)[i] {
			return false
		}
	}

	return true
}

func (path *ContextualPath) DirectParent() ContextualPath {
	if len(*path) == 0 {
		return nil
	}

	return (*path)[:len(*path)-1]
}

func (path *ContextualPath) Child(child string) ContextualPath {
	return append(*path, child)
}

func RootContextualPath() ContextualPath {
	return make(ContextualPath, 0)
}
