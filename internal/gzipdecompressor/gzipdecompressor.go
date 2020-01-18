package gzipdecompressor

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
)

// DecompressByteStream
// Warning: This function is blocking until an EOF is given to reader
func DecompressByteStream(reader io.Reader, cancelFunc context.CancelFunc) (io.ReadCloser, error) {
	isGzip := false

	pr2, pw2 := io.Pipe()

	gzipDetermined := make(chan bool)

	go func() {
		buf := make([]byte, 2)

		nr, err := reader.Read(buf)
		var bytesWritten int64

		if err != nil {
			fmt.Println("[GZip Decompressor] Error:", err.Error())
			cancelFunc()
			return
		}

		// Gzip check
		if nr > 0 {
			gzipPressemble := []byte{31, 139}
			if bytes.Equal(buf, gzipPressemble) {
				isGzip = true
			}
			gzipDetermined <- true
		}

		nw, err := pw2.Write(buf)
		if err != nil {
			fmt.Println("[GZip Decompressor] Stream 2 Write Error:", err.Error())
			cancelFunc()
			return
		}
		bytesWritten += int64(nw)

		remainingBuf := bytes.Buffer{}
		_, err = remainingBuf.ReadFrom(reader)
		if err != nil {
			if err != io.EOF {
				fmt.Println("[GZip Decompressor] Stream 1 Read Error:", err.Error())
				cancelFunc()
				return
			}

			fmt.Println("[GZip Decompressor] Stream 1 EOF Reached")
			// fmt.Println(newbuf)
			return
		}

		go func() {
			defer pw2.Close()

			writtenBytes, err := remainingBuf.WriteTo(pw2)
			if err != nil {
				fmt.Println("[GZip Decompressor] Stream 2 Write Error:", err.Error())
				cancelFunc()
				return
			}
			bytesWritten += writtenBytes
			fmt.Printf("[GZip Decompressor] Stream 2 Write Completed, %v bytes written\n", bytesWritten)
		}()
	}()

	<-gzipDetermined
	if isGzip {
		fmt.Println("isgzip")
		return gzip.NewReader(pr2)
	}
	return pr2, nil
}
