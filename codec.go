package zhanio

type Codec interface {
	Encode(c Conn, buf []byte) ([]byte, error)
	Decode(c Conn) ([]byte, error)
}

type defaultCodec struct {
}

func (codec *defaultCodec) Encode(c Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (codec *defaultCodec) Decode(c Conn) ([]byte, error) {
	return c.Read(), nil
}
