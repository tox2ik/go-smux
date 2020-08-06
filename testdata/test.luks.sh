#!/bin/bash

sudo cryptsetup close test-smux 
losetup | awk '/crypto-fs.block/ { print $1}' | xargs --no-run-if-empty -n1 losetup -d

if ! cryptsetup isLuks crypto-fs.block; then
	rm -fv        crypto-fs.block
	fallocate     crypto-fs.block -l $((1024**2 *20)) 
	echo -n jwf > crypto-fs.key
	cryptsetup luksFormat -q crypto-fs.block crypto-fs.key
fi


loop=$(losetup crypto-fs.block -f --show)

echo "Type jwf; 
                (aka cryptsetup open $loop test-smux --key-file=crypto-fs.key)"
sudo ../smux "cryptsetup open $loop test-smux" --key-file=crypto-fs.key ' (jwf) '
lsblk /dev/mapper/test-smux -o type,size,path
