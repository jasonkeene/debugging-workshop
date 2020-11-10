
## Deterministic Debugging

Derek will be presenting about deterministic debugging using the `rr` tool.
Since `rr` requires access to the CPU's PMCs you will need to have access to
either a physical machine running Linux or use a hypervisor that exposes the
PMCs required by `rr`. The VirtualBox VM we created for this class does not
support these PMCs.

If you would like to follow along and experiment with `rr` I setup a bare
metal AWS instance you can SSH into. I will share the IP address and password
with you during the workshop:

```
$ ssh ubuntu@<ipaddress>
$ rr bash -c "exit 1"
rr: Saving execution to trace directory `/home/ubuntu/.local/share/rr/bash-0'.
```

Please be mindful that other workshop participants will be logged into the
same machine as you. Feel free to upload whatever test programs you would
like to experiment with. The Go tooling is installed on the machine so you
can compile test programs from source code. You can also cross-compile
programs locally and upload binaries to test with. To upload a local
directory to the machine:

```
$ rsync -azi ./some-dir/ ubuntu@<ipaddress>:~/<your_name>/some-dir/
```

Note that all information stored on this machine will be destroyed after the
workshop has ended!
