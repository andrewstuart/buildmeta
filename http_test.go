package buildmeta

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServeInfo(t *testing.T) {
	assert := assert.New(t)

	gitCommit = "commit"
	gitCommitTime = "now"
	gitTag = "tag"
	buildTime = "tomorrow"

	h := Handler()

	req := httptest.NewRequest("GET", "/info", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	var info Info
	_ = json.Unmarshal(body, &info)

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal(Info{
		GitCommit:     gitCommit,
		GitCommitTime: gitCommitTime,
		GitRepo:       gitRepo,
		GitTag:        gitTag,
		BuildTime:     buildTime,
	}, info)
}

func TestNoPathFound(t *testing.T) {
	assert := assert.New(t)

	h := Handler()

	req := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusNotFound, resp.StatusCode)
}
