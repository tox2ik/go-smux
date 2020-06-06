# Standard input multiplexer

This program will ask you for a password and forward that to specified commands.
There are utilities that do this already, but they do not work with `ssh-add`.

## Example

    smux "ssh-add $HOME/.ssh/id_ra $HOME/secret/key.rsa" 'sudo cryptsetup open /dev/encrypted my-secrets'
    passphrase: *****************
    added /home/user/secret/key.rsa
    added /home/user/.ssh/id_rsa


