protoc --go_out=plugins=grpc:./common --go_opt=module=github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/common common/common.proto
protoc --go_out=plugins=grpc:./admin --go_opt=module=github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/admin admin/admin.proto
protoc --go_out=plugins=grpc:./chat --go_opt=module=github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/chat chat/chat.proto
