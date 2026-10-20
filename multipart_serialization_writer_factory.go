package multipartserialization

import (
	"errors"
	"strings"

	absser "github.com/microsoft/kiota-abstractions-go/serialization"
)

// MultipartSerializationWriterFactory implements SerializationWriterFactory for multipart/form-data.
type MultipartSerializationWriterFactory struct {
}

// NewMultipartSerializationWriterFactory creates a new instance of the MultipartSerializationWriterFactory.
func NewMultipartSerializationWriterFactory() *MultipartSerializationWriterFactory {
	return &MultipartSerializationWriterFactory{}
}

// GetValidContentType returns the valid content type for the SerializationWriterFactoryRegistry
func (f *MultipartSerializationWriterFactory) GetValidContentType() (string, error) {
	return "multipart/form-data", nil
}

// GetSerializationWriter returns the relevant SerializationWriter instance for the given content type
func (f *MultipartSerializationWriterFactory) GetSerializationWriter(contentType string) (absser.SerializationWriter, error) {
	validType, err := f.GetValidContentType()
	if err != nil {
		return nil, err
	} else if contentType == "" {
		return nil, errors.New("contentType is empty")
	}

	normalized := strings.ToLower(strings.TrimSpace(contentType))
	if normalized != validType && !strings.HasPrefix(normalized, validType+";") {
		return nil, errors.New("contentType is not valid")
	}

	return NewMultipartSerializationWriter(), nil
}

