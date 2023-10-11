package sentencepiece

import (
	"io"
	"os"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
)

type Model struct {
	bos     int64
	eos     int64
	tk2id   map[string]uint64
	id2tk   map[uint64]string
	maxSize int
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
			if len(piece) > ret.maxSize {
				ret.maxSize = len(piece)
			}
		case ModelProto_SentencePiece_BYTE:
			piece = parseByte(piece)
			ret.tk2id[piece] = uint64(i)
			ret.id2tk[uint64(i)] = piece
			if len(piece) > ret.maxSize {
				ret.maxSize = len(piece)
			}
		case ModelProto_SentencePiece_UNKNOWN:
			ret.tk2id[piece] = uint64(i)
			ret.id2tk[uint64(i)] = piece
			if len(piece) > ret.maxSize {
				ret.maxSize = len(piece)
			}
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
	for i := 0; i < len(str); {
		var tk string
		var size int
		for j := m.maxSize; j > 0; j-- {
			if i+j > len(str) {
				continue
			}
			tk = str[i : i+j]
			if _, ok := m.tk2id[tk]; ok {
				break
			}
		}
		size = len(tk)
		if _, ok := m.tk2id[tk]; !ok {
			panic("unknown token")
		}
		ret = append(ret, m.tk2id[tk])
		i += size
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
