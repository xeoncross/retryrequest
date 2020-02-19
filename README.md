## Retry Request (Go)

Simple package to retry http.Client requests when:

- Status code is invalid/missing
- Status code is 500 range (except 501)
- Connection timeout reached

This package respects the request `context.Context`. Requests will not retry if context has been canceled or `DeadlineExceeded` is reached.

This package mostly is a drop-in replacement to help with the common [net.Error.Timeout()](https://pkg.go.dev/net?tab=doc#Error) when calling TCP, API endpoints.

Usage:

Normal requests are performed the following way:

```go
resp, err := client.Do(req)
```

To use this library simply call `retryrequest.Do` instead:

```go
resp, err := retryrequest.Do(client, req, nil)
```

The third argument is the optional retry policy:

```go
var DefaultPolicy = &Policy{
	Retry500Status:     true,
	RetryInvalidStatus: true,
	Attempts:           2,
	Delay:              time.Millisecond * 500,
}
```


---

Recommendation: use [github.com/hashicorp/go-retryablehttp](https://github.com/hashicorp/go-retryablehttp) for more types of retry scenarios.

`hashicorp/go-retryablehttp` is too heavy of a wrapper for me (though it provides more features). This is a lightweight wrapper that is intended to be used to repeat the same request X times for endpoints that are often overloaded. `hashicorp/go-retryablehttp` also [provides it's own internal pool](https://github.com/hashicorp/go-retryablehttp/blob/master/client.go#L326) of [github.com/hashicorp/go-cleanhttp](https://github.com/hashicorp/go-cleanhttp) clients by default.
