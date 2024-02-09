package megadumper

type MegaDumperWriter interface {
	Read() ([]byte, error)
}
