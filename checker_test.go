package buildmeta

import (
	"net/http"
	"net/http/httptest"
	"testing"

	health "astuart.co/go-healthcheck"
	"github.com/stretchr/testify/assert"
)

func TestChecker(t *testing.T) {
	asrt := assert.New(t)
	h := Handler()

	var retErr error
	h.Ready.Register("foo", health.CheckFunc(func() error {
		return retErr
	}))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}))
	defer ts.Close()

	ch := MetaChecker{Root: ts.URL}

	asrt.NoError(ch.Check())

	retErr = &DownstreamError{
		StatusCode: 500,
	}

	err := ch.Check()
	asrt.Error(err)
	asrt.IsType((*DownstreamError)(nil), err)

	d := err.(*DownstreamError)
	asrt.Len(d.Body, 1)
}
