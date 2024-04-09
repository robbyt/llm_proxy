package addons

type LLM_Addon interface {
	String() string
	Close() error
}
