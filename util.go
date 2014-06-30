import (
    "encoding/gob"
    "bytes"
	"crypto/sha1"
)

func hash(obj interface{}) error, byte[]{
	var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(obj)
    if err != nil {
        return err, nil
    }
    return nil, sha1.Sum(buf.Bytes())
}