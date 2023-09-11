# sentencepiece

golang sentencepiece library for read llama tokenizer.model

`sentencepiece_model.proto` is from https://github.com/google/sentencepiece/blob/master/src/sentencepiece_model.proto

## install

```shell
go get -u github.com/lwch/sentencepiece
```

## useage

```go
m, err := sentencepiece.Load("./tokenizer.model")
if err != nil {
    panic(err)
}

// Encode to ids
fmt.Println(m.Encode("Hello, world!", true, true))

// Decode from ids
fmt.Println(m.Decode([]uint64{1, 2374, 2}))
```