# Basic architecture of mbdns

- [x] Read JSON-encoded `[domain token host ttl]` record tuples from a (checked) `0400` file called `mbdns.conf` that lives next to the binary
- [x] Iterate over each tuple and:
    - [x] https `POST` `"domain=t.domain;password=t.token;command=REPLACE t.host t.ttl A DYNAMIC_IP"` to `dnsapi.mythic-beasts.com`
    - [x] Record success and failure but don't handle failure. Just try again later next time around the loop.
- [x] Do that in loop with a compile-time loop sleep in a background thread/goroutine, running daemonised
- [x] Some basic logging to stdout

# JSON file format

Because of the way golang does magic unmarshalling into structs, our golang code self-documents the JSON format.

```go
type record struct {
	Domain string
	Token  string
	Host   string
    TTL    string
    Record string
}
```

Is unmarshalled from:

```json
[
    {
        "domain" : "some_domain",
        "token"  : "mythic_beasts_api_token",
        "host"   : "hostname",
        "ttl"    : "3600",
        "record" : "A"
    }
]
```

# Communicating with the Mythic Beasts Primary DNS API

Documentation: [here](https://www.mythic-beasts.com/support/api/primary)

API URL: `https://dnsapi.mythic-beasts.com/`

`GET` is supported but we'll use `POST`. `POST` needs an `application/x-www-form-urlencoded`, which we get from [`net/http`](https://golang.org/pkg/net/http/#pkg-overview)'s `PostForm()` API.

Building the payload is easy:

`http.PostForm(mythicbeastsUrl, url.Values{"key": {"value"}, "key2": {"value2"}} ...)`

So we'll do:

`http.PostForm(mbUrl, url.Values{"domain": {record.Domain}, "password": {record.Token}, "command": {builtCommand}})`

We use the `REPLACE` command and build the rest of the `builtCommand` payload with a `Sprintf()` formatted const string.

`REPLACE` takes the form `REPLACE host TTL record DYNAMIC_IP`

We get HTTP response code 200 (OK) back on success, 4xx in the event of a failure.