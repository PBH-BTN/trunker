init_api:
	hz new --model_dir biz/hertz_gen -mod github.com/PBH-BTN/trunker -idl idl/api.thrift


update_api:
	hz update --model_dir biz/hertz_gen -idl idl/api.thrift