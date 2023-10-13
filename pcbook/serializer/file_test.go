package serializer_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"github.com/yenonn/pcbook/sample"
	"github.com/yenonn/pcbook/serializer"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()
	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	laptop1 := sample.NewLaptop()
	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	require.NoError(t, err)

	err = serializer.WriteProtobufToJsonFile(laptop1, jsonFile)
	require.NoError(t, err)

	laptop2 := sample.NewLaptop()
	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	require.NoError(t, err)

	//make sure laptop1 == laptop2
	require.True(t, proto.Equal(laptop1, laptop2))
}
