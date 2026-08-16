package trace

const TraceName = "mxshop"

type Options struct {
	Name     string  `json:"name"`
	Endpoint string  `json:"endpoint"`
	Sampler  float64 `json:"sampler"`
	Batcher  string  `json:"batcher"`
}
