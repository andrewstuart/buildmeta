package buildmeta_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	health "astuart.co/go-healthcheck"
	"github.com/andrewstuart/buildmeta"
)

var db *sql.DB

func ExampleHandler() {
	h := buildmeta.Handler()

	h.Alive.Register("postgres", health.PeriodicChecker(health.CheckFunc(func() error {
		return db.Ping()
	}), time.Minute))

	h.Alive.Register("someservie", health.PeriodicChecker(buildmeta.MetaChecker{
		Root: "https://some-buildmeta-using-service.test.local:8443",
	}, time.Minute))

	h.Ready.Register("random", health.PeriodicThresholdChecker(health.CheckFunc(func() error {
		if rand.Int()%2 == 0 {
			// If we get 5 coin tosses in a minute then return not ready
			return fmt.Errorf("not ready")
		}
		return nil
	}), time.Minute, 5))

	http.Handle(buildmeta.BasePath, h)
	http.ListenAndServe("localhost:8080", nil)
}

func ExampleHandlerNonStandardPath() {
	http.Handle("/manage/", buildmeta.Handler())
	http.ListenAndServe("localhost:8080", nil)
}
