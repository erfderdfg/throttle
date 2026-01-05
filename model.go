package limiter

type Namespace string

// Identity uniquely identifies the subject being rate-limited.
type Identity struct {
	Namespace Namespace
	Key       string
}
