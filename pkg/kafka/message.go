package kafka

type Header struct {
	Key   string
	Value []byte
}

type Message struct {
	Key     []byte
	Value   []byte
	Headers []Header

	partition int
	offset int64
}
