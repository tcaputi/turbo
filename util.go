package turbo

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
)

func hash(obj interface{}) (error, []byte) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		return err, make([]byte, 20)
	}
	array := sha1.Sum(buf.Bytes())
	slice := make([]byte, 20)
	copy(array[:], slice)
	return nil, slice
}
