package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

/*
	BitReader implementation
*/

const (
	BufferCapacity int = 8
)

type BitReader struct {
	// Used buffer capacity in bits.
	capacity int

	// Internal buffer represented as an int. Only the least significant byte is used.
	// An int is used instead of a byte int-to-byte conversions in the VM.
	buffer int

	// Underlying input stream.
	reader *bufio.Reader
}

func NewBitReader(reader *bufio.Reader) *BitReader {
	return &BitReader{
		capacity: 0,
		buffer:   0,
		reader:   reader,
	}
}

/**
 * Resets this instance and sets a new underlying input stream. This method
 * allows instances of this class to be re-used. The resulting state after
 * calling this method is identical to that of a newly created instance.
 */
func (r *BitReader) SetReader(reader *bufio.Reader) {
	r.reader = reader
	r.buffer = 0
	r.capacity = 0
}

func (r *BitReader) readDirectByte() (int, error) {
	b, err := r.reader.ReadByte()
	if err != nil {
		return -1, err
	}
	return int(b), nil
}

/**
 * If buffer is empty, read byte from underlying stream.
 */
func (r *BitReader) readBuffer() error {
	b, err := r.readDirectByte()
	if err != nil {
		return err
	}
	r.buffer = b
	r.capacity = BufferCapacity
	return nil
}

/**
 * Discard any bits currently in the buffer to byte-align stream
 */
func (r *BitReader) Align() error {
	if r.capacity != 0 {
		r.capacity = 0
	}
	return nil
}

/**
 * Returns current byte buffer without actually reading data
 */
func (r *BitReader) LookAhead() (int, error) {
	if r.capacity == 0 {
		if err := r.readBuffer(); err != nil {
			return -1, err
		}
	}
	return r.buffer, nil
}

/**
 * Skip n bytes
 */
func (r *BitReader) Skip(n int64) error {
	if r.capacity == 0 {
		// algined
		for n != 0 {
			skipped, err := r.reader.Discard(int(n))
			if err != nil {
				return err
			}
			n -= int64(skipped)
		}
	} else {
		// not aligned
		//TODO: Wierd range in Java code: for i := 0; i < n; n++ {...}
		//for range n {
		for i := int64(0); i < n; n++ {
			if _, err := r.ReadBits(8); err != nil {
				return err
			}
		}
	}

	return nil
}

/**
 * Return next bit from underlying stream.
 */
func (r *BitReader) ReadBit() (int, error) {
	if r.capacity == 0 {
		if err := r.readBuffer(); err != nil {
			return -1, err
		}
	}
	r.capacity--
	return (r.buffer >> r.capacity) & 0x1, nil
}

/**
 * Read the next n bits and return the result as an integer.
 */
func (r *BitReader) ReadBits(n int) (int, error) {
	if n <= 0 {
		return -1, fmt.Errorf("number bits to read must have positive value")
	}

	var result int
	var err error

	if n <= r.capacity {
		// buffer already holds all necessary bits
		r.capacity -= n
		result = (r.buffer >> r.capacity) & (0xff >> (BufferCapacity - n))
	} else if r.capacity == 0 && n == BufferCapacity {
		// possible to read direct byte, nothing else to do
		result, err = r.readDirectByte()
		if err != nil {
			return -1, err
		}
	} else {
		// get as many bits from buffer as possible
		result = r.buffer & (0xff >> (BufferCapacity - r.capacity))
		n -= r.capacity
		r.capacity = 0

		// possibly read whole bytes
		for n > 7 {
			if r.capacity == 0 {
				if err := r.readBuffer(); err != nil {
					return -1, err
				}
			}

			result = (result << BufferCapacity) | r.buffer
			n -= BufferCapacity
			r.capacity = 0
		}

		// read the rest of the bits
		if n > 0 {
			if r.capacity == 0 {
				if err := r.readBuffer(); err != nil {
					return -1, err
				}
			}
			r.capacity = BufferCapacity - n
			result = (result << n) | (r.buffer >> r.capacity)
		}
	}

	return result, nil
}

/**
 * Reads one byte (8 bits) of data from the input stream
 */
func (r *BitReader) Read() (int, error) {
	// possible to read direct byte?
	if r.capacity == 0 {
		return r.readDirectByte()
	} else {
		return r.ReadBits(BufferCapacity)
	}
}

/**
 * Reads one byte (8 bits) of data from the input stream
 */
func (r *BitReader) ReadToBuffer(buffer []byte, offset, length int) error {
	if length < 0 {
		return fmt.Errorf("length must have postive value")
	} else if length == 0 {
		return nil
	}

	if r.capacity == 0 {
		// byte-aligned --> read all bytes at byte-border (at once?)
		readBytes := 0
		for readBytes < length {
			br, err := r.reader.Read(buffer[readBytes : length+readBytes])
			if err == io.EOF {
				return errors.New("premature EOS found while reading data")
			}
			if err != nil {
				return err
			}
			readBytes += br
		}
	} else {
		shift := BufferCapacity - r.capacity

		for i := range length {
			nextByte, err := r.readDirectByte()
			if err != nil {
				return err
			}
			buffer[i] = byte((r.buffer << shift) | (nextByte >> r.capacity))
			r.buffer = nextByte
		}
	}

	return nil
}

/*
	BitWriter implementation
*/

const (
	BitsInByte int = 8
)

type BitWriter struct {
	buffer   int
	capacity int
	writer   bufio.Writer
	len      int
}

func NewBitWriter(writer bufio.Writer) *BitWriter {
	return &BitWriter{
		buffer:   0,
		capacity: BitsInByte,
		writer:   writer,
		len:      0,
	}
}

/**
 * Returns a reference to underlying output stream.
 */
func (w *BitWriter) GetUnderlyingWriter() *bufio.Writer {
	return &w.writer
}

func (w *BitWriter) GetLength() int {
	return w.len
}

func (w *BitWriter) flushBuffer() error {
	if w.capacity == 0 {
		if err := w.writer.WriteByte(byte(w.buffer & 0xFF)); err != nil {
			return err
		}
		w.capacity = BitsInByte
		w.buffer = 0
		w.len++
	}

	return nil
}

func (w *BitWriter) IsByteAligned() bool {
	return w.capacity == BitsInByte
}

func (w *BitWriter) GetBitsInByffer() int {
	return BitsInByte - w.capacity
}

func (w *BitWriter) Flush() error {
	if err := w.Align(); err != nil {
		return err
	}
	return w.writer.Flush()
}

func (w *BitWriter) Align() error {
	if w.capacity < BitsInByte {
		if err := w.writer.WriteByte(byte((w.buffer << w.capacity) & 0xFF)); err != nil {
			return err
		}
		w.capacity = BitsInByte
		w.buffer = 0
		w.len++
	}

	return nil
}

func (w *BitWriter) WriteBit0() error {
	w.buffer <<= 1
	w.capacity--
	return w.flushBuffer()
}

func (w *BitWriter) WriteBit1() error {
	w.buffer = (w.buffer << 1) | 0x1
	w.capacity--
	return w.flushBuffer()
}

func (w *BitWriter) WriteBit(b int) error {
	w.buffer = (w.buffer << 1) | (b & 0x1)
	w.capacity--
	return w.flushBuffer()
}

func (w *BitWriter) WriteBits(b, n int) error {
	if n <= w.capacity {
		// all bits fit into the current buffer
		w.buffer = (w.buffer << n) | (b & (0xFF >> (BitsInByte - n)))
		w.capacity -= n
		if w.capacity == 0 {
			if err := w.writer.WriteByte(byte(w.buffer & 0xFF)); err != nil {
				return err
			}
			w.capacity = BitsInByte
			w.len++
		}
	} else {
		// fill as many bits into buffer as possible
		w.buffer = (w.buffer << w.capacity) | (int(uint32(b)>>(n-w.capacity)) & (0xFF >> (BitsInByte - w.capacity)))
		n -= w.capacity
		if err := w.writer.WriteByte(byte(w.buffer & 0xFF)); err != nil {
			return err
		}
		w.len++

		// possibly write whole bytes
		for n >= 8 {
			n -= 8
			if err := w.writer.WriteByte(byte(int(uint32(b) >> n))); err != nil {
				return err
			}
			w.len++
		}

		// put the rest of bits into the buffer
		w.buffer = b // Note: the high bits will be shifted out during further filling
		w.capacity = BitsInByte - n
	}

	return nil
}

func (w *BitWriter) writeDirectByte(b int) error {
	if err := w.writer.WriteByte(byte(b & 0xFF)); err != nil {
		return err
	}
	w.len++
	return nil
}

func (w *BitWriter) writeDirectBytes(b []byte, offset, length int) error {
	if _, err := w.writer.Write(b[offset : offset+length]); err != nil {
		return err
	}
	w.len += length // Due to misleading variable naming this code is confusing is Java source.
	return nil
}

func (w *BitWriter) Write(b int) error {
	return w.WriteBits(b, 8)
}
