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
sudo setcap cap_net_raw=eip ddump-worker

# Run
./ddump-worker

# Query
curl http://localhost:8475/capture -vo ./full-capture.pcap
curl http://localhost:8475/capture?interface=eth0&filter=tcp%20port%2080 -vo ./ethernet-http-capture.pcap
curl http://localhost:8475/capture\?filter\=icmp -sN | tcpdump -vnl -r-
```

Reference
---------

Settings
++++++++

* ``-listen``: Address to listen on. May be an "<address>:<port>" or "unix:<path>". (default: :8475)

API
+++

``GET /capture``

Start a network capture.

**Request parameters**:

* ``interface``: Name of the interface to capure on. (default: any)
* ``filter``: Tcpdump BPF filter. (default: no filter)

**Response**:

* ``Content-Type``: application/vnd.tcpdump.pcap
* Body: Raw pcap stream
