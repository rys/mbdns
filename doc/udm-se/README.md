# Using mbdns on a Ubiquiti UniFi Dream Machine SE

The instructions here are for a UDM SE, which has a different filesystem layout to other Dream Machine products (and also a very similar one to the Dream Router/UDR). If you want to adapt any of this to a UniFi router product that's not the UDM SE, please be careful, check your filesystem, and act accordingly.

Install [boostchicken's `on-boot-script`](https://github.com/boostchicken-dev/udm-utilities/tree/master/on-boot-script) to allow your UDM SE to run software that persists across reboots and firmware upgrades.

Copy a Linux arm64 release to your UDM SE and place it in `/mnt/data/on_boot.d/binaries`, then use the following script to symlink it and the systemd service config into the right place on boot.

Fix up the `MBDNS_SRC_BINARY` path as needed if you're not using a `release-1.0.2` binary. Use a different number if you have more going on in your `on_boot.d` script setup, if needed. The system loads them in numerical order.

## `/mnt/data/on_boot.d/25-mbdns.sh`

```
#!/bin/sh

MBDNS_SRC_BINARY=/mnt/data/on_boot.d/binaries/mbdns-release-1.0.2-linux-arm64
MBDNS_BINARY=/usr/local/bin/mbdns
MBDNS_SRC_SYSTEMD_SERVICE=/mnt/data/on_boot.d/settings/mbdns/mbdns.service
MBDNS_SYSTEMD_SERVICE=/etc/systemd/system/mbdns.service

if ! test -f $MBDNS_BINARY; then
    ln -s $MBDNS_SRC_BINARY $MBDNS_BINARY
fi

if ! test -f $MBDNS_SYSTEMD_SERVICE; then
    ln -s $MBDNS_SRC_SYSTEMD_SERVICE $MBDNS_SYSTEMD_SERVICE
fi

systemctl start mbdns
```

## `/mnt/data/on_boot.d/settings/mbdns/mbdns.service`

```
[Unit]
Description=mbdns
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mbdns -config /mnt/data/on_boot.d/settings/mbdns/mbdns.conf

[Install]
WantedBy=multi-user.target
```

## Configuring mbdns

Configure it as you would for any other platform. Put the config in `/mnt/data/on_boot.d/settings/mbdns` and remember to `chmod 0400` it.

## Upgrading mbdns

Put a new release in the same `binaries` folder, update the `25-mbdns.sh` script, `rm -f /usr/local/bin/mbdns` and re-run the script to recreate the symlink to the new version.
