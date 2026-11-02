// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/cheggaaa/pb.v1"
)

const displayWidth = 79

// progressBar is an instance of ReadCloser with an associated ProgressBar.
// Closing this instance closes the delegate as well as clearing the terminal
// line of progress output.
type progressBar struct {
	r   io.ReadCloser
	bar *pb.ProgressBar
}

// WrapInputFile creates an instance of os.File with an associated
// ProgressBar that tracks the bytes read relative to the total.
func WrapInputFile(f *os.File) (io.ReadCloser, error) {
	if f == os.Stdin {
		// don't bother wrapping stdin
		return os.Stdin, nil
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	total := int(fi.Size())

	bar := pb.New(total).SetUnits(pb.U_BYTES_DEC).SetWidth(displayWidth)
	bar.Output = os.Stderr
	bar.Start()

	return progressBar{
		r:   bar.NewProxyReader(f),
		bar: bar,
	}, nil
}

// Read implements io.Reader.Read by simple delegation.
func (b progressBar) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

// Close implements io.Closer.Close by closing the delegate instance of
// ReadCloser as well as clearing the terminal line of progress output.
func (b progressBar) Close() error {
	// make sure newline is not printed by Finish()
	b.bar.Output = nil
	b.bar.NotPrint = true

	b.bar.Finish()

	fmt.Fprintf(os.Stderr, "\033[2K\r") // clear status bar

	return b.r.Close()
}
