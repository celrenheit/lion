package htest

import "bytes"

type buffer struct {
	*bytes.Buffer
}

func (*buffer) Close() error {
	return nil
}

func newBody(buf []byte) *buffer {
	return &buffer{bytes.NewBuffer(buf)}
}

func newBodyString(buf string) *buffer {
	return &buffer{bytes.NewBufferString(buf)}
}
