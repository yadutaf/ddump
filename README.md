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

* ``filter``: Tcpdump BPF filter. (default: no filter)

**Note**: There is no "interface" parameter. This is intentional. Indeed, to merge
multiple pcap streams, all streams must have the same type. This implies the use
of a generic link layer. This layer could be one of Linux's ``SLL`` or ``SLL2``.
While SLL2 would have allowed per-interface filters, it is not supported by the
underlying gopacket library.

**Response**:

* ``Content-Type``: application/vnd.tcpdump.pcap
* Body: Raw pcap stream
