all: model/trivial.so model/trivial2.so \
	model/no_init.so model/wrong_init.so model/error_init.so \
	model/no_req.so model/wrong_req.so model/error_req.so \
	decision/test.so \
	decision/no_check.so decision/wrong_check.so decision/error_check.so

%.so: %.go
	go build $(FLAGS) -buildmode=plugin -o $@ $<

clean:
	rm -f model/*.so decision/*.so
