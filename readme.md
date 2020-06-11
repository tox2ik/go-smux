# Standard input multiplexer

This program will ask you for a password and forward that to specified commands.
There are utilities that do this already, but they do not work with `ssh-add`.

## Example

Init ssh-agent and unlock an encrypted volume

    $ smux "ssh-add $HOME/.ssh/id_rsa $HOME/secret/key.rsa" 'sudo cryptsetup open /dev/encrypted my-secrets'
    passphrase: *****************
    added /home/user/secret/key.rsa
    added /home/user/.ssh/id_rsa

Pipe input to several utilities

    $ echo Works similar to moreutils pee | smux sha224sum sha1sum md5sum
    c015d10f322ab6c5e221262acb598872  -
    debeca6bf61ffa188cb359258c9ad99e69ab15cf  -
    5fbedfb30d010478ff9e9bdbd75a835d7b17e2782f2f95b3b106ce89  -

## Environment variables

    SSH_ADD_LIFE  - life time of stored keys, in seconds


## References

- https://stackoverflow.com/questions/38094555/golang-read-os-stdin-input-but-dont-echo-it
- https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go
- https://stackoverflow.com/questions/10385551/get-exit-code-go
- https://flaviocopes.com/go-shell-pipes/
- keysAvailable(agent Socket, identities []string)  
  https://bitbucket.org/rw_grim/convey/src/default/ssh/agent.go
- https://unix.stackexchange.com/questions/28503/how-can-i-send-stdout-to-multiple-commands
- https://stackoverflow.com/questions/11886531/terminating-a-process-started-with-os-exec-in-golang
