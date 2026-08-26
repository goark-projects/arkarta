package multipart

import "io"

type limitReadCloser struct {
	reader    io.ReadCloser
	remaining int64
}

func (r *limitReadCloser) Read(buffer []byte) (int, error) {
	if r.remaining < 0 {
		return r.reader.Read(buffer)
	}
	if r.remaining == 0 {
		var probe [1]byte
		n, err := r.reader.Read(probe[:])
		if n > 0 {
			return 0, ErrBodyTooLarge
		}
		return 0, err
	}
	if int64(len(buffer)) > r.remaining+1 {
		buffer = buffer[:r.remaining+1]
	}
	n, err := r.reader.Read(buffer)
	if int64(n) > r.remaining {
		allowed := int(r.remaining)
		r.remaining = 0
		return allowed, ErrBodyTooLarge
	}
	r.remaining -= int64(n)
	return n, err
}

func (r *limitReadCloser) Close() error {
	return r.reader.Close()
}
