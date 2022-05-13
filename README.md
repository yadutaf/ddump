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

# Query
curl http://localhost:8475/capture -vo ./full-capture.pcap
curl http://localhost:8475/capture?interface=eth0&filter=tcp%20port%2080 -vo ./ethernet-http-capture.pcap
```

Reference
---------

Settings
++++++++

* ``-listen``: Address and port to listen on. (default: :8475)

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
