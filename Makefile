update_idl:
	kitex -module "github.com/PBH-BTN/trunker" -thrift frugal_tag,template=slim,no_default_serdes idl/server.thrift
