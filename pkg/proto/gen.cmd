#protoc --go_out=plugins=grpc:./common --go_opt=module=github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/common common/common.proto
#protoc --go_out=plugins=grpc:./admin --go_opt=module=github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/admin admin/admin.proto
#protoc --go_out=plugins=grpc:./chat --go_opt=module=github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/chat chat/chat.proto


protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative  common/common.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative  admin/admin.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:. --go-grpc_opt=paths=source_relative  chat/chat.proto
