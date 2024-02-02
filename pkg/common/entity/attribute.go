package entity

import (
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/db/table/chat"
)

type AttributeExpand struct {
	chat.Attribute
	Address   string `json:"address"  description:"地址"`
	PublicKey string `json:"publicKey" description:"公钥"`
}
