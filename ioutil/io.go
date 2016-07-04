package ioutil

import "io"

// ReadData always reads exactly n bytes from reader.
func Read(r io.Reader, n int) ([]byte, error) {
	read := 0

	result := make([]byte, n)

	for read < n {
		chunk := result[read:]
		nn, err := r.Read(chunk)
		if err != nil {
			return nil, err
		}

		read += nn
	}
	return result, nil
}
