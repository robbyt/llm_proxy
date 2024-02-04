# LLM Proxy
This is a very simple proxy server that can be used to intercept and modify requests to the OpenAI.
It is heavily based on [github.com/kardianos/mitmproxy](kardianos/mitmproxy) which is based on
[mitmproxy](https://mitmproxy.org/).

## How to use

1. Install Go
2. Run `go install github.com/robbyt/llm_proxy@latest`
3. Run `llm_proxy` with a few command line options set
4. Set your HTTP_PROXY environment variable to `http://localhost:8080`
5. Use the OpenAI API as you normally would
    
```bash
$ llm_proxy daemon --verbose --ca_dir ~/.mitmproxy
```

## Why is this useful?

For my personal use case, I want to save all requests and responses to the OpenAI API, for purposes
of creating Fine Tuning datasets. Read more about this in the
[OpenAI API documentation](https://platform.openai.com/docs/api-reference/fine-tuning/).

Other possible uses include:
* Security and auditing
* Debugging
* DMZ for internal services
* Mocking API responses for testing (feature pending...)

## TLS / HTTPs Support

The fastest way to use this proxy is to send requests to `http://api.openai.com`, and this proxy
will upgrade and MITM those requests/responses. This is fine if you are connecting to this proxy
over  `localhost` or a secure network, in this case the flow would be: 
`http://api.openai.com->llm_proxy->https://api.openai.com`

This is *fine* for connecting via `localhost` because local traffic on a single computer does not
need to be encrypted, and traffic leaving the proxy will be encrypted and the upstream server's
certificate will be validated by this proxy.

However, if you need to send requests from your application to `https://api.openai.com`, you will
need to add a trust for the self-signed cert this app generates (or generate/trust your own cert.)

By default, this proxy will generate a certificate at `~/.mitmproxy/mitmproxy-ca-cert.pem`.
If you want to use a different directory, use the `-ca_dir` flag when starting this proxy daemon.

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