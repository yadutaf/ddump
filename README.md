Distributed tcpdump
###################

Dependencies
------------

Ubuntu:

```
sudo apt install libpcap-dev
```

Usage
-----

```
# Build
make

# Allow running without root
sudo setcap cap_net_raw=eip distributed-tcpdump

# Run
./distributed-tcpdump
```

