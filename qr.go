package blhelper

import (
	"strings"

	"github.com/skip2/go-qrcode"
)

// console block
const (
	BLACK = "\033[40m  \033[0m"
	WHITE = "\033[47m  \033[0m"
)

// QRCode wraps qrcode.QRCODE
type QRCode struct {
	*qrcode.QRCode
}

// OutPut qrcode string for console
func (q QRCode) OutPut() string {
	b := strings.Builder{}
	bitmap := q.Bitmap()
	for _, row := range bitmap[3 : len(bitmap)-3] {
		for _, cell := range row[3 : len(row)-3] {
			if cell {
				b.WriteString(BLACK)
			} else {
				b.WriteString(WHITE)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// NewQR returns a qrCode type
func NewQR(msg string) (*QRCode, error) {
	qr, err := qrcode.New(msg, qrcode.High)
	if err != nil {
		return nil, err
	}
	return &QRCode{qr}, nil
}
