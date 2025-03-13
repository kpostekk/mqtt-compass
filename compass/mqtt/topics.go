package mqtt

// import (
// 	"mqui/compass/mqtt/topics"
// 	"mqui/compass/utils"
// 	"strings"
// )

// type TopicTreeLeaf struct {
// 	TopicPath string
// 	Packet    topics.IncomingPacket
// 	Children  TopicTree
// 	Parent    *TopicTreeLeaf
// }

// type TopicTree map[string]*TopicTreeLeaf

// func (t *TopicTree) Add(topicPath string, packet *topics.IncomingPacket, parent *TopicTreeLeaf) *TopicTreeLeaf {
// 	parts := strings.Split(topicPath, "/")

// 	if len(parts) == 0 {
// 		return nil
// 	}

// 	// log.Default().Println(parts)

// 	currentPart := parts[0]
// 	reminder := parts[1:]

// 	// if reminder is empty, we are at the end of the topic
// 	if len(reminder) == 0 {
// 		(*t)[currentPart] = &TopicTreeLeaf{
// 			Packet:   *packet,
// 			Parent:   parent,
// 			Children: make(TopicTree),
// 		}

// 		return (*t)[currentPart]
// 	}

// 	// if reminder is not empty, we need to go deeper
// 	if _, ok := (*t)[currentPart]; !ok {
// 		(*t)[currentPart] = &TopicTreeLeaf{
// 			Packet:   nil,
// 			Parent:   parent,
// 			Children: make(TopicTree),
// 		}
// 	}

// 	(*t)[currentPart].Children.Add(strings.Join(reminder, "/"), packet, (*t)[currentPart])

// 	return (*t)[currentPart]
// }

// func (t *TopicTreeLeaf) RenderString(depth int) string {
// 	const INDENT = "    "
// 	sb := strings.Builder{}

// 	for i := 0; i < depth; i++ {
// 		sb.WriteString(INDENT)
// 	}

// 	sb.WriteString(t.Packet.Topic)

// 	if t.Packet != nil {
// 		sb.WriteString(" = ")
// 		sb.WriteString(utils.TruncateString(string(t.Packet.Payload), 48))
// 	}

// 	for _, v := range t.Children {
// 		sb.WriteString("\n")
// 		sb.WriteString(INDENT)
// 		sb.WriteString(v.RenderString(depth + 1))
// 	}

// 	return sb.String()
// }

// func (t *TopicTreeLeaf) String() string {
// 	return t.RenderString(0)
// }

