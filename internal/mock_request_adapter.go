package internal

import (
	"context"

	a "github.com/microsoft/kiota-abstractions-go"
	s "github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/microsoft/kiota-abstractions-go/store"
)

type MockRequestAdapter struct {
	SerializationWriterFactory s.SerializationWriterFactory
}

func (r *MockRequestAdapter) Send(context context.Context, requestInfo *a.RequestInformation, constructor s.ParsableFactory, errorMappings a.ErrorMappings) (s.Parsable, error) {
	return nil, nil
}
func (r *MockRequestAdapter) SendEnum(context context.Context, requestInfo *a.RequestInformation, parser s.EnumFactory, errorMappings a.ErrorMappings) (any, error) {
	return nil, nil
}
func (r *MockRequestAdapter) SendCollection(context context.Context, requestInfo *a.RequestInformation, constructor s.ParsableFactory, errorMappings a.ErrorMappings) ([]s.Parsable, error) {
	return nil, nil
}
func (r *MockRequestAdapter) SendEnumCollection(context context.Context, requestInfo *a.RequestInformation, parser s.EnumFactory, errorMappings a.ErrorMappings) ([]any, error) {
	return nil, nil
}
func (r *MockRequestAdapter) SendPrimitive(context context.Context, requestInfo *a.RequestInformation, typeName string, errorMappings a.ErrorMappings) (any, error) {
	return nil, nil
}
func (r *MockRequestAdapter) SendPrimitiveCollection(context context.Context, requestInfo *a.RequestInformation, typeName string, errorMappings a.ErrorMappings) ([]any, error) {
	return nil, nil
}
func (r *MockRequestAdapter) SendNoContent(context context.Context, requestInfo *a.RequestInformation, errorMappings a.ErrorMappings) error {
	return nil
}
func (r *MockRequestAdapter) ConvertToNativeRequest(context context.Context, requestInfo *a.RequestInformation) (any, error) {
	return nil, nil
}
func (r *MockRequestAdapter) GetSerializationWriterFactory() s.SerializationWriterFactory {
	return r.SerializationWriterFactory
}
func (r *MockRequestAdapter) EnableBackingStore(factory store.BackingStoreFactory) {
}
func (r *MockRequestAdapter) SetBaseUrl(baseUrl string) {
}
func (r *MockRequestAdapter) GetBaseUrl() string {
	return ""
}
