/*
MIT License

Copyright (c) 2021 zxdev

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package tgz

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Bytes takes a bytes.Buffer and writes an archinve file. Pass opt as nil to
// use default value or specify Name, Gname, Uname, Mode, and ModTime in opt.
//
// Pass multiple writers to create an archive that duplicates writes to generate
// an archive as well as generate a md5 or sha25 hash at the same time.
func Bytes(b *bytes.Buffer, opt *tar.Header, w ...io.Writer) (int64, error) {

	// apply default options when nil is passed
	if opt == nil {
		opt = &tar.Header{
			Name:    time.Now().UTC().Format("20060102T150405"),
			Gname:   "user",
			Uname:   "user",
			Mode:    0644,
			ModTime: time.Now().UTC().Round(time.Second),
		}
	}

	// create a writer that duplicates its writes
	mw := io.MultiWriter(w...)

	gzw := gzip.NewWriter(mw) // compression
	defer gzw.Close()

	tw := tar.NewWriter(gzw) // tarball
	defer tw.Close()

	// write a header to the tarball archive
	tw.WriteHeader(&tar.Header{
		Name:    opt.Name,
		Size:    int64(b.Len()),
		Uname:   opt.Uname,
		Gname:   opt.Gname,
		Mode:    opt.Mode,
		ModTime: opt.ModTime,
	})

	// copy bytes to archive
	return io.Copy(tw, b)

}

// Tar takes a source path along with one or more writers and then writes the file
// or walks the directory writing each file found to the tar writer. Pass opt as nil
// to use defaults, opt will accept custom Gname, Uname, Mode, and ModTime for custom
// header settings for the file header. Execuable files are always ignored.
//
// Pass multiple writers to create an archive that duplicates its writes go generate
// an archive as well as generate a md5 or sha25 hash at the same time.
func Tar(src string, opt *tar.Header, writers ...io.Writer) error {

	// apply default options when nil is passed
	if opt == nil {
		opt = &tar.Header{Mode: 0644, Gname: "user", Uname: "user"}
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// create a writer that duplicates its writes
	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw) // compression
	defer gzw.Close()

	tw := tar.NewWriter(gzw) // tarball
	defer tw.Close()

	// path is a single file not a directory
	if !info.IsDir() {

		// fail when mode bits are set; no executables
		if !info.Mode().IsRegular() {
			return nil
		}

		// write a header to the tarball archive
		tw.WriteHeader(&tar.Header{
			Name:  filepath.Base(src),
			Size:  int64(info.Size()),
			Uname: opt.Uname,
			Gname: opt.Gname,
			Mode:  opt.Mode,
		})

		// copy the file source
		f, _ := os.Open(src)
		_, err = io.Copy(tw, f)
		f.Close()

		return err
	}

	// walk path and all sub directory tree
	return filepath.Walk(src, func(file string, info os.FileInfo, err error) error {

		// walk failed, so we fail too
		if err != nil {
			return err
		}

		// fail when mode bits are set, no executables
		if !info.Mode().IsRegular() {
			return nil
		}

		// create a new file header for the archive
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Gname = opt.Gname // set group
		header.Uname = opt.Uname // set user
		header.Mode = opt.Mode   // set permissions

		// use updated modifcation time
		if !opt.ModTime.IsZero() {
			header.AccessTime = opt.ModTime
			header.ChangeTime = opt.ModTime
			header.ModTime = opt.ModTime
		}

		// utilize an updated name for the correct path when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the file header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// copy the file source
		f, _ := os.Open(file)
		_, err = io.Copy(tw, f)
		f.Close()

		return err
	})
}

// Untar takes a destination path and an io.Reader that loops over the tarfile
// contents and will create the file structure within the destination
func Untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {

		header, err := tr.Next()
		switch {
		case header == nil:
			continue // what?! skip it

		case err == io.EOF:
			return nil

		case err != nil:
			return err
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:

			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}

		case tar.TypeReg:

			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
			f.Close()
		}
	}
}
