package telemetry

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Define our labels here so that we can easily reuse them.
var defaultLabels = []attribute.KeyValue{
	attribute.Key("application").String(ServiceName),
	attribute.Key("container_id").String(os.Getenv("HOSTNAME")),
}

func newCounter(mc metric.Int64Counter) *MetricCounter {
	attributes := append([]attribute.KeyValue{}, defaultLabels...)
	return &MetricCounter{attributes, mc, 1}
}

type MetricCounter struct {
	attributes []attribute.KeyValue
	counter    metric.Int64Counter
	value      int64
}

func (c *MetricCounter) Commit(ctx context.Context) {
	c.counter.Add(ctx, c.value, c.attributes...)
}

func (c *MetricCounter) WithAttribute(key string, value interface{}) *MetricCounter {
	switch v := value.(type) {
	case bool:
		c.attributes = append(c.attributes, attribute.Key(key).Bool(v))

	case int:
		c.attributes = append(c.attributes, attribute.Key(key).Int(v))

	case int64:
		c.attributes = append(c.attributes, attribute.Key(key).Int64(v))

	case float64:
		c.attributes = append(c.attributes, attribute.Key(key).Float64(v))

	case string:
		c.attributes = append(c.attributes, attribute.Key(key).String(v))

	case []string:
		c.attributes = append(c.attributes, attribute.Key(key).Array(v))

	default:
		panic(fmt.Sprintf("unsupported values type: %T", value))
	}

	return c
}

func (c *MetricCounter) WithEntryFields(e log.Entry) *MetricCounter {
	for k, v := range e.Data {
		c.WithAttribute(k, v)
	}

	return c
}

func (c *MetricCounter) WithError(err error) *MetricCounter {
	return c.WithAttribute("error", err.Error())
}

func (c *MetricCounter) WithValue(value int64) *MetricCounter {
	c.value = value
	return c
}
