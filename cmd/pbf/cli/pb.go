// Copyright 2017 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"io"
	"os"

	pb "gopkg.in/cheggaaa/pb.v1"
)

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

	bar := pb.New(total).SetUnits(pb.U_BYTES_DEC).SetWidth(79)
	bar.Output = os.Stderr
	bar.Start()

	return progressBar{
		r:   bar.NewProxyReader(f),
		bar: bar,
	}, nil
}

// Read implements io.Reader.Read by simple delegation.
func (pb progressBar) Read(p []byte) (int, error) {
	return pb.r.Read(p)
}

// Close implements io.Closer.Close by closing the delegate instance of
// ReadCloser as well as clearing the terminal line of progress output.
func (pb progressBar) Close() error {

	// make sure newline is not printed by Finish()
	pb.bar.Output = nil
	pb.bar.NotPrint = true

	pb.bar.Finish()

	fmt.Fprintf(os.Stderr, "\033[2K\r") // clear status bar

	if err := pb.r.Close(); err != nil {
		return err
	}

	return nil
}
