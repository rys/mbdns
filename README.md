# mbdns

`mbdns` is a dynamic DNS update client for the [Mythic Beasts](https://www.mythic-beasts.com/support/api/primary) Primary DNS v1 system, supporting their IPv4 and IPv6 endpoints. Written in [golang](https://golang.org), it supports common home network infrastructure operating systems (FreeBSD, EdgeOS, Linux) and common platform architectures (amd64, arm64 and MIPS).

## Configuring mbdns

Take [doc/mbdns.conf.sample](/doc/mbdns.conf.sample) and configure it to your needs, saving as `mbdns.conf`. `mbdns.conf` must be valid JSON or `mbdns` will fail to start.

## Deploying mbdns

* Copy the `mbdns` binary (usually named `mbdns-VERSION-OS-ARCH`) and `mbdns.conf` to your target platform
* chmod 0400 mbdns.conf so the Mythic Beasts API tokens can only be read by the user you deploy as (ideally use a specific user). `mbdns` will check for you and fail to start if the config is insecurely readable.
* Run the `mbdns` binary. It will run until you kill it.
* By default the `mbdns` binary will look for `./mbdns.conf`. You can relocate it (and rename it) and tell `mbdns` with `--config`.
  * `mbdns --config /etc/mbdns/mbdns.conf`

`mbdns` is written in golang and builds as a statically linked library with no dependencies.

## Practical running

`mbdns` logs to `stdout`. If your target OS supports it, redirect stdout to a file, logrotate that file, and run the binary in the background via the daemonising system of your choice. `mbdns` makes no attempt to daemonise itself.

systemd will run it nicely if you use the `simple` service type, running the process for you and capturing `stdout`.

`mbdns --version` prints the version information and exits immediately.

Running on a Unix-like might go something like this (assuming `mbdns` is in `$PATH`):

`nohup mbdns --config /etc/mbdns/mbdns.conf > /var/log/mbdns 2>&1 &`

## Logging

`mbdns` logs the following:

* version, build date and git commit SHA on startup
* the path to the config file
* the update command it is attempting to run for the host and domain
* success or failure including record type and TTL in both cases
* a small handful of at-startup failure messages if it can't run (no conf, invalid JSON, insecure conf)
* a message saying it is processing records if it can cleanly start

## Platform specific instructions

- [UniFi Dream Machine SE](/doc/udm-se/README.md)

## License

`mbdns` is MIT licensed.