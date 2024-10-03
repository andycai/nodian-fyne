package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"html"
	"net/url"

	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/sha3"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type HashTool struct {
	window       fyne.Window
	input        *widget.Entry
	output       *widget.Entry
	hashSelect   *widget.Select
	encodeSelect *widget.Select
}

func NewHashTool(window fyne.Window) *HashTool {
	tool := &HashTool{
		window: window,
		input:  widget.NewMultiLineEntry(),
		output: widget.NewMultiLineEntry(),
		hashSelect: widget.NewSelect([]string{
			"CRC-32", "MD4", "MD5", "SHA1", "SHA224", "SHA256", "SHA384", "SHA512",
			"SHA512/224", "SHA512/256", "SHA3-224", "SHA3-256", "SHA3-384", "SHA3-512",
			"Keccak-224", "Keccak-256", "Keccak-384", "Keccak-512",
		}, nil),
		encodeSelect: widget.NewSelect([]string{
			"Base32", "Base64", "HTML", "URL",
		}, nil),
	}
	tool.input.SetPlaceHolder("Enter text to hash or encode/decode")
	tool.output.SetPlaceHolder("Result will appear here")
	tool.output.Disable()
	tool.hashSelect.SetSelected("MD5")
	tool.encodeSelect.SetSelected("Base64")
	return tool
}

func (h *HashTool) CreateUI() fyne.CanvasObject {
	hashButton := widget.NewButton("Hash", func() {
		h.performHash()
	})

	encodeButton := widget.NewButton("Encode", func() {
		h.performEncode()
	})

	decodeButton := widget.NewButton("Decode", func() {
		h.performDecode()
	})

	content := container.NewVBox(
		widget.NewLabel("Input:"),
		h.input,
		widget.NewLabel("Hash Function:"),
		h.hashSelect,
		hashButton,
		widget.NewLabel("Encoding:"),
		h.encodeSelect,
		container.NewHBox(encodeButton, decodeButton),
		widget.NewLabel("Output:"),
		h.output,
	)

	return container.NewPadded(content)
}

func (h *HashTool) performHash() {
	input := []byte(h.input.Text)
	var result string

	switch h.hashSelect.Selected {
	case "CRC-32":
		crc := crc32.ChecksumIEEE(input)
		result = fmt.Sprintf("%08x", crc)
	case "MD4":
		hash := md4.New()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil))
	case "MD5":
		hash := md5.Sum(input)
		result = hex.EncodeToString(hash[:])
	case "SHA1":
		hash := sha1.Sum(input)
		result = hex.EncodeToString(hash[:])
	case "SHA224":
		hash := sha256.Sum224(input)
		result = hex.EncodeToString(hash[:])
	case "SHA256":
		hash := sha256.Sum256(input)
		result = hex.EncodeToString(hash[:])
	case "SHA384":
		hash := sha512.Sum384(input)
		result = hex.EncodeToString(hash[:])
	case "SHA512":
		hash := sha512.Sum512(input)
		result = hex.EncodeToString(hash[:])
	case "SHA512/224":
		hash := sha512.Sum512_224(input)
		result = hex.EncodeToString(hash[:])
	case "SHA512/256":
		hash := sha512.Sum512_256(input)
		result = hex.EncodeToString(hash[:])
	case "SHA3-224":
		hash := sha3.New224()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil))
	case "SHA3-256":
		hash := sha3.New256()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil))
	case "SHA3-384":
		hash := sha3.New384()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil))
	case "SHA3-512":
		hash := sha3.New512()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil))
	case "Keccak-224":
		hash := sha3.NewLegacyKeccak256()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil)[:28]) // 截断到 224 位
	case "Keccak-256":
		hash := sha3.NewLegacyKeccak256()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil))
	case "Keccak-384":
		hash := sha3.NewLegacyKeccak512()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil)[:48]) // 截断到 384 位
	case "Keccak-512":
		hash := sha3.NewLegacyKeccak512()
		hash.Write(input)
		result = hex.EncodeToString(hash.Sum(nil))
	default:
		result = "Unsupported hash function"
	}

	h.output.SetText(result)
}

func (h *HashTool) performEncode() {
	input := h.input.Text
	var result string

	switch h.encodeSelect.Selected {
	case "Base32":
		result = base32.StdEncoding.EncodeToString([]byte(input))
	case "Base64":
		result = base64.StdEncoding.EncodeToString([]byte(input))
	case "HTML":
		result = html.EscapeString(input)
	case "URL":
		result = url.QueryEscape(input)
	default:
		result = "Unsupported encoding"
	}

	h.output.SetText(result)
}

func (h *HashTool) performDecode() {
	input := h.input.Text
	var result string

	switch h.encodeSelect.Selected {
	case "Base32":
		decoded, err := base32.StdEncoding.DecodeString(input)
		if err != nil {
			result = "Error: Invalid Base32 input"
		} else {
			result = string(decoded)
		}
	case "Base64":
		decoded, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			result = "Error: Invalid Base64 input"
		} else {
			result = string(decoded)
		}
	case "HTML":
		result = html.UnescapeString(input)
	case "URL":
		decoded, err := url.QueryUnescape(input)
		if err != nil {
			result = "Error: Invalid URL-encoded input"
		} else {
			result = decoded
		}
	default:
		result = "Unsupported decoding"
	}

	h.output.SetText(result)
}
