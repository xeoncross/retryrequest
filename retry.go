package retryrequest

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Policy for retrying requests
type Policy struct {
	Retry500Status     bool
	RetryInvalidStatus bool
	Attempts           int
	Delay              time.Duration
}

// DefaultPolicy used when a policy is not provided
var DefaultPolicy = &Policy{
	Retry500Status:     true,
	RetryInvalidStatus: true,
	Attempts:           2,
	Delay:              time.Millisecond * 500,
}

// Do HTTP request, retrying if server times out or sends a 500 error
func Do(client *http.Client, req *http.Request, policy *Policy) (*http.Response, error) {

	if policy == nil {
		policy = DefaultPolicy
	}

	var err error
	var resp *http.Response

	for i := 0; i < policy.Attempts; i++ {

		// After first attempt, response might be populated
		// Close to prevent memory leak
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		resp, err = client.Do(req)

		shouldRetry := checkRetryPolicy(req.Context(), resp, err, *policy)

		if !shouldRetry {
			break
		}

		select {
		case <-req.Context().Done():
			err = req.Context().Err()
			break
		case <-time.After(policy.Delay):
		}

	}

	return resp, err
}

// customRetryPolicy will retry on certain connection and server errors
// TODO considering moving to a whitelist approach with options provided by caller
// Based on https://github.com/hashicorp/go-retryablehttp/blob/f1bc72b7b3c24d61ec70f911dbe703af3ea67df2/client.go#L356-L395
func checkRetryPolicy(ctx context.Context, resp *http.Response, err error, policy Policy) bool {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false
	}

	// If a timeout
	if err != nil {
		if ne, ok := err.(net.Error); ok && ne.Timeout() {
			return true
		}

		return false
	}

	// Check the response code. We retry on 500-range responses to allow
	// the server time to recover, as 500's are typically not permanent
	// errors and may relate to outages on the server side.
	if policy.Retry500Status && (resp.StatusCode >= 500 && resp.StatusCode < 600 && resp.StatusCode != 501) {
		return true
	}

	if policy.RetryInvalidStatus && (resp.StatusCode == 0 || resp.StatusCode >= 599) {
		return true
	}

	return false
}
