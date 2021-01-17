# TGZ

Provides simple package to create and process tar.gz files.

Attributes
---
* Bytes - accepts a *bytes.Buffer as the source
* Tar - accepts a file or directory as the source
* Untar - unpacks a tar.gz file to the destination

```golang

	b := new(bytes.Buffer)
	b.WriteString("test file\nline1\nline2\n")

	h := sha256.New()
	w, _err_ := os.Create(filepath.Join(path, target))
	defer w.Close()

	tgz.Bytes(b, nil, w, h)

```

```golang

	h := sha256.New()
	w, _ := os.Create(target)
	defer w.Close()

    // specify specific header option settings
	tgz.Tar(path, &tar.Header{
        Gname:"server",
        Uname:"server",
        Mode:0644,
        ModTime:time.Now(),
        }, w, h)

```