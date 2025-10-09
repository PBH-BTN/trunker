update_idl:
	kitex -module "github.com/PBH-BTN/trunker" -thrift frugal_tag,template=slim,no_default_serdes idl/server.thrift
	go mod tidy

install_tool:
	go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
