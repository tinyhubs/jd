package jd

import (
	"testing"
	"bytes"
	"github.com/tinyhubs/et/assert"
	"fmt"
)

// This example uses a Decoder to decode a stream of distinct JSON values.
func Test_ExampleDecoder_Token(t *testing.T) {
	const jsonStream = `
		{"Message": "Hello", "Array": [1, 2, 3], "Null": null, "Number": 1.234678998765546, "bool": false}
	`

	o, err := LoadObject(bytes.NewBufferString(jsonStream))
	fmt.Print(err)
	assert.NotNili(t, "json流式是正确的,必须加载成功", err)

	buf := bytes.NewBufferString("")
	o.Accept(NewSimplePrinter(buf, PrintOptions{}))

	fmt.Print(buf.String())

	var f float64 = 12121212121212121.445
	//var i int = 0
	//i = int(f)
	fmt.Println("------", int64(f))
}
