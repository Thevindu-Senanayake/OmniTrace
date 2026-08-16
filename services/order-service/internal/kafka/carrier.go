package kafka

import kafkago "github.com/segmentio/kafka-go"

// headerCarrier adapts a kafka-go message's headers to OTel's TextMapCarrier so
// the W3C traceparent can be injected on publish and extracted on consume. This
// is the only place the project hand-wires trace propagation — the Kafka hop is
// what stitches order-service to inventory (and payment back to order-service),
// and unlike the HTTP hops there is no off-the-shelf instrumentation for the
// segmentio/kafka-go client.
//
// It holds a pointer to the header slice so Set can grow it in place.
type headerCarrier struct {
	headers *[]kafkago.Header
}

func (c headerCarrier) Get(key string) string {
	for _, h := range *c.headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c headerCarrier) Set(key, value string) {
	// W3C keys are unique, so replace an existing header rather than duplicate it.
	for i, h := range *c.headers {
		if h.Key == key {
			(*c.headers)[i].Value = []byte(value)
			return
		}
	}
	*c.headers = append(*c.headers, kafkago.Header{Key: key, Value: []byte(value)})
}

func (c headerCarrier) Keys() []string {
	keys := make([]string, len(*c.headers))
	for i, h := range *c.headers {
		keys[i] = h.Key
	}
	return keys
}
