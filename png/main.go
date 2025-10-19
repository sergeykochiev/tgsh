package png

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"errors"
	"os"
	"encoding/binary"
	"hash/crc32"
	"compress/zlib"
	"bytes"
)

const (
	PngCT_Grayscale byte = 0
	PngCT_RGB byte = 2
	PngCT_Palette byte = 3
	PngCT_GrayscaleAlpha byte = 4
	PngCT_RGB_Alpha byte = 6
)

var PngCTSizeMap [7]int = [7]int{1,0,3,1,2,0,4}

const (
	PngIHDR string = "IHDR"
	PngIDAT string = "IDAT"
	PngIEND string = "IEND"
)

const (
	PngChunk_BaseLen uint32 = 12
	PngChunk_IHDRLen uint32 = 13
)

const PngSignatureLength int = 8
var PngSignature []byte = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

const PngCRCInitVal = 0x0

const PngMinLength int = PngSignatureLength + int(PngChunk_BaseLen) * 2 + int(PngChunk_IHDRLen)

type PngImage struct {
	w uint32
	h uint32
	depth byte
	colorType byte
	compression byte
	filter byte
	interlace byte
	
	imageData []byte
}

func CompressZlib(data []byte) []byte {
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()
	
	return b.Bytes()
}

func DecompressZlib(data []byte) []byte {
	b := bytes.NewReader(data)
	r, err := zlib.NewReader(b)
	if err != nil {
		log.Fatal(err)
	}
	output, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	r.Close()

	return output
}

func ComputeCRC(crc uint32, data []byte) uint32 {
	return crc32.Update(crc, crc32.IEEETable, data)
}

func ConstructPngChunk(data []byte, typ string, length uint32) []byte {
	output := make([]byte, 0, length + PngChunk_BaseLen)
	uint32Buf := make([]byte, 4)
	
	binary.BigEndian.PutUint32(uint32Buf, length)
	output = append(output, uint32Buf...)

	output = append(output, typ...)
	output = append(output, data...)

	binary.BigEndian.PutUint32(uint32Buf, ComputeCRC(ComputeCRC(PngCRCInitVal, []byte(typ)), data))
	output = append(output, uint32Buf...) 

	return output
}

func (p PngImage) actualWidth() int {
	return (int(p.w) * PngCTSizeMap[p.colorType] * int(p.depth / 8)) + 1
}

func (p PngImage) constructIHDR() []byte {
	uint32Buf := make([]byte, 4)
	data := make([]byte, 0, PngChunk_IHDRLen)

	binary.BigEndian.PutUint32(uint32Buf, p.w)
	data = append(data, uint32Buf...)

	binary.BigEndian.PutUint32(uint32Buf, p.h)
	data = append(data, uint32Buf...)

	data = append(data, p.depth)
	data = append(data, p.colorType)
	data = append(data, p.compression)
	data = append(data, p.filter)
	data = append(data, p.interlace)

	return ConstructPngChunk(data, PngIHDR, PngChunk_IHDRLen)
}

func (p PngImage) constructIEND() []byte {
	return ConstructPngChunk([]byte{}, PngIEND, 0)
}

func (p PngImage) constructData() []byte {
	w := p.actualWidth()

	lengthData := len(p.imageData)
	output := make([]byte, int(p.h) * w)

	var index int
	for i := range(int(p.h)) {
		output[i * w] = 0
		for j := 1; j < w; j++ {
			index = i * w + j
			if index - i < lengthData {
				output[index] = p.imageData[index - i]
				continue
			} else if index - i == lengthData {
				output[index] = 255
				continue
			}
			output[index] = byte(rand.Intn(255))
		}
	}

	return output
}

func (p PngImage) constructIDAT() []byte {
	data := CompressZlib(p.constructData())
	return ConstructPngChunk(data, PngIDAT, uint32(len(data)))
}

func (p PngImage) Encode() []byte {
	output := make([]byte, 0, PngMinLength)

	output = append(output, PngSignature...)
	output = append(output, p.constructIHDR()...)
	output = append(output, p.constructIDAT()...)
	output = append(output, p.constructIEND()...)

	return output
}

func deconstructSignature(data []byte) error {
	if len(data) < PngSignatureLength {
		return errors.New("not enough data")
	}
	for i := range(PngSignatureLength) {
		if data[i] != PngSignature[i] {
			return errors.New("invalid signature")
		}
	}
	return nil
}

func (p* PngImage) deconstructChunk(data []byte, expectTyp *string) (int, error) {
	if len(data) < int(PngChunk_BaseLen) {
		return 0, errors.New("not enough data")
	}
	chunkLength := binary.BigEndian.Uint32(data)
	typ := string(data[4:8])
	if *expectTyp != "" && *expectTyp != typ {
		return 4, fmt.Errorf("unexpected chunk type: %x; expected: %x (%s)", []byte(typ), []byte(*expectTyp), *expectTyp)
	} else {
		*expectTyp = typ
  }
	parsedCount, err := p.deconstructChunkData(data[8:], typ, int(chunkLength))
	if err != nil {
		return 4, fmt.Errorf("failed to deconstruct chunk of type %x: %s", []byte(typ), err)
	}
	return parsedCount + int(PngChunk_BaseLen), nil
}

func (p* PngImage) deconstructChunkData(data []byte, typ string, toParse int) (int, error) {
	switch typ {
		case PngIHDR: return int(PngChunk_IHDRLen), p.deconstructIHDRData(data)
		case PngIDAT: return toParse, p.deconstructIDATData(data, toParse)
		case PngIEND: return 0, nil
		default: return 0, fmt.Errorf("unsupported or invalid chunk type")
	}
}

func (p* PngImage) deconstructIHDRData(data []byte) error {
	if len(data) < int(PngChunk_IHDRLen) {
		return errors.New("not enough data")
	}
	p.w = binary.BigEndian.Uint32(data);
	if p.w == 0 {
		return errors.New("invalid width info")
	}
	p.h = binary.BigEndian.Uint32(data[4:]);
	if p.h == 0 {
		return errors.New("invalid height info")
	}
	p.depth = data[8]
	p.colorType = data[9]
	p.compression = data[10]
	if p.compression != 0 { return errors.New("invalid compression value") }
	p.filter = data[11]
	p.interlace = data[12]
	return nil
}

func (p* PngImage) deconstructIDATData(data []byte, toParse int) error {
	if len(data) < toParse {
		return errors.New("not enough length")
	}

	w := p.actualWidth()

	decompressed := DecompressZlib(data)

	p.imageData = nil
	p.imageData = []byte{}

	var index int
	for i := range(int(p.h)) {
		// filter := decompressed[i * w]
		for j := 1; j < w; j++ {
			index = i * w + j
			if decompressed[index] == 255 {
				return nil
			}
		  p.imageData = append(p.imageData, decompressed[index])
		}
	}

	return nil
}

func (p* PngImage) From(data []byte) (int, error) {
	err := deconstructSignature(data)
	if err != nil {
		return 0, err
	}

	var offset int
	globalOffset := PngSignatureLength

	typ := PngIHDR
	offset, err = p.deconstructChunk(data[globalOffset:], &typ)
	globalOffset += offset
	if err != nil {
		return offset, err
	}

	for typ != PngIEND && globalOffset < len(data) {
		typ = ""
		offset, err = p.deconstructChunk(data[globalOffset:], &typ)
		globalOffset += offset
		if err != nil {
			return globalOffset, err
		}
	}
	return globalOffset, nil
}

func (p* PngImage) Default(w, h uint32, imageData []byte) {
	p.w = w
	p.h = h
	p.imageData = imageData
	p.depth = 8
	p.colorType = PngCT_RGB_Alpha
	p.compression = 0
	p.filter = 0
	p.interlace = 0
}

func DisplayByteArr(data []byte) {
	for _, b := range(data) {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()
}

// TODO
// minify image where w*h > len(imageData) when constructing
// 
func main() {
	if len(os.Args) < 4 {
		fmt.Println("oops: not enough arguments. Usage: [program] [encode|decode] [filename1] [filename2]")
		return
	}
	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		fmt.Println("oops: what a file Mark!")
		return
	}
	var p PngImage 
	switch os.Args[1] {
		case "encode":
			p.Default(512, 512, data)
			encoded := p.Encode()
			os.WriteFile(os.Args[3], encoded, 0644)
		case "decode":
			parsed, err := p.From(data)
			if err != nil {
				fmt.Println("decoding error: ", err)
				fmt.Println("\ton byte", parsed, ": ", data[parsed:min(len(data),parsed + 8)])
				return
			}
			fmt.Println(string(p.imageData))
			os.WriteFile(os.Args[3], p.imageData, 0644)
		default:
			fmt.Println("oops: unknown option")
			return
	}
	fmt.Println("done")
}
