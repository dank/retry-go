/*
Simple library for retry mechanism

slightly inspired by [Try::Tiny::Retry](https://metacpan.org/pod/Try::Tiny::Retry)

SYNOPSIS

http get with retry:

	url := "http://example.com"
	var body []byte

	err := retry.Do(
		func() error {
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			return nil
		},
	)

	fmt.Println(body)

[next examples](https://github.com/avast/retry-go/tree/master/examples)


SEE ALSO

* [giantswarm/retry-go](https://github.com/giantswarm/retry-go) - slightly complicated interface.

* [sethgrid/pester](https://github.com/sethgrid/pester) - only http retry for http calls with retries and backoff

* [cenkalti/backoff](https://github.com/cenkalti/backoff) - Go port of the exponential backoff algorithm from Google's HTTP Client Library for Java. Really complicated interface.

* [rafaeljesus/retry-go](https://github.com/rafaeljesus/retry-go) - looks good, slightly similar as this package, don't have 'simple' `Retry` method

* [matryer/try](https://github.com/matryer/try) - very popular package, nonintuitive interface (for me)

BREAKING CHANGES

0.3.0 -> 1.0.0

* `retry.Retry` function are changed to `retry.Do` function

* `retry.RetryCustom` (OnRetry) and `retry.RetryCustomWithOpts` functions are now implement via functions produces Options (aka `retry.OnRetry`)


*/
package retry

import (
	"time"
)

// Function signature of retryable function
type RetryableFunc func() error

func Do(retryableFunc RetryableFunc, opts ...Option) error {
	var n uint

	//default
	config := &config{
		attempts: 10,
		delay:    100,
		units:    time.Millisecond,
		onRetry:  func(n uint, err error) {},
		retryIf:  func(err error) bool { return true },
	}

	//apply opts
	for _, opt := range opts {
		opt(config)
	}

	errorLog := make(Error, 0)

	cond := n < config.attempts
	if n == 0 {
		cond = true
	}

	for cond {
		err := retryableFunc()

		if err != nil {
			config.onRetry(n, err)
			errorLog = append(errorLog, err)

			if !config.retryIf(err) {
				break
			}

			// if this is last attempt - don't wait
			if n == config.attempts-1 {
				break
			}

			time.Sleep((time.Duration)(config.delay) * config.units)
		} else {
			return nil
		}

		n++
	}

	return errorLog
}

// Error type represents list of errors in retry
type Error []error

// Error method return string representation of Error
// It is an implementation of error interface
func (e Error) Error() string {
	return e[len(e) - 1].Error()
}
