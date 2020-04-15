package handle

import (
	"fmt"
	"tea/src/manage"
	"tea/src/mqtt/protocol"
	"tea/src/mqtt/sub"
	"tea/src/utils"
)

type PublishPack struct {
	protocol.Pack
	dup        int
	qos        int
	retain     int
	topicName  string
	identifier uint16
	payload    []byte
}

func (A PublishPack) GetCmd() byte {
	return 0x3
}

func (A PublishPack) GetFixedHeaderWithoutLength() byte {

	return 0x3 << 4
}

func (A PublishPack) GetVariableHeader() ([]byte, bool) {
	return append([]byte(A.topicName), utils.Uint16ToBytes(A.identifier)...), true
}

func (A PublishPack) GetPayload() ([]byte, bool) {
	return A.payload, true
}

type Publish struct {
}

func newPublishPack(pack protocol.Pack) *PublishPack {
	fmt.Println("hello")
	p := new(PublishPack)
	p.Pack = pack
	p.retain = int(pack.FixedHeader[0] & 0x1)
	p.qos = int(pack.FixedHeader[0] & 0x6)
	p.dup = int(pack.FixedHeader[0] & 0x8)

	plc := pack.FixHeaderLength

	topicNameLength := utils.UtfLength(pack.Data[plc : plc+2])
	plc += 2
	p.topicName = string(pack.Data[plc : plc+topicNameLength])
	fmt.Println(p.topicName, "topicName")
	plc += topicNameLength
	if p.qos > 0 {
		p.identifier = utils.BytesToUint16(pack.Data[plc : plc+2])
		plc += 2
	}
	payloadLength := pack.BodyLength - (plc - pack.FixHeaderLength)

	p.payload = pack.Data[plc : plc+payloadLength]

	fmt.Println(string(p.payload))

	return p

}

func NewPublish() *Publish {

	return new(Publish)
}
func (p *Publish) Handle(pack protocol.Pack, client *manage.Client) {

	publishPack := newPublishPack(pack)

	clientList, ok := sub.GetTopicSubClients(publishPack.topicName)

	fmt.Println("receiver publish payload", string(publishPack.payload))

	if ok {
		clients := clientList.GetNode()

		for _, clientId := range clients {

			if c, ok := client.Manage.GetClient(clientId); ok {

				c.Write(pack.Data)
				//protocol.Encode(publishPack, c)
			}

		}
		fmt.Println(clients)
	}

	//todo 根据topic 找寻订阅topic的客户端将消息发布给客户端

}