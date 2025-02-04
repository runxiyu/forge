# Notes

## SSH Custom Authentication

This doesn't work yet because

* `ForceCommand` appears to take precedence over `command=` supplied by
  AuthorizedKeysCommand
* `AuthorizedKeysCommand` appears to not run at all when logging in via an
  empty password is available



### User creation

The `forge` user should have a restricted shell (?). The second field in its
`/etc/passwd` and `/etc/shadow` entries shall be empty.

### `/etc/ssh/sshd_config`

```
Match User forge
	AuthorizedKeysFile none
	AuthorizedKeysCommand /usr/libexec/lindenii/forge/ssh_authorized_keys_command %C %D %f %h %k %t %U %u
	AuthorizedKeysCommandUser root
	PasswordAuthentication yes
	PermitEmptyPasswords yes
	DisableForwarding yes
	PermitTTY no
	ForceCommand /usr/libexec/lindenii/forge/ssh_shell
```

### `/usr/libexec/lindenii/forge/ssh_authorized_keys_command`

This file and its corresponding directory must be owned by root and must not be
writable by group or other.

```sh
#!/bin/sh

# Allows any key to log in

# Prototype, so let's not care about race conditions and whatever in logging
printf 'Endpoints: %s\n' "$1" > /var/log/lindenii/authkeys.last
printf 'Routing domain: %s\n' "$2" >> /var/log/lindenii/authkeys.last
printf 'Fingerprint: %s\n' "$3" >> /var/log/lindenii/authkeys.last
printf 'Home: %s\n' "$4" >> /var/log/lindenii/authkeys.last
printf 'Key/cert base64: %s\n' "$5" >> /var/log/lindenii/authkeys.last
printf 'Key/cert type: %s\n' "$6" >> /var/log/lindenii/authkeys.last
printf 'UID: %s\n' "$7" >> /var/log/lindenii/authkeys.last
printf 'Username: %s\n' "$8" >> /var/log/lindenii/authkeys.last

[ "$8" != "forge" ] && exit

# BUG: validate that key/cert type and base64 are sanitized and will not lead to injections
printf 'command="/usr/libexec/lindenii/forge/ssh_shell --ssh-key-type %s --ssh-key-base64 %s" %s %s\n' "$6" "$5" "$6" "$5"
```

Now... uh, if I change the last line to `printf 'command="/usr/libexec/lindenii/forge/ssh_shell --ssh-key-type %s --ssh-key-base64 %s" %s %s\n' "$6" "$5" "$6" "$5"`, I can't log in any more and I get `debug1: /usr/libexec/lindenii/forge/ssh_authorized_keys_command %C %D %f %h %k %t %U %u:1: bad key options: unknown key option` in the debug log?

### Alternatives to the mess above

`ExposeAuthInfo`?
