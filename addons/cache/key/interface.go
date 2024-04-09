package key

type Key interface {
	Get() []byte
	String() string
}
