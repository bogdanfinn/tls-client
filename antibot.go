package tls_client

import (
	http "github.com/bogdanfinn/fhttp"
)

type CFSolvingHandler interface {
	Solve(logger Logger, client HttpClient, response *http.Response) (*http.Response, error)
}
