package multipartserialization

import (
	"errors"
	"strings"
	"testing"
	"time"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoft/kiota-serialization-multipart-go/internal"

	assert "github.com/stretchr/testify/assert"

	absser "github.com/microsoft/kiota-abstractions-go/serialization"
	jsonser "github.com/microsoft/kiota-serialization-json-go"
)

func TestItReturnsAnErrorOnParsable(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()
	err := serializer.WriteObjectValue("key", internal.NewTestEntity())
	assert.NotNil(t, err)
	assert.Equal(t, "only the serialization of multipart bodies is supported with MultipartSerializationWriter", err.Error())
}

func TestItWritesAByteArray(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()
	value := make([]byte, 3)
	value[0] = 0x01
	value[1] = 0x02
	value[2] = 0x03
	serializer.WriteByteArrayValue("key", value)
	result, err := serializer.GetSerializedContent()
	assert.Nil(t, err)
	assert.Equal(t, "\u0001\u0002\u0003", string(result[:]))
}

func TestItWritesAStructuredObject(t *testing.T) {
	testEntity := internal.NewTestEntity()
	idValue := "48d31887-5fad-4d73-a9f5-3c356e68a038"
	testEntity.SetId(&idValue)
	durationValue, err := absser.ParseISODuration("P1M")
	assert.Nil(t, err)
	testEntity.SetWorkDuration(durationValue)
	workTimeValue, err := absser.ParseTimeOnly("08:00:00")
	assert.Nil(t, err)
	testEntity.SetStartWorkTime(workTimeValue)
	birthDayValue, err := absser.ParseDateOnly("2017-09-04")
	assert.Nil(t, err)
	testEntity.SetBirthDay(birthDayValue)
	testEntity.GetAdditionalData()["mobilePhone"] = nil
	testEntity.GetAdditionalData()["jobTitle"] = "Author"
	testEntity.GetAdditionalData()["accountEnabled"] = true
	createdDateTimeValue, err := time.Parse(time.RFC3339, "0001-01-01T00:00:00Z")
	assert.Nil(t, err)
	testEntity.GetAdditionalData()["createdDateTime"] = createdDateTimeValue
	testEntity.GetAdditionalData()["otherPhones"] = [][]string{{"device1", "device2"}}
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()
	multipartBody := abstractions.NewMultipartBody()
	mockRequestAdapter := &internal.MockRequestAdapter{
		SerializationWriterFactory: jsonser.NewJsonSerializationWriterFactory(),
	}
	multipartBody.SetRequestAdapter(mockRequestAdapter)
	binaryValue := make([]byte, 3)
	binaryValue[0] = 0x01
	binaryValue[1] = 0x02
	binaryValue[2] = 0x03
	multipartBody.AddOrReplacePart("testEntity", "application/json", testEntity)
	multipartBody.AddOrReplacePart("image", "application/octet-stream", binaryValue)
	serializer.WriteObjectValue("", multipartBody)
	result, err := serializer.GetSerializedContent()
	assert.Nil(t, err)
	strResult := string(result[:])
	assert.Contains(t, strResult, "--"+multipartBody.GetBoundary()+"\r\nContent-Type: application/octet-stream\r\nContent-Disposition: form-data; name=\"image\"\r\n\r\n"+string(binaryValue[:])+"\r\n")
	assert.Contains(t, strResult, "--"+multipartBody.GetBoundary()+"\r\nContent-Type: application/json\r\nContent-Disposition: form-data; name=\"testEntity\"\r\n\r\n")
	assert.Contains(t, strResult, "\r\n--"+multipartBody.GetBoundary()+"--\r\n")
	jsonOpenCurlyIndex := strings.Index(strResult, "{")
	jsonCloseCurlyIndex := strings.LastIndex(strResult, "}")
	jsonPayload := strResult[jsonOpenCurlyIndex : jsonCloseCurlyIndex+1]
	parsNodeFactory := jsonser.NewJsonParseNodeFactory()
	jsonParseNode, err := parsNodeFactory.GetRootParseNode("application/json", []byte(jsonPayload))
	assert.Nil(t, err)
	deserializedValue, err := jsonParseNode.GetObjectValue(internal.CreateTestEntityFromDiscriminator)
	assert.Nil(t, err)
	deserializedEntity, ok := deserializedValue.(internal.TestEntityable)
	if !ok {
		assert.Fail(t, "deserialized value is not of type TestEntityable")
	}
	assert.Equal(t, testEntity.GetId(), deserializedEntity.GetId())
	assert.Equal(t, testEntity.GetWorkDuration(), deserializedEntity.GetWorkDuration())
	assert.Equal(t, testEntity.GetStartWorkTime(), deserializedEntity.GetStartWorkTime())
	assert.Equal(t, testEntity.GetBirthDay(), deserializedEntity.GetBirthDay())
	assert.Equal(t, testEntity.GetAdditionalData()["mobilePhone"], deserializedEntity.GetAdditionalData()["mobilePhone"])
	// doing this as maps in Go don't guarantee order
}
func TestNilWriterDoesNotPanic(t *testing.T) {
	var nilWriter *MultipartSerializationWriter
	assert.NoError(t, nilWriter.Close())

	content, err := nilWriter.GetSerializedContent()
	assert.Error(t, err)
	assert.Nil(t, content)

	val := "test"
	assert.Error(t, nilWriter.WriteStringValue("k", &val))
	assert.Error(t, nilWriter.WriteByteArrayValue("k", []byte{1, 2, 3}))
	assert.Error(t, nilWriter.WriteObjectValue("k", nil))
}

func TestNilWriterCallbacks(t *testing.T) {
	var nilWriter *MultipartSerializationWriter

	assert.Nil(t, nilWriter.GetOnBeforeSerialization())
	assert.EqualError(t, nilWriter.SetOnBeforeSerialization(func(absser.Parsable) error { return nil }), "writer cannot be nil")

	assert.Nil(t, nilWriter.GetOnAfterObjectSerialization())
	assert.EqualError(t, nilWriter.SetOnAfterObjectSerialization(func(absser.Parsable) error { return nil }), "writer cannot be nil")

	assert.Nil(t, nilWriter.GetOnStartObjectSerialization())
	assert.EqualError(t, nilWriter.SetOnStartObjectSerialization(func(absser.Parsable, absser.SerializationWriter) error { return nil }), "writer cannot be nil")
}

func TestSerializationCallbackLifecycle(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	var executionOrder []string

	beforeAction := func(p absser.Parsable) error {
		executionOrder = append(executionOrder, "before")
		assert.NotNil(t, p)
		return nil
	}
	startAction := func(p absser.Parsable, w absser.SerializationWriter) error {
		executionOrder = append(executionOrder, "start")
		assert.NotNil(t, p)
		assert.NotNil(t, w)
		return nil
	}
	afterAction := func(p absser.Parsable) error {
		executionOrder = append(executionOrder, "after")
		assert.NotNil(t, p)
		return nil
	}

	assert.NoError(t, serializer.SetOnBeforeSerialization(beforeAction))
	assert.NoError(t, serializer.SetOnStartObjectSerialization(startAction))
	assert.NoError(t, serializer.SetOnAfterObjectSerialization(afterAction))

	assert.NotNil(t, serializer.GetOnBeforeSerialization())
	assert.NotNil(t, serializer.GetOnStartObjectSerialization())
	assert.NotNil(t, serializer.GetOnAfterObjectSerialization())

	body := abstractions.NewMultipartBody()
	body.SetRequestAdapter(&internal.MockRequestAdapter{
		SerializationWriterFactory: jsonser.NewJsonSerializationWriterFactory(),
	})
	assert.NoError(t, body.AddOrReplacePart("field", "text/plain", "content"))

	err := serializer.WriteObjectValue("body", body)
	assert.NoError(t, err)

	assert.Equal(t, []string{"before", "start", "after"}, executionOrder)
}

func TestSerializationCallbackErrorPropagationAndNilActions(t *testing.T) {
	t.Run("start callback error propagation aborts serialization", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		afterCalled := false
		assert.NoError(t, serializer.SetOnStartObjectSerialization(func(p absser.Parsable, w absser.SerializationWriter) error {
			return errors.New("start hook failed")
		}))
		assert.NoError(t, serializer.SetOnAfterObjectSerialization(func(p absser.Parsable) error {
			afterCalled = true
			return nil
		}))

		body := abstractions.NewMultipartBody()
		err := serializer.WriteObjectValue("body", body)
		assert.EqualError(t, err, "start hook failed")
		assert.False(t, afterCalled, "after callback should not be invoked when start callback fails")
	})

	t.Run("nil callback actions can be set and clear previous handlers", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		assert.NoError(t, serializer.SetOnBeforeSerialization(func(absser.Parsable) error { return nil }))
		assert.NoError(t, serializer.SetOnStartObjectSerialization(func(absser.Parsable, absser.SerializationWriter) error { return nil }))
		assert.NoError(t, serializer.SetOnAfterObjectSerialization(func(absser.Parsable) error { return nil }))

		assert.NotNil(t, serializer.GetOnBeforeSerialization())
		assert.NotNil(t, serializer.GetOnStartObjectSerialization())
		assert.NotNil(t, serializer.GetOnAfterObjectSerialization())

		assert.NoError(t, serializer.SetOnBeforeSerialization(nil))
		assert.NoError(t, serializer.SetOnStartObjectSerialization(nil))
		assert.NoError(t, serializer.SetOnAfterObjectSerialization(nil))

		assert.Nil(t, serializer.GetOnBeforeSerialization())
		assert.Nil(t, serializer.GetOnStartObjectSerialization())
		assert.Nil(t, serializer.GetOnAfterObjectSerialization())
	})
}


func TestWriteObjectValueDoesNotInvokeCallbackOnUnsupportedParsable(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	called := false
	err := serializer.SetOnStartObjectSerialization(func(p absser.Parsable, w absser.SerializationWriter) error {
		called = true
		return nil
	})
	assert.NoError(t, err)

	err = serializer.WriteObjectValue("key", internal.NewTestEntity())
	assert.Error(t, err)
	assert.Equal(t, "only the serialization of multipart bodies is supported with MultipartSerializationWriter", err.Error())
	assert.False(t, called, "start callback should not be invoked for unsupported Parsable types")
}

func TestNilWriterUnsupportedMethodsReturnNilError(t *testing.T) {
	var nilWriter *MultipartSerializationWriter

	boolVal := true
	assert.EqualError(t, nilWriter.WriteBoolValue("k", &boolVal), "writer cannot be nil")

	byteVal := byte(1)
	assert.EqualError(t, nilWriter.WriteByteValue("k", &byteVal), "writer cannot be nil")

	int8Val := int8(1)
	assert.EqualError(t, nilWriter.WriteInt8Value("k", &int8Val), "writer cannot be nil")

	int32Val := int32(1)
	assert.EqualError(t, nilWriter.WriteInt32Value("k", &int32Val), "writer cannot be nil")

	int64Val := int64(1)
	assert.EqualError(t, nilWriter.WriteInt64Value("k", &int64Val), "writer cannot be nil")

	float32Val := float32(1.0)
	assert.EqualError(t, nilWriter.WriteFloat32Value("k", &float32Val), "writer cannot be nil")

	float64Val := float64(1.0)
	assert.EqualError(t, nilWriter.WriteFloat64Value("k", &float64Val), "writer cannot be nil")

	timeVal := time.Now()
	assert.EqualError(t, nilWriter.WriteTimeValue("k", &timeVal), "writer cannot be nil")

	assert.EqualError(t, nilWriter.WriteISODurationValue("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteTimeOnlyValue("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteDateOnlyValue("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteUUIDValue("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteAnyValue("k", "any"), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteAdditionalData(map[string]interface{}{}), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteNullValue("k"), "writer cannot be nil")

	assert.EqualError(t, nilWriter.WriteCollectionOfObjectValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfStringValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfInt32Values("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfInt64Values("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfFloat32Values("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfFloat64Values("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfTimeValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfISODurationValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfTimeOnlyValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfDateOnlyValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfUUIDValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfBoolValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfByteValues("k", nil), "writer cannot be nil")
	assert.EqualError(t, nilWriter.WriteCollectionOfInt8Values("k", nil), "writer cannot be nil")
}
