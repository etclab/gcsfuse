package akeso

import (
	"encoding/hex"
	"fmt"

	"github.com/etclab/aes256"
)

const (
	DataNonce = "akeso_data_nonce"
	DataTag   = "akeso_data_tag"
)

// value is the string
type MetadataNoExistError string

func (e MetadataNoExistError) Error() string {
	return fmt.Sprintf("akeso: metadata for key %q does not exist", string(e))
}

type MetadataDecodeError struct {
	Key string
	Err error
}

func NewMetadataDecodeError(key string, err error) MetadataDecodeError {
	return MetadataDecodeError{Key: key, Err: err}
}

func (e MetadataDecodeError) Error() string {
	return fmt.Sprintf("akeso: can't decode metadata for key %q: %v", e.Key, e.Err)
}

type MetadataEncodeError struct {
	Key string
	Err error
}

func NewMetadataEncodeError(key string, err error) MetadataEncodeError {
	return MetadataEncodeError{Key: key, Err: err}
}

func (e MetadataEncodeError) Error() string {
	return fmt.Sprintf("akeso: can't encode metadata for key %q: %v", e.Key, e.Err)
}

func MetadataDataNonce(metadata map[string]string) ([]byte, error) {
	key := DataNonce

	hexNonce, ok := metadata[key]
	if !ok {
		return nil, MetadataNoExistError(key)
	}

	nonce, err := hex.DecodeString(hexNonce)
	if err != nil {
		return nil, NewMetadataDecodeError(key, err)
	}

    if len(nonce) != aes256.NonceSize {
        return nil, NewMetadataDecodeError(key, aes256.NonceSizeError(len(nonce)))
    }

	return nonce, nil
}

func SetMetadataDataNonce(metadata map[string]string, nonce []byte) error {
	key := DataNonce

	if len(nonce) != aes256.NonceSize {
		return aes256.NonceSizeError(len(nonce))
	}

	hexNonce := hex.EncodeToString(nonce)
	metadata[key] = hexNonce

	return nil
}

func MetadataDataTag(metadata map[string]string) ([]byte, error) {
	key := DataTag

	hexTag, ok := metadata[key]
	if !ok {
		return nil, MetadataNoExistError(key)
	}

	tag, err := hex.DecodeString(hexTag)
	if err != nil {
		return nil, NewMetadataDecodeError(key, err)
	}

    if len(tag) != aes256.TagSize {
        return nil, NewMetadataDecodeError(key, aes256.TagSizeError(len(tag)))
    }

	return tag, nil
}

func SetMetadataDataTag(metadata map[string]string, tag []byte) error {
	key := DataTag

	if len(tag) != aes256.TagSize {
		return aes256.TagSizeError(len(tag))
	}

	hexTag := hex.EncodeToString(tag)
	metadata[key] = hexTag

	return nil
}