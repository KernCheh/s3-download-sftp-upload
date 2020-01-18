package gzipdecompressor

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
)

// DecompressByteStream decompresses an incoming stream of bytes.
// Returns a io.ReadCloser with the decompressed stream
func DecompressByteStream(reader io.Reader, cancelFunc context.CancelFunc) (io.ReadCloser, error) {
	isGzip := false

	pr2, pw2 := io.Pipe()

	gzipDetermined := make(chan bool)

	go func() {
		defer pw2.Close()

		buf := make([]byte, 2)
		var bytesWritten int64

		// Read the first two bytes to check if file is gzipped.
		// Gzipped files have bytes 31 and 139 preamble
		nr, err := reader.Read(buf)
		if err != nil {
			fmt.Println("[GZip Decompressor] Error:", err.Error())
			cancelFunc()
			return
		}

		// Gzip check
		if nr > 0 {
			gzipPreamble := []byte{31, 139}
			if bytes.Equal(buf, gzipPreamble) {
				isGzip = true
			}
		}
		gzipDetermined <- true

		// Pipe the first two bytes and everything else into a new pipe
		nw, err := pw2.Write(buf)
		if err != nil {
			fmt.Println("[GZip Decompressor] Outgoing stream Write Error:", err.Error())
			cancelFunc()
			return
		}
		bytesWritten += int64(nw)

		// Drain the remaining bytes into the new pipe
		remainingBuf := make([]byte, 2<<15)
		for {
			nr, err = reader.Read(remainingBuf)
			if err != nil {
				if err != io.EOF {
					fmt.Println("[GZip Decompressor] Incoming stream Read Error:", err.Error())
					cancelFunc()
				}
				//EOF
				fmt.Printf("[GZip Decompressor] Outgoing stream Write Completed, %v bytes written\n", bytesWritten)
				return
			}

			nw, err = pw2.Write(remainingBuf[:nr])
			if err != nil {
				fmt.Println("[GZip Decompressor] Outgoing stream Write Error:", err.Error())
				cancelFunc()
				return
			}
			bytesWritten += int64(nw)
		}
	}()

	<-gzipDetermined
	if isGzip {
		fmt.Println("[GZip Decompressor] Gzipped content detected. Output stream will be unzipped bytes.")
		return gzip.NewReader(pr2)
	}
	return pr2, nil
}
