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

	"github.com/google/uuid"
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

func TestMultipartSerializationWriterHonoursInterface(t *testing.T) {
	instance := NewMultipartSerializationWriter()
	assert.Implements(t, (*absser.SerializationWriter)(nil), instance)
}

func TestWriteStringValue(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	val := "text/plain"
	err := serializer.WriteStringValue("Content-Type", &val)
	assert.NoError(t, err)

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Equal(t, "Content-Type: text/plain\r\n", string(content))
}

func TestWriteStringValueWithNilValueOmitsHeader(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	// Named header with nil value should leave the buffer unchanged
	err := serializer.WriteStringValue("Content-Type", nil)
	assert.NoError(t, err)

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, 0, len(content))

	// Pre-populated buffer remains unchanged when a named header with nil value is written
	initialVal := "form-data; name=\"file\""
	err = serializer.WriteStringValue("Content-Disposition", &initialVal)
	assert.NoError(t, err)

	contentBefore, err := serializer.GetSerializedContent()
	assert.NoError(t, err)

	err = serializer.WriteStringValue("Content-Type", nil)
	assert.NoError(t, err)

	contentAfter, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Equal(t, contentBefore, contentAfter)
}

func TestWriteStringValueWithEmptyKeyAndValue(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	val := "raw content line"
	err := serializer.WriteStringValue("", &val)
	assert.NoError(t, err)

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Equal(t, "raw content line\r\n", string(content))
}

func TestWriteStringValueHeaderKeyNormalization(t *testing.T) {
	t.Run("trims leading and trailing whitespace from header key", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "application/json"
		err := serializer.WriteStringValue("  Content-Type  ", &val)
		assert.NoError(t, err)

		content, err := serializer.GetSerializedContent()
		assert.NoError(t, err)
		assert.Equal(t, "Content-Type: application/json\r\n", string(content))
	})

	t.Run("whitespace-only key is treated as unkeyed content", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "unkeyed body line"
		err := serializer.WriteStringValue("   \t   ", &val)
		assert.NoError(t, err)

		content, err := serializer.GetSerializedContent()
		assert.NoError(t, err)
		assert.Equal(t, "unkeyed body line\r\n", string(content))
	})
}

func TestWriteStringValueCRLFValidation(t *testing.T) {
	t.Run("returns error when header key contains carriage return", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "application/json"
		err := serializer.WriteStringValue("Content\r-Type", &val)
		assert.EqualError(t, err, "header key cannot contain CRLF characters")
	})

	t.Run("returns error when header key contains newline", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "application/json"
		err := serializer.WriteStringValue("Content\n-Type", &val)
		assert.EqualError(t, err, "header key cannot contain CRLF characters")
	})

	t.Run("returns error when header value contains carriage return", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "text/plain\r\nX-Injected: value"
		err := serializer.WriteStringValue("Content-Type", &val)
		assert.EqualError(t, err, "header value cannot contain CRLF characters")
	})

	t.Run("returns error when header value contains newline", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "text/plain\nInjected: true"
		err := serializer.WriteStringValue("Content-Type", &val)
		assert.EqualError(t, err, "header value cannot contain CRLF characters")
	})

	t.Run("returns error when header key is CRLF-only", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		crlfKeys := []string{"\r\n", "\r", "\n", "\r\n\r\n"}
		val := "application/json"
		for _, key := range crlfKeys {
			err := serializer.WriteStringValue(key, &val)
			assert.EqualError(t, err, "header key cannot contain CRLF characters")
		}
	})

	t.Run("returns error when header key is CRLF-padded", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		paddedKeys := []string{"  Content-Type\r\n", "\r\nContent-Type  ", "\r\nContent-Type\r\n", "\nContent-Type\r"}
		val := "application/json"
		for _, key := range paddedKeys {
			err := serializer.WriteStringValue(key, &val)
			assert.EqualError(t, err, "header key cannot contain CRLF characters")
		}
	})

	t.Run("allows multiline value when key is empty", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "first line\r\nsecond line"
		err := serializer.WriteStringValue("", &val)
		assert.NoError(t, err)

		content, err := serializer.GetSerializedContent()
		assert.NoError(t, err)
		assert.Equal(t, "first line\r\nsecond line\r\n", string(content))
	})

	t.Run("allows multiline value when key is whitespace-only", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		val := "first line\r\nsecond line"
		err := serializer.WriteStringValue("   \t   ", &val)
		assert.NoError(t, err)

		content, err := serializer.GetSerializedContent()
		assert.NoError(t, err)
		assert.Equal(t, "first line\r\nsecond line\r\n", string(content))
	})
}

func TestWriteStringValueColonValidation(t *testing.T) {
	t.Run("returns error when header key contains colon", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		colonKeys := []string{
			"Content-Type:",
			":Content-Type",
			"Content:Type",
			"Content-Type: text/plain",
			":",
			" : ",
		}
		val := "text/plain"
		for _, key := range colonKeys {
			err := serializer.WriteStringValue(key, &val)
			assert.EqualError(t, err, "header key cannot contain colon characters")
		}
	})
}

func TestWriteStringValueWhitespaceValidation(t *testing.T) {
	t.Run("returns error when header key contains internal whitespace", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		whitespaceKeys := []string{
			"Content Type",
			"Content\tType",
			"Content  Type",
			"X Custom Header",
			" Content Type ",
		}
		val := "text/plain"
		for _, key := range whitespaceKeys {
			err := serializer.WriteStringValue(key, &val)
			assert.EqualError(t, err, "header key cannot contain whitespace")
		}
	})
}

func TestWriteStringValueNullByteKeyValidation(t *testing.T) {
	t.Run("returns error when header key contains null character", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		nullKeys := []string{
			"Content\x00-Type",
			"\x00Content-Type",
			"Content-Type\x00",
			"\x00",
			" \x00 ",
		}
		val := "text/plain"
		for _, key := range nullKeys {
			err := serializer.WriteStringValue(key, &val)
			assert.EqualError(t, err, "header key cannot contain null characters")
		}
	})
}

func TestWriteStringValueNullByteValueValidation(t *testing.T) {
	t.Run("returns error when named header value contains null character", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		nullValues := []string{
			"text/plain\x00injected",
			"\x00text/plain",
			"text/plain\x00",
			"\x00",
		}
		for _, val := range nullValues {
			err := serializer.WriteStringValue("Content-Type", &val)
			assert.EqualError(t, err, "header value cannot contain null characters")
		}
	})

	t.Run("allows null characters in unkeyed content line", func(t *testing.T) {
		serializer := NewMultipartSerializationWriter()
		defer serializer.Close()

		rawVal := "binary\x00data\x00payload"
		err := serializer.WriteStringValue("", &rawVal)
		assert.NoError(t, err)

		content, err := serializer.GetSerializedContent()
		assert.NoError(t, err)
		assert.Equal(t, "binary\x00data\x00payload\r\n", string(content))
	})
}

func TestGetSerializedContentOnUninitializedWriter(t *testing.T) {
	uninitializedWriter := &MultipartSerializationWriter{}
	content, err := uninitializedWriter.GetSerializedContent()
	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, 0, len(content))
}

func TestWriteAdditionalDataWithEmptyAndNilValues(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	// Nil map should safely succeed without error
	err := serializer.WriteAdditionalData(nil)
	assert.NoError(t, err)

	// Empty map should safely succeed without error
	err = serializer.WriteAdditionalData(map[string]interface{}{})
	assert.NoError(t, err)

	// Non-empty map should return unsupported error
	err = serializer.WriteAdditionalData(map[string]interface{}{"key": "value"})
	assert.Error(t, err)
	assert.EqualError(t, err, "serialization of additional data is not supported with MultipartSerializationWriter")

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Empty(t, content)
}

func TestWriteAnyValueWithNilAndNonNilValues(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	// Nil value should safely succeed without error
	err := serializer.WriteAnyValue("key", nil)
	assert.NoError(t, err)

	// Non-nil string should return unsupported error
	err = serializer.WriteAnyValue("key", "test")
	assert.Error(t, err)
	assert.EqualError(t, err, "serialization of any value is not supported with MultipartSerializationWriter")

	// Non-nil number should return unsupported error
	err = serializer.WriteAnyValue("key", 123)
	assert.Error(t, err)
	assert.EqualError(t, err, "serialization of any value is not supported with MultipartSerializationWriter")

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Empty(t, content)
}

func TestWriteCollectionMethodsWithNilAndEmptyValues(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	// Nil collections should safely succeed without error
	assert.NoError(t, serializer.WriteCollectionOfObjectValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfStringValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfInt32Values("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfInt64Values("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfFloat32Values("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfFloat64Values("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfTimeValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfISODurationValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfTimeOnlyValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfDateOnlyValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfUUIDValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfBoolValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfByteValues("key", nil))
	assert.NoError(t, serializer.WriteCollectionOfInt8Values("key", nil))

	// Empty collections should safely succeed without error
	assert.NoError(t, serializer.WriteCollectionOfObjectValues("key", []absser.Parsable{}))
	assert.NoError(t, serializer.WriteCollectionOfStringValues("key", []string{}))
	assert.NoError(t, serializer.WriteCollectionOfInt32Values("key", []int32{}))
	assert.NoError(t, serializer.WriteCollectionOfInt64Values("key", []int64{}))
	assert.NoError(t, serializer.WriteCollectionOfFloat32Values("key", []float32{}))
	assert.NoError(t, serializer.WriteCollectionOfFloat64Values("key", []float64{}))
	assert.NoError(t, serializer.WriteCollectionOfTimeValues("key", []time.Time{}))
	assert.NoError(t, serializer.WriteCollectionOfISODurationValues("key", []absser.ISODuration{}))
	assert.NoError(t, serializer.WriteCollectionOfTimeOnlyValues("key", []absser.TimeOnly{}))
	assert.NoError(t, serializer.WriteCollectionOfDateOnlyValues("key", []absser.DateOnly{}))
	assert.NoError(t, serializer.WriteCollectionOfUUIDValues("key", []uuid.UUID{}))
	assert.NoError(t, serializer.WriteCollectionOfBoolValues("key", []bool{}))
	assert.NoError(t, serializer.WriteCollectionOfByteValues("key", []byte{}))
	assert.NoError(t, serializer.WriteCollectionOfInt8Values("key", []int8{}))

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Empty(t, content)
}

func TestWriteCollectionMethodsWithNonEmptyValues(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	expectedErr := "collections serialization is not supported with MultipartSerializationWriter"

	// Non-empty collections should return the unsupported error
	err := serializer.WriteCollectionOfObjectValues("key", []absser.Parsable{internal.NewTestEntity()})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfStringValues("key", []string{"value"})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfInt32Values("key", []int32{42})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfInt64Values("key", []int64{100})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfFloat32Values("key", []float32{3.14})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfFloat64Values("key", []float64{2.718})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfTimeValues("key", []time.Time{time.Now()})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfISODurationValues("key", []absser.ISODuration{{}})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfTimeOnlyValues("key", []absser.TimeOnly{*absser.NewTimeOnly(time.Now())})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfDateOnlyValues("key", []absser.DateOnly{*absser.NewDateOnly(time.Now())})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfUUIDValues("key", []uuid.UUID{uuid.New()})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfBoolValues("key", []bool{true})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfByteValues("key", []byte{0x01})
	assert.EqualError(t, err, expectedErr)

	err = serializer.WriteCollectionOfInt8Values("key", []int8{8})
	assert.EqualError(t, err, expectedErr)
}

func TestWriteStringValueWithNilAndNonNilValues(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	// Nil string values with or without key should be safe no-ops
	err := serializer.WriteStringValue("key", nil)
	assert.NoError(t, err)

	err = serializer.WriteStringValue("", nil)
	assert.NoError(t, err)

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Empty(t, content, "must not emit key or newlines for nil string value")

	// Non-nil string values should serialize properly
	headerVal := "application/json"
	err = serializer.WriteStringValue("Content-Type", &headerVal)
	assert.NoError(t, err)

	boundaryVal := "--boundary"
	err = serializer.WriteStringValue("", &boundaryVal)
	assert.NoError(t, err)

	emptyVal := ""
	err = serializer.WriteStringValue("", &emptyVal)
	assert.NoError(t, err)

	content, err = serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Equal(t, "Content-Type: application/json\r\n--boundary\r\n\r\n", string(content))
}

func TestScalarWriteMethodsWithNilValues(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	// All scalar write methods with nil value should safely succeed as no-op
	assert.NoError(t, serializer.WriteBoolValue("key", nil))
	assert.NoError(t, serializer.WriteByteValue("key", nil))
	assert.NoError(t, serializer.WriteInt8Value("key", nil))
	assert.NoError(t, serializer.WriteInt32Value("key", nil))
	assert.NoError(t, serializer.WriteInt64Value("key", nil))
	assert.NoError(t, serializer.WriteFloat32Value("key", nil))
	assert.NoError(t, serializer.WriteFloat64Value("key", nil))
	assert.NoError(t, serializer.WriteTimeValue("key", nil))
	assert.NoError(t, serializer.WriteISODurationValue("key", nil))
	assert.NoError(t, serializer.WriteTimeOnlyValue("key", nil))
	assert.NoError(t, serializer.WriteDateOnlyValue("key", nil))
	assert.NoError(t, serializer.WriteUUIDValue("key", nil))
	assert.NoError(t, serializer.WriteNullValue("key"))

	content, err := serializer.GetSerializedContent()
	assert.NoError(t, err)
	assert.Empty(t, content, "buffer must remain empty when nil scalar values are written")
}

func TestScalarWriteMethodsWithNonNilValuesReturnUnsupportedError(t *testing.T) {
	serializer := NewMultipartSerializationWriter()
	defer serializer.Close()

	boolVal := true
	assert.EqualError(t, serializer.WriteBoolValue("k", &boolVal), "serialization of bool value is not supported with MultipartSerializationWriter")

	byteVal := byte(1)
	assert.EqualError(t, serializer.WriteByteValue("k", &byteVal), "serialization of byte value is not supported with MultipartSerializationWriter")

	int8Val := int8(1)
	assert.EqualError(t, serializer.WriteInt8Value("k", &int8Val), "serialization of int8 value is not supported with MultipartSerializationWriter")

	int32Val := int32(1)
	assert.EqualError(t, serializer.WriteInt32Value("k", &int32Val), "serialization of int32 value is not supported with MultipartSerializationWriter")

	int64Val := int64(1)
	assert.EqualError(t, serializer.WriteInt64Value("k", &int64Val), "serialization of int64 value is not supported with MultipartSerializationWriter")

	float32Val := float32(1.0)
	assert.EqualError(t, serializer.WriteFloat32Value("k", &float32Val), "serialization of float32 value is not supported with MultipartSerializationWriter")

	float64Val := float64(1.0)
	assert.EqualError(t, serializer.WriteFloat64Value("k", &float64Val), "serialization of float64 value is not supported with MultipartSerializationWriter")

	timeVal := time.Now()
	assert.EqualError(t, serializer.WriteTimeValue("k", &timeVal), "serialization of time value is not supported with MultipartSerializationWriter")

	isoVal := absser.NewDuration(1, 0, 0, 0, 0, 0, 0)
	assert.EqualError(t, serializer.WriteISODurationValue("k", isoVal), "serialization of ISODuration value is not supported with MultipartSerializationWriter")

	timeOnlyVal := absser.NewTimeOnly(time.Now())
	assert.EqualError(t, serializer.WriteTimeOnlyValue("k", timeOnlyVal), "serialization of TimeOnly value is not supported with MultipartSerializationWriter")

	dateOnlyVal := absser.NewDateOnly(time.Now())
	assert.EqualError(t, serializer.WriteDateOnlyValue("k", dateOnlyVal), "serialization of DateOnly value is not supported with MultipartSerializationWriter")

	uuidVal := uuid.New()
	assert.EqualError(t, serializer.WriteUUIDValue("k", &uuidVal), "serialization of UUID value is not supported with MultipartSerializationWriter")
}
