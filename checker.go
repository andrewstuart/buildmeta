package buildmeta

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	health "astuart.co/go-healthcheck"
)

// MetaChecker is a health.Checker implementation that just abstracts over a
// simple base endpoint, and attempts to contact the buildmeta readiness probe
// under the well-known buildmeta path.
type MetaChecker struct {
	Root string
}

var _ health.Checker = &MetaChecker{}

// Check implements health.Checker
func (g MetaChecker) Check() (err error) {
	res, err := http.Get(strings.TrimRight(g.Root, "/") + BasePath + ReadyPath)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		var d DownstreamError
		json.NewDecoder(res.Body).Decode(&d.Body) // error can be ignored; zero value is fine
		return &d
	}

	return nil
}

// A DownstreamError is returned by the MetaChecker so that further details
// can be inspected in the output messages if it is json encoded.
type DownstreamError struct {
	StatusCode   int
	Root, Status string
	Body         interface{}
}

// Error implements error
func (d *DownstreamError) Error() string {
	return fmt.Sprintf("%q gave response code: %d status: %q", d.Root, d.StatusCode, d.Status)
}
