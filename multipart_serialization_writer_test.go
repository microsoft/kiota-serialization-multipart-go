package multipartserialization

import (
	"bytes"
	"io"
	"mime/multipart"
	"strings"
	"testing"
	"time"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoft/kiota-serialization-multipart-go/internal"

	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"

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

	assert.Nil(t, nilWriter.GetOnBeforeSerialization())
	assert.Error(t, nilWriter.SetOnBeforeSerialization(func(absser.Parsable) error { return nil }))
	assert.Nil(t, nilWriter.GetOnAfterObjectSerialization())
	assert.Error(t, nilWriter.SetOnAfterObjectSerialization(func(absser.Parsable) error { return nil }))
	assert.Nil(t, nilWriter.GetOnStartObjectSerialization())
	assert.Error(t, nilWriter.SetOnStartObjectSerialization(func(absser.Parsable, absser.SerializationWriter) error { return nil }))
}

func TestWriteObjectValueWithAdditionalValuesToMerge(t *testing.T) {
	writer := NewMultipartSerializationWriter()
	mockRequestAdapter := &internal.MockRequestAdapter{
		SerializationWriterFactory: jsonser.NewJsonSerializationWriterFactory(),
	}

	part1 := abstractions.NewMultipartBody()
	part1.SetRequestAdapter(mockRequestAdapter)
	assert.NoError(t, part1.AddOrReplacePart("field1", "application/octet-stream", []byte("hello")))

	part2 := abstractions.NewMultipartBody()
	part2.SetRequestAdapter(mockRequestAdapter)
	assert.NoError(t, part2.AddOrReplacePart("field2", "application/octet-stream", []byte("world")))

	err := writer.WriteObjectValue("", part1, part2)
	assert.NoError(t, err)

	content, err := writer.GetSerializedContent()
	assert.NoError(t, err)

	// Verify using standard MIME multipart reader with primary boundary
	reader := multipart.NewReader(bytes.NewReader(content), part1.GetBoundary())
	parsedParts := make(map[string]string)
	for {
		p, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		body, err := io.ReadAll(p)
		require.NoError(t, err)
		parsedParts[p.FormName()] = string(body)
	}

	assert.Len(t, parsedParts, 2)
	assert.Equal(t, "hello", parsedParts["field1"])
	assert.Equal(t, "world", parsedParts["field2"])
}

func TestWriteObjectValueWithAdditionalValuesEdgeCases(t *testing.T) {
	mockRequestAdapter := &internal.MockRequestAdapter{
		SerializationWriterFactory: jsonser.NewJsonSerializationWriterFactory(),
	}

	t.Run("nil item with additional values", func(t *testing.T) {
		writer := NewMultipartSerializationWriter()
		part1 := abstractions.NewMultipartBody()
		part1.SetRequestAdapter(mockRequestAdapter)
		assert.NoError(t, part1.AddOrReplacePart("field1", "application/octet-stream", []byte("hello")))

		part2 := abstractions.NewMultipartBody()
		part2.SetRequestAdapter(mockRequestAdapter)
		assert.NoError(t, part2.AddOrReplacePart("field2", "application/octet-stream", []byte("world")))

		err := writer.WriteObjectValue("", nil, part1, nil, part2)
		assert.NoError(t, err)

		content, err := writer.GetSerializedContent()
		assert.NoError(t, err)

		reader := multipart.NewReader(bytes.NewReader(content), part1.GetBoundary())
		parsedParts := make(map[string]string)
		for {
			p, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			body, err := io.ReadAll(p)
			require.NoError(t, err)
			parsedParts[p.FormName()] = string(body)
		}

		assert.Len(t, parsedParts, 2)
		assert.Equal(t, "hello", parsedParts["field1"])
		assert.Equal(t, "world", parsedParts["field2"])
	})

	t.Run("invalid additional value returns error", func(t *testing.T) {
		writer := NewMultipartSerializationWriter()
		part1 := abstractions.NewMultipartBody()
		part1.SetRequestAdapter(mockRequestAdapter)
		assert.NoError(t, part1.AddOrReplacePart("field1", "application/octet-stream", []byte("hello")))

		err := writer.WriteObjectValue("", part1, internal.NewTestEntity())
		assert.Error(t, err)
		assert.Equal(t, "only the serialization of multipart bodies is supported with MultipartSerializationWriter", err.Error())
	})

	t.Run("empty item and empty additional values", func(t *testing.T) {
		writer := NewMultipartSerializationWriter()
		err := writer.WriteObjectValue("", nil, nil)
		assert.NoError(t, err)

		content, err := writer.GetSerializedContent()
		assert.NoError(t, err)
		assert.Empty(t, content)
	})
}
