package harness

type API int

// TODO: Auto-generate this list from the target Kafka version
const (
	Metadata API = iota
	Produce
	Fetch
	CreateTopic
	DeleteTopic
	Binary
)
