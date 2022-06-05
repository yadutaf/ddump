Distributed tcpdump
###################

``ddump`` is a small utility to run multiple network captures on multiple targets
and live-merging the captured streams. This is usefull when investigating network
issues in a distributed system.

The main process runs as a command line program and outputs the merged stream on
stdout for easy integration.

The worker processes run as a long running HTTP server, suitable for integration
behind a streaming (i.e. non buffering) HTTP reverse proxy. Authentication shall
be handled by the proxy, if any.

Dependencies
------------

Ubuntu:

```
sudo apt install libpcap-dev
```

Usage
-----

Build all:

```
# Build
make
```

Run & test the worker:

```
# Allow running without root
sudo setcap cap_net_raw=eip ddump-worker

# Run
./ddump-worker

# Query 
curl http://localhost:8475/capture -vo ./full-capture.pcap
curl http://localhost:8475/capture?interface=eth0&filter=tcp%20port%2080 -vo ./ethernet-http-capture.pcap
curl http://localhost:8475/capture\?filter\=icmp -sN | tcpdump -vnl -r-
```

Run & test the CLI:

```
./ddump \
    "http://0.0.0.0:8475/capture?filter=icmp%20and%20host%201.1.1.1" \
    "http://0.0.0.0:8475/capture?filter=icmp%20and%20host%208.8.8.8" \
    | tcpdump -vnlr-
```

The command above will start 2 capture streams on a locally running worker:

* Ping to/from 1.1.1.1
* Ping to/from 8.8.8.8

The command then outputs to stdout the merged capture stream in a format suitable
for ingestion by ``tcpdump`` (as demonstrated in the example).

CLI reference
-------------

Command line
++++++++++++

The command line is of the form:

```
./ddump TARGET_URL_1 TARGET_URL_2 ... TARGET_URL_N 
```

Where target URLs are of the form `PROTOCOL://[USER[:PASSWORD]@]HOST[:PORT(8475)][?filter=FILTER]`:

* ``PROTOCOL``: One of ``http``, ``https``
* ``USER``: Optional username, if going through and authenticating reverse proxy
* ``PASSWORD``: Optional password, if going through and authenticating reverse proxy
* ``HOST``: IP address or domaine name of the target capture agent
* ``PORT``: Optional tcp port, if the capture agent listens on a non-standard port
* ``FILTER``: Optional urlencoded capture filter

Output
++++++

The main output stream is the merged pcap stream. Messages an errors are sent to stderr.

Worker Reference
----------------

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
