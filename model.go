package sentencepiece

import (
	"os"

	"google.golang.org/protobuf/proto"
)

type Model struct {
	bos     uint64
	eos     uint64
	tk2id   map[string]uint64
	id2tk   map[uint64]string
	maxSize int
}

func Load(dir string) (*Model, error) {
	data, err := os.ReadFile(dir)
	if err != nil {
		return nil, err
	}
	var m ModelProto
	err = proto.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	var ret Model
	ret.tk2id = make(map[string]uint64)
	ret.id2tk = make(map[uint64]string)
	for i, p := range m.GetPieces() {
		piece := p.GetPiece()
		switch p.GetType() {
		case ModelProto_SentencePiece_CONTROL:
			switch piece {
			case "<s>":
				ret.bos = uint64(i)
			case "</s>":
				ret.eos = uint64(i)
			}
		case ModelProto_SentencePiece_NORMAL:
			ret.tk2id[piece] = uint64(i)
			ret.id2tk[uint64(i)] = piece
			if len(piece) > ret.maxSize {
				ret.maxSize = len(piece)
			}
		}
	}
	return &ret, nil
}

func (m *Model) Encode(str string, bos, eos bool) []uint64 {
	var ret []uint64
	if bos {
		ret = append(ret, m.bos)
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
		if tk == " " {
			tk = string(rune(0x2581)) // replace space to U+2581
			size = 1
		}
		if _, ok := m.tk2id[tk]; !ok {
			panic("unknown token")
		}
		ret = append(ret, m.tk2id[tk])
		i += size
	}
	if eos {
		ret = append(ret, m.eos)
	}
	return ret
}

func (m *Model) Decode(tks []uint64) string {
	var ret string
	for _, tk := range tks {
		if tk == m.bos || tk == m.eos {
			continue
		}
		ret += m.id2tk[tk]
	}
	return ret
}
