package networkestop

import (
	"context"
	"testing"

	"go.viam.com/test"

	"go.viam.com/rdk/resource"
)

func TestDNS(t *testing.T) {
	err := doDns(context.Background(), "8.8.8.8", "www.viam.com.")
	test.That(t, err, test.ShouldBeNil)
}

func TestSuccess(t *testing.T) {
	cfg := Config{
		Server: "8.8.8.8",
		Lookup: "www.viam.com.",
		Stop:   []string{"mm"},
	}
	req, err := cfg.Validate("x")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, 1, test.ShouldEqual, len(req))

	ps := dnsSensor{cfg: &cfg}
	ps.doLoop(context.Background())

	r, err := ps.Readings(context.Background(), nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, r["last_error"], test.ShouldBeNil)
	test.That(t, r["last_success_latency_ms"], test.ShouldBeGreaterThan, 0)
	test.That(t, r["last_success_latency_ms"], test.ShouldBeLessThan, 2000)
}

type fakeA struct {
	called int
}

func (a *fakeA) IsMoving(context.Context) (bool, error) {
	return false, nil
}

func (a *fakeA) Stop(context.Context, map[string]interface{}) error {
	a.called++
	return nil
}

func TestFail(t *testing.T) {
	cfg := Config{
		Server: "127.0.0.1:55555",
		Lookup: "www.viam.com.",
		Stop:   []string{"mm"},
	}
	req, err := cfg.Validate("x")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, 1, test.ShouldEqual, len(req))

	a := &fakeA{}
	ps := dnsSensor{cfg: &cfg, actuators: []resource.Actuator{a}}

	ps.doLoop(context.Background())

	r, err := ps.Readings(context.Background(), nil)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, r["last_error"], test.ShouldNotBeNil)

	test.That(t, a.called, test.ShouldEqual, 1)
}
