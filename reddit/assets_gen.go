// Code generated by "esc -o assets_gen.go -pkg reddit -ignore .go assets/"; DO NOT EDIT.

package reddit

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		_ = f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/assets/settings.html": {
		local:   "assets/settings.html",
		size:    3577,
		modtime: 1527261704,
		compressed: `
H4sIAAAAAAAC/7xXS2/jNhC++1dMiR5aYGlt3Z4WsoEgAYqeCjS9LyhxZLHgQyVHdgLD/72gHo7kWLLS
FpuDI5Hz+IbD+UZzOkkslEVgefXVo5SK2Pm8Wp1OhKbSgtqdEoVksD6fV2l8RA+5FiFsWSX2yNsltlsB
AKTlZvdHYwgKRBnSpNzsVmnSCu3emRYaPYXO+HecQ7L27gic71apVIfekXfH3sFgNXea6z3/adPtNfsB
c1LOXmSEl4PtFuMoiCgxDmIsvBkJkiKNbPcgJVg8NkG2MY6ULvGOVofQo63MyddbLqvREcJRaQ2VC9T8
BCi8MxDqrE1YAGXh1dUepAq58xIC+gP6NKlumv6zxDdlKBRqCSrAX3UgoBLBCoPgiub5Te4H6yDxSfRV
eGcpSij68RPgizCVxvAF2F4YDOwTMFNrUrmrCD17D6Fw3vTHEJ+5slpZZGCQSie3LEbJQDRp3LLECCv2
mJxO64ec1AF/rZWW69+ezueku7Lv42wcaZGhhsL5LbN45Jdg2O65f0yTRmjCgLJVTUCvFW4Z4QuxEezc
WfJOg3nhGwZKXntpTnLLBguVFjmWTkv0W9Zl66vymu1WiyLIS2EtarZ7bjIMj+37fBABNeZ0wdfbmAml
xX0RlIIE9/h3rTxW6E3gaDKUt93Fv1jiL9SB+72KaQwwSl63F+AzFEIHBBZp5yb8pMU/dUJZTeRsl6NQ
Z0a9ZSkjCxlZHuo8xxDAaP5LU7lp0qrdqI8knsd1MUt1GBBM0jHMN+ecx9p7tDQo/f+deUhkGnvJ9qX5
5R5D5WxQB+RGdmvBgMn456nyo4hk+pak5Kc3OwPDQqXyvvhTx4B5XxhLlB6a3IV54TSZghv1JgNNKZ70
tNmZGvp+H0sFvmzhHe/NaeV9aV0r9jU3q+6F3SOs2/bz6Gyh9jPy91M45PqjCPwgtJKCUN5n+zb+C8lH
/m84v66iATbveJ6+SyUl2p7olGRwELrGLet8LDFO8r7Q8ibS9o9LYfMeyJ0eMgT+XGfLkCdLoC+Pb9Bf
5nrLf28rS7rM2/Xv7/u9BjPdcL71QX6wlQXDf2bx28B8vHaexQGnm+C/Ricje/iPgpKoMYJ6av4vh3X/
/G9182Wk3l4wtDJOQxOqt5k9TZrm+OFPiHb2qUS8snH6uVLr9i8TTzshtftXc9P1iFU4Fz/EmxFr1QX1
TwAAAP//Pr3oe/kNAAA=
`,
	},

	"/": {
		isDir: true,
		local: "",
	},

	"/assets": {
		isDir: true,
		local: "assets",
	},
}
