all: build pack install
build:
	go build
pack:
	upx -q --lzma smux > /dev/null || true
install:
	install smux ~/bin/smux 

keys:
	yes | ssh-keygen -N ''    -f jwf.clear -b 256 -t ecdsa >/dev/null
	yes | ssh-keygen -N 12345 -f jwf.ossh -b 1024         >/dev/null
	yes | ssh-keygen -N 12345 -f jwf.pem -b 1024 -m pem   >/dev/null
	yes | ssh-keygen -N 12345 -f jwf.ec19 -t ed25519      >/dev/null
	rm *.pub

test_crude: keys
	echo -n 12345 | go run smux.go "ssh-add ./jwf.ossh ./jwf.ec19 ./jwf.clear"
	                go run smux.go "ssh-add ./jwf.pem" md5sum
	ssh-add -l
	echo -n 12345|md5sum

clean:
	ssh-add -D
	rm jwf.* 2>/dev/null || true


