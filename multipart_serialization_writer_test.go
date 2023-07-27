package multipartserialization

import (
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
