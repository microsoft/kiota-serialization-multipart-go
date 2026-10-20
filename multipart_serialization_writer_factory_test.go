package multipartserialization

import (
	"testing"

	assert "github.com/stretchr/testify/assert"

	absser "github.com/microsoft/kiota-abstractions-go/serialization"
)

func TestMultipartSerializationFactoryWriterHonoursInterface(t *testing.T) {
	instance := NewMultipartSerializationWriterFactory()
	assert.Implements(t, (*absser.SerializationWriterFactory)(nil), instance)
}

func TestItGetsASerializationWriter(t *testing.T) {
	instance := NewMultipartSerializationWriterFactory()
	validContentTypes := []string{
		"multipart/form-data",
		"multipart/form-data; boundary=---------------------------974767299852498929531610575",
		"MULTIPART/FORM-DATA",
		" multipart/form-data; charset=utf-8 ",
		"multipart/form-data ; boundary=abc",
		"multipart/form-data\t; boundary=abc",
	}
	for _, ct := range validContentTypes {
		writer, err := instance.GetSerializationWriter(ct)
		assert.NoError(t, err)
		assert.NotNil(t, writer)
	}
}

func TestItReturnsAndErrorOnTheWrongContentType(t *testing.T) {
	instance := NewMultipartSerializationWriterFactory()
	invalidContentTypes := []string{
		"application/json",
		"multipart/related",
		"text/plain",
	}
	for _, ct := range invalidContentTypes {
		writer, err := instance.GetSerializationWriter(ct)
		assert.Error(t, err)
		assert.Nil(t, writer)
	}
}

func TestItReturnsErrorOnEmptyContentType(t *testing.T) {
	instance := NewMultipartSerializationWriterFactory()
	writer, err := instance.GetSerializationWriter("")
	assert.Error(t, err)
	assert.Nil(t, writer)
}
