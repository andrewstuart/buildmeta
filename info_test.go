package buildmeta

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestGenerateInfo(t *testing.T) {
	asrt := assert.New(t)
	i, err := GenerateInfo(".")

	asrt.NoError(err)
	asrt.NotEqual(Info{}, i)
}

func TestProm(t *testing.T) {
	asrt := assert.New(t)

	i := Info{
		GitTag: "foo",
	}

	coll := i.PrometheusCollector()
	ch := make(chan prometheus.Metric)
	go coll.Collect(ch)

	m := <-ch
	asrt.NotNil(m.Desc())

	var met dto.Metric
	asrt.NoError(m.Write(&met))
	asrt.Equal(1.0, *met.Gauge.Value)
}

func TestGetInfoDontLeakCreds(t *testing.T) {
	asrt := assert.New(t)
	gitRepo = "https://user:mypassword@example.com/myrepo.git"
	info := GetInfo()

	asrt.Equal("https://example.com/myrepo.git", info.GitRepo, "error: credentials were leaked by a call GetInfo made")
}

func TestFlags(t *testing.T) {
	asrt := assert.New(t)

	i := Info{
		GitTag: "foobar",
	}

	asrt.Contains(i.LDFlags(), "foobar")
}
