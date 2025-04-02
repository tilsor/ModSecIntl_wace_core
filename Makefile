all: waceproto/wace.pb.go waceproto/wace_grpc.pb.go plugins

plugins: _plugins/model/trivial.so _plugins/model/trivial2.so _plugins/model/trivial_async.so	\
	_plugins/model/no_init.so _plugins/model/wrong_init.so	\
	_plugins/model/error_init.so _plugins/model/no_req.so	\
	_plugins/model/wrong_req.so _plugins/model/error_req.so	\
	_plugins/decision/test.so _plugins/decision/no_check.so	\
	_plugins/decision/wrong_check.so _plugins/decision/weighted_sum.so \
	_plugins/decision/error_check.so _plugins/decision/simple.so


FLAGS=
# For debuggin FLAGS = -gcflags="all=-N -l"

# Generate both pb.go files from proto file:
waceproto/wace.pb.go waceproto/wace_grpc.pb.go: wace.proto.intermediate
.INTERMEDIATE: wace.proto.intermediate
wace.proto.intermediate: wace.proto
	protoc --go_out=. --go-grpc_out=. $<

%.so: %.go
	go build $(FLAGS) -buildmode=plugin -o $@ $<

clean:
	rm -rf waceproto/* _plugins/model/*.so _plugins/decision/*.so
