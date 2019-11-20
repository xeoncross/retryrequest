package retryrequest

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// CustomRetryPolicy will retry on certain connection and server errors
// Based on https://github.com/hashicorp/go-retryablehttp/blob/f1bc72b7b3c24d61ec70f911dbe703af3ea67df2/client.go#L356-L395
// func CustomRetryPolicy(ctx context.Context, resp *http.Response, err error) bool {
// 	// do not retry on context.Canceled or context.DeadlineExceeded
// 	if ctx.Err() != nil {
// 		return false
// 	}
//
// 	if err != nil {
// 		if v, ok := err.(*url.Error); ok {
// 			// Don't retry if the error was due to TLS cert verification failure.
// 			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
// 				return false
// 			}
// 		}
//
// 		// If a timeout
// 		// if ne, ok := err.(net.Error); ok && ne.Timeout() {
// 		// 	return true
// 		// }
//
// 		// The error is likely recoverable so retry.
// 		return true
// 	}
//
// 	// Check the response code. We retry on 500-range responses to allow
// 	// the server time to recover, as 500's are typically not permanent
// 	// errors and may relate to outages on the server side. This will catch
// 	// invalid response codes as well, like 0 and 999.
// 	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != 501) {
// 		return true
// 	}
//
// 	return false
// }

// Do HTTP request, retrying if server times out or sends a 500 error
func Do(client *http.Client, req *http.Request, attempts int, delay time.Duration) (*http.Response, error) {

	var err error
	var resp *http.Response

	for i := 0; i < attempts; i++ {

		resp, err = client.Do(req)

		shouldRetry, _ := retryablehttp.DefaultRetryPolicy(req.Context(), resp, err)

		if !shouldRetry {
			if err != nil {
				if resp.Body != nil {
					resp.Body.Close()
				}
				return nil, err
			}

			return resp, nil
		}

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
		}

	}

	return nil, err
}
