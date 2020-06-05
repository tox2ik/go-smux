

keys:
	yes | ssh-keygen -N 12345 -f jwf.ossh -b 1024       >/dev/null
	yes | ssh-keygen -N 12345 -f jwf.pem -b 1024 -m pem >/dev/null
	rm *.pub

test_crude: keys
	go run smux.go "ssh-add ./jwf.ossh jwf.pem" 'md5sum'
	ssh-add -l

clean:
	ssh-add -D
	rm jwf.*
