package networkestop

import (
	"context"
	"sync"
	"time"
	
	"github.com/edaniels/golog"
	
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils"	
)

var family = resource.ModelNamespace("erh").WithFamily("networkestop")

var DnsModel = family.WithModel("dns")

func init() {
	resource.RegisterComponent(
		sensor.API,
		DnsModel,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newDnsSensor,
		})
}

type Config struct {
	Server string
	Lookup string
	Stop []string
	IntervalMS int `json:"interval_ms"`
	TimeoutMS int `json:"timeout_ms"`
}

func (cfg *Config) Validate(path string) ([]string, error) {
	if cfg.Server == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "host")
	}

	if cfg.Lookup == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "lookup")
	}

	if len(cfg.Stop) == 0 {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "stop")
	}

	if cfg.IntervalMS <= 0 {
		cfg.IntervalMS = 2000
	}

	if cfg.TimeoutMS <= 0 {
		cfg.TimeoutMS = 1000
	}

	return cfg.Stop, nil
}

func newDnsSensor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger golog.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	s := &dnsSensor{
		name: rawConf.ResourceName(),
		logger: logger,
		cfg: conf,
	}

	for n, r := range deps {
		a, ok := r.(resource.Actuator)
		if !ok {
			return fmt.Errorf("%v is not an actuator", n)
		}
		s.actuators = append(s.actuators, a)
	}
	
	s.start()
	
	return s, nil
}

type dnsSensor struct {
	resource.AlwaysRebuild

	name    resource.Name
	logger golog.Logger
	cfg *Config

	actuators []resource.Actuator
	
	cancel context.CancelFunc
	
	mu sync.Mutex
	
	lastAttempt time.Time
	lastSuccess time.Time
	lastSuccessTimeMS int64
	lastError error
}

func (ps *dnsSensor) start() {
	ctx := context.Background()
	ctx, ps.cancel = context.WithCancel(ctx)

	go func() {
		for utils.SelectContextOrWait(ctx, time.Duration(ps.cfg.IntervalMS) * time.Millisecond) {
			start := time.Now()
			err := doDns(ctx, ps.cfg.Server, ps.cfg.Lookup)
			end := time.Now()

			ps.mu.Lock()
			ps.lastAttempt = start
			
			if err == nil {
				ps.lastSuccess = start
				ps.lastSuccessTimeMS = end.Sub(start).Milliseconds()
			} else {
				ps.stopComponents(ctx)
			}
			
			ps.lastError = err
			ps.mu.Unlock()
		}
	}()

}

func doDns(ctx context.Context, server, lookup string) error {
	panic(1)
}

func (ps *dnsSensor) stopComponents(ctx context.Context) {
	for _, a := range ps.actuators {
		err := a.Stop(ctx, nil)
		if err != nil {
			ps.logger.Warnf("cannot stop actuator %v", err)
		}
	}
}

func (ps *dnsSensor) Close(ctx context.Context) error {
	ps.cancel()
	return nil
}

func (ps *dnsSensor) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func (ps *dnsSensor) Name() resource.Name {
	return ps.name
}

func (ps *dnsSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	
	return map[string]interface{}{
		"last_attempt" : ps.lastAttempt,
		"last_success" : ps.lastSuccess,
		"last_success_latency_ms" : ps.lastSuccessTimeMS,
		"last_error" : ps.lastError,
	}, nil
}
