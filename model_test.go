package sentencepiece

import (
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	_, err := Load("./tokenizer.model")
	if err != nil {
		t.Error(err)
	}
}

func TestEncode(t *testing.T) {
	m, err := Load("./tokenizer.model")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m.Encode("Hello, world!", true, true))
}

func TestDecode(t *testing.T) {
	m, err := Load("./tokenizer.model")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(m.Decode([]uint64{1, 2374, 2}))
}
