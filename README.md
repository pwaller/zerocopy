# zerocopy

A zero-copy `Reader` interface, which returns a byte slice pointing at the
underlying memory rather than copying it to you.

[Documentation](https://godoc.org/github.com/pwaller/zerocopy)

Zero copy streams of [`*bytes.Reader`](http://golang.org/pkg/bytes/#Reader) and [`*os.File`](http://golang.org/pkg/os/#File).

These are inherently unsafe since we violate the laws of the language to obtain
access to the underlying byte slice of the `*bytes.Reader`.

It's all in good fun, though?

Caveat Emptor.
