# diskled: a basic app that scans /proc/diskstats and makes lights blink.
---
syntax: ./diskled --interval=<interval in milliseconds to scan> --disks=sda:496,sdb:498  
```
--disks=<disk in /proc/diskstats to monitor, ':' gpio id to set value to 1 on disk access.'>  
```
<br>
diskled.service file to install for bootup, if diskled is put in /usr/local/sbin:

```
[Unit]
Description=Drive LED

[Service]
Type=simple
User=root
ExecStart=/usr/local/sbin/diskled --interval=300 --disks=sda:436,sdb:438
Restart=on-failure

[Install]
WantedBy=default.target
```
