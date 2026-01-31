package limiter

// MockRecorder captures metrics in memory for assertion
type MockRecorder struct {
	Counters map[string]float64
	Timings  map[string][]float64
}

func NewMockRecorder() *MockRecorder {
	return &MockRecorder{
		Counters: make(map[string]float64),
		Timings:  make(map[string][]float64),
	}
}

func (m *MockRecorder) Add(name string, value float64, tags map[string]string) {
	m.Counters[name] += value
}

func (m *MockRecorder) Observe(name string, value float64, tags map[string]string) {
	m.Timings[name] = append(m.Timings[name], value)
}
