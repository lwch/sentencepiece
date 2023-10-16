package sentencepiece

import (
	"io"
	"os"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

type Model struct {
	bos   int64
	eos   int64
	unk   int64
	tk2id map[string]uint64
	id2tk map[uint64]string
}

func Load(dir string) (*Model, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadFrom(f)
}

func LoadFrom(r io.Reader) (*Model, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var m ModelProto
	err = proto.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	var ret Model
	ret.bos = -1
	ret.eos = -1
	ret.unk = -1
	ret.tk2id = make(map[string]uint64)
	ret.id2tk = make(map[uint64]string)
	for i, p := range m.GetPieces() {
		piece := p.GetPiece()
		switch p.GetType() {
		case ModelProto_SentencePiece_CONTROL:
			switch piece {
			case "<s>":
				ret.bos = int64(i)
			case "</s>":
				ret.eos = int64(i)
			}
		case ModelProto_SentencePiece_NORMAL:
			ret.tk2id[piece] = uint64(i)
			ret.id2tk[uint64(i)] = piece
		case ModelProto_SentencePiece_BYTE:
			piece = parseByte(piece)
			ret.tk2id[piece] = uint64(i)
			ret.id2tk[uint64(i)] = piece
		case ModelProto_SentencePiece_UNKNOWN:
			ret.unk = int64(i)
			ret.tk2id[piece] = uint64(i)
			ret.id2tk[uint64(i)] = piece
		}
	}
	return &ret, nil
}

func parseByte(str string) string {
	str = str[1+2 : len(str)-1]
	ch, _ := strconv.ParseUint(str, 16, 8)
	return string(rune(ch))
}

func (m *Model) Encode(str string, bos, eos bool) []uint64 {
	str = strings.ReplaceAll(str, " ", "▁")
	var ret []uint64
	if bos && m.bos != -1 {
		ret = append(ret, uint64(m.bos))
	}
	var prev int64
	var cache string
	var size int
	prev = -1
	for _, tk := range str {
		cache += string(tk)
		size++
		if id, ok := m.tk2id[cache]; ok {
			prev = int64(id)
			continue
		}
		if prev == -1 {
			if m.unk != -1 {
				for i := 0; i < size; i++ {
					ret = append(ret, uint64(m.unk))
				}
				prev = -1
				cache = ""
				size = 0
				continue
			}
			panic("unknown token")
		}
		ret = append(ret, uint64(prev))
		cache = string(tk)
		size = 1
		if n, ok := m.tk2id[cache]; ok {
			prev = int64(n)
			continue
		}
		prev = -1
	}
	if prev != -1 {
		ret = append(ret, uint64(prev))
	}
	if eos && m.eos != -1 {
		ret = append(ret, uint64(m.eos))
	}
	return ret
}

func (m *Model) Decode(tks []uint64) string {
	var ret string
	for _, tk := range tks {
		if m.bos != -1 && int64(tk) == m.bos {
			continue
		}
		if m.eos != -1 && int64(tk) == m.eos {
			continue
		}
		if m.unk != -1 && int64(tk) == m.unk {
			ret += "<unk>"
			continue
		}
		ret += m.id2tk[tk]
	}
	return strings.ReplaceAll(ret, "▁", " ")
}

func (m *Model) Count() int {
	size := len(m.id2tk)
	if m.bos != -1 {
		size++
	}
	if m.eos != -1 {
		size++
	}
	return size
}

func (m *Model) Bos() int64 {
	return m.bos
}

func (m *Model) Eos() int64 {
	return m.eos
}
