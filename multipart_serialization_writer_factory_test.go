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
	writer, err := instance.GetSerializationWriter("multipart/form-data")
	assert.Nil(t, err)
	assert.NotNil(t, writer)
}
func TestItReturnsAndErrorOnTheWrongContentType(t *testing.T) {
	instance := NewMultipartSerializationWriterFactory()
	writer, err := instance.GetSerializationWriter("application/json")
	assert.NotNil(t, err)
	assert.Nil(t, writer)
}
