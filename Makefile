all: build pack install
build:
	go build -ldflags '-s -w'
pack:
	upx -q --lzma smux > /dev/null || true
install:
	install smux ~/bin/smux 

keys:
	@mkdir -p test/fixture
	if [ ! -f test/fixture/jwf.clear ]; then \
	yes | ssh-keygen -N ''    -C ecdsa-256 -f test/fixture/jwf.clear -b 256 -t ecdsa >/dev/null ;\
	yes | ssh-keygen -N 12345 -C rsa1k-ssh -f test/fixture/jwf.ossh -b 1024         >/dev/null ;\
	yes | ssh-keygen -N 12345 -C rsa1k-pem -f test/fixture/jwf.pem -b 1024 -m pem   >/dev/null ;\
	yes | ssh-keygen -N 12345 -C ec25519   -f test/fixture/jwf.ec19 -t ed25519      >/dev/null ;\
	fi
	#rm *.pub

test_crude: keys
	echo -n 12345 | go run smux.go "ssh-add ./test/fixture/jwf.ossh ./test/fixture/jwf.ec19 ./test/fixture/jwf.clear"
	                go run smux.go "ssh-add ./test/fixture/jwf.pem" md5sum
	ssh-add -l
	echo -n 12345|md5sum

clean:
	ssh-add -D
	rm test/fixture/jwf.* 2>/dev/null || true


