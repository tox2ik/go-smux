all: build pack install
build:
	go build -ldflags '-s -w'
pack:
	upx -q --lzma smux > /dev/null || true
install:
	install smux ~/bin/smux 

keys:
	@mkdir -p testdata
	if [ ! -f testdata/jwf.clear ]; then \
	yes | ssh-keygen -N ''    -C ecdsa-256 -f testdata/jwf.clear -b 256 -t ecdsa >/dev/null ;\
	yes | ssh-keygen -N 12345 -C rsa1k-ssh -f testdata/jwf.ossh -b 1024         >/dev/null ;\
	yes | ssh-keygen -N 12345 -C rsa1k-pem -f testdata/jwf.pem -b 1024 -m pem   >/dev/null ;\
	yes | ssh-keygen -N 12345 -C ec25519   -f testdata/jwf.ec19 -t ed25519      >/dev/null ;\
	fi

clean:
	rm -f testdata/jwf.* 2>/dev/null || true
	rm -f smux

test:
	go test
