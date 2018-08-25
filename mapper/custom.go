package mapper

type Custom interface {
	Read([]byte)
	Write() []byte
}