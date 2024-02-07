# LLM Proxy
This is a simple proxy server designed to intercept and modify requests to OpenAI.
It is heavily based on [kardianos/mitmproxy](https://www.github.com/kardianos/mitmproxy) which is based on
[mitmproxy](https://mitmproxy.org/).

## How to use

1. Install Go
2. Run `go install github.com/robbyt/llm_proxy@latest`
3. Golang will download and compile this code, storing the binary in your `$GOPATH/bin` directory.
4. By default this will compile the binary to `~/go/bin/llm_proxy`, try running `llm_proxy --help`
5. Set your `HTTP_PROXY` and `HTTPS_PROXY` environment variables to `http://localhost:8080`
6. Use the OpenAI API as you normally would (but, you may need to adjust this for your app.)

### Running the proxy server
```bash
$ llm_proxy dir_logger --verbose
```

### Using cURL to query, and use the proxy
```bash
$ OPENAI_API_KEY=sk-XXXXXXX curl \
    -x http://localhost:8080 \
    -X GET \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    http://api.openai.com/v1/models
```
Note: this example sends the request to `http://api.openai.com/...` instead of `https://`. This is
because the proxy will perform the SSL termination, and will upgrade the outbound request to
`https://` when it sends the request to the upstream API server. There is more information about
this in the TLS section below.

### Using the proxy with the OpenAI Python client

In this example, we are changing the `base_url` to connect via `http` so the proxy can MITM the
connection without needing to add a self-signed cert to the Python client.

```python
import httpx
from openai import OpenAI

proxies = {
    "http://": "http://localhost:8080",
    "https://": "http://localhost:8080",
}

client = OpenAI(
    # max_retries=0,
    base_url="http://api.openai.com/v1",
    http_client=httpx.Client(
        proxies=proxies,
    ),
)

chat_completion = client.chat.completions.create(
    messages=[
        {
            "role": "user",
            "content": "Hello, you are amazing.",
        }
    ],
    model="gpt-3.5-turbo",
)
```
More info here: [httpx proxy config](https://www.python-httpx.org/advanced/#client-instances)

## Why is this proxy useful?

For my personal use case, I want to save all requests and responses to the OpenAI API, to enable
__Fine Tuning__ models by sending these conversations to the OpenAI API. Read more about this in
the [OpenAI API documentation](https://platform.openai.com/docs/api-reference/fine-tuning/).

Other possible uses include:
* Security and auditing
* Debugging
* DMZ for internal services
* Mocking API responses for testing (feature pending...)

## TLS / HTTPs Support

When you send requests to this proxy directed at `http://api.openai.com`, the proxy will upgrade
those requests to `https://api.openai.com` when it sends the request to the upstream server.

However, if you must requests from your application to `https://api.openai.com`, you will need to
add the self-signed cert to your trust store, or disable TLS validation in your client (not
recommended).

This proxy will generate a certificate at `~/.mitmproxy/mitmproxy-ca-cert.pem`, but if you want to
use a different directory, use the `-ca_dir` flag when starting this proxy daemon.

More info here on self-signed certs and MITM:
[https://docs.mitmproxy.org/stable/concepts-certificates/]


For reference, here's how you can generate and trust a self-signed cert in MacOS:
```bash
# Create a directory for the cert files
$ mkdir -p ~/.mitmproxy
$ cd ~/.mitmproxy

# you only need to generate this cert if you do not want the llm_proxy to generate it for you
$ openssl genrsa -out mitmproxy-ca-cert.key 2048
# this self-signed cert expires in 10 years, and I hope you are using something else by that point
$ openssl req -x509 -new -nodes -key mitmproxy-ca-cert.key -sha256 -days 3650 -out mitmproxy-ca-cert.pem

# Trust the CA
$ sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain mitmproxy-ca-cert.pem
```

Using curl with proxy and self-signed cert: 
(Set your OpenAI API key as the environment variable prefix!)
```bash
$ OPENAI_API_KEY=sk-XXXXXXX curl \
    -x http://localhost:8080 \
    --cacert ~/.mitmproxy/mitmproxy-ca.pem \
    -X GET \
    -H "Authorization: Bearer $OPENAI_API_KEY" \
    https://api.openai.com/v1/models
```