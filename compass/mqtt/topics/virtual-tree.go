package topics

import (
	"strings"
)

type VirtualTopicTree struct {
	Parent         *VirtualTopicTree
	RelatedPacket  *IncomingPacket
	ContextualPath ContextualPath

	Children map[string]*VirtualTopicTree
}

func (tree *VirtualTopicTree) AddPacket(packet *IncomingPacket) (*VirtualTopicTree) {
	// if packet and current path are equal, we are at the end of the path
	packetPath := packet.ContextualPath()
	currentContextualPath := &tree.ContextualPath

	if len(packetPath) == len(*currentContextualPath) && !packetPath.Equal(currentContextualPath) {
		panic("Packet path and current path are not equal, but have the same length!")
	}

	if currentContextualPath.Equal(&packetPath) {
		tree.RelatedPacket = packet
		return tree
	}

	// if not, create leaf and go deeper
	nextLeafName := packetPath[len(*currentContextualPath)]
	nextLeaf := CreateVirtualTopicTree()

	// set parent and path
	nextLeaf.Parent = tree
	nextLeaf.ContextualPath = append(*currentContextualPath, nextLeafName)

	// if leaf does not exist, create it
	if _, ok := tree.Children[nextLeafName]; !ok {
		tree.Children[nextLeafName] = nextLeaf
	}

	// recursively insert packet
	return tree.Children[nextLeafName].AddPacket(packet)
}

func (tree *VirtualTopicTree) IsRoot() bool {
	return len(tree.ContextualPath) == 0
}

func (tree *VirtualTopicTree) IsVirtual() bool {
	return tree.RelatedPacket == nil
}

func (tree *VirtualTopicTree) String() string {
	sb := strings.Builder{}

	if tree.IsRoot() {
		sb.WriteString("[Root]")
	}

	sb.WriteString(tree.ContextualPath.String())

	if tree.IsVirtual() {
		sb.WriteString(" [virtual]")
	}

	sb.WriteString("\n")

	for _, child := range tree.Children {
		sb.WriteString(child.String())
	}

	return sb.String()
}

func CreateVirtualTopicTree() *VirtualTopicTree {
	return &VirtualTopicTree{
		RelatedPacket:  nil,
		ContextualPath: RootContextualPath(),
		Children:       make(map[string]*VirtualTopicTree),
	}
}
