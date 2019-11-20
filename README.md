## Retry Request (Go)

Simple package to retry http.Client requests.

---

Recommendation: use [github.com/hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) instead

`hashicorp/go-retryablehttp` is too heavy of a wrapper for me (though it provides more features). This is a lightweight wrapper that is intended to be used to repeat the same request X times for endpoints that are often overloaded. `hashicorp/go-retryablehttp` also [provides it's own internal pool](https://github.com/hashicorp/go-retryablehttp/blob/master/client.go#L326) of [github.com/hashicorp/go-cleanhttp](https://github.com/hashicorp/go-cleanhttp) clients by default.
