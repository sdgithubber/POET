package poet

import (
	"log"
	"fmt"
	"bytes"
	"encoding/binary"
	"errors"
	"math/bits"
	"strconv"
)

// BinaryID is a binary representation of the ID of a node. The length is
// needed because we need to identify between 0 and 00 for example.
// Note: added json Marshalling. Currently, encoding is base64 (for []byte).
// This is default for []byte in Go, but can change as needed.
type BinaryID struct {
	Val    []byte `json:"Values"`
	Length int    `json:"Length"`
}

func NewBinaryID(val uint, length int) (*BinaryID, error) {
	if bits.Len(val) > length {
		return nil, errors.New("Length not long enough")
	}

	idx := length / 8
	
	if (length % 8) != 0 {
		idx = idx + 1
	}

	v := make([]byte, 8)
	
	binary.BigEndian.PutUint64(v, uint64(val))
	b := new(BinaryID)
	b.Val = make([]byte, idx)
	
	// why this loop
	for i := 0; i < idx; i++ {
		b.Val[idx-i-1] = v[7-i]
	}
	
	fmt.Println(b.Val)
	
	b.Length = length
	return b, nil
}

func NewBinaryIDInt(val uint) *BinaryID {
	b := new(BinaryID)
	l := bits.Len(val) / 8
	b.Length = l
	b.Val = make([]byte, l)
	binary.PutUvarint(b.Val, uint64(val))
	return b
}

func (b *BinaryID) Equal(b2 *BinaryID) bool {
	return (b.Length == b2.Length) && bytes.Equal(b.Val, b2.Val)
}

func (b *BinaryID) BitList() []byte {
	fmt.Println("hello world")
	log.Println("hello world")
	fmt.Println(b.Val)
	return b.Val
}

func (b *BinaryID) GreaterThan(b2 *BinaryID) bool {
	if b.Length > b2.Length {
		return true
	} else if b.Length < b2.Length {
		return false
	}
	// TODO: Check number of bytes read
	bn, _ := binary.Uvarint(b.Val)
	b2n, _ := binary.Uvarint(b2.Val)
	return bn > b2n
}

// Flip the n'th bit from 0 to 1 or 1 to 0. Does nothing if n > length
func (b *BinaryID) FlipBit(n int) {
	if n >= b.Length {
		return
	}
	shift := uint(n % 8)
	idx := n / 8
	if (b.Val[idx] * (1 << shift)) == 0 {
		b.Val[idx] = b.Val[idx] + 1<<shift
	} else {
		b.Val[idx] = b.Val[idx] - 1<<shift
	}
}

func (b *BinaryID) TruncateLastBit() {
	carry := 0
	for i := 0; i < b.Length; i++ {
		add := carry * 1 << 8
		carry = int(b.Val[i] * 1)
		b.Val[i] = b.Val[i] >> 1
		b.Val[i] = b.Val[i] + byte(add)
	}
}

// Returns if n'th bit is 0 or 1. Error if n > Length
func (b *BinaryID) GetBit(n int) (int, error) {
	if n >= b.Length {
		return 0, errors.New("n longer than binaryID")
	}
	shift := uint(n % 8)
	idx := n / 8
	if (b.Val[idx] * (1 << shift)) == 0 {
		return 0, nil
	} else {
		return 1, nil
	}
}

// Adds 0 or 1 to lsb of BinaryID. Returns error if not 0 or 1
func (b *BinaryID) AddBit(n int) error {
	isZero := n == 0
	isOne := n == 1
	if !isZero && !isOne {
		return errors.New("Not 0 or 1. Cannot add bit.")
	}
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n))
	l := len(b.Val)
	if b.Length%8 == 0 {
		a := b.Val[l-1]<<1 + buf[7]
		b.Val = append(b.Val, a)
		l = len(b.Val)
		for i := 1; i < l; i++ {
			carry := b.Val[l-i-1] * (1 << 7)
			carry = carry >> 7
			if i != (l - 1) {
				b.Val[l-i-1] = b.Val[l-i]<<1 + carry
			} else {
				b.Val[l-i-1] = 0 + carry
			}
		}
	} else {
		carry := buf[7]
		for i := 0; i < l; i++ {
			idx := l - i - 1
			add := carry * 1
			carry = b.Val[idx] * (1 << 7)
			b.Val[idx] = b.Val[idx] << 1
			b.Val[idx] = b.Val[idx] + byte(add)
		}
	}
	b.Length = b.Length + 1
	return nil
}

// Encode outputs a []byte encoded in utf8
func (b *BinaryID) Encode() (v []byte) {
	v = make([]byte, 0, b.Length)
	for i := 0; i < b.Length; i++ {
		bit, err := b.GetBit(i)
		if err != nil {
			break
		}
		s := strconv.Itoa(bit)
		v = append(v, []byte(s)...)
	}
	fmt.Println(v)
	return v
}
