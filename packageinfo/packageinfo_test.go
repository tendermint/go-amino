package packageinfo

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jaekwon/testify/assert"
)

type Foo struct {
	FieldA int
	FieldB string
}

func TestNewPackageInfo(t *testing.T) {
	// This should panic, as slashes in p3pkg is not allowed.
	assert.Panics(t, func() {
		NewPackageInfo("foobar.com/some/path", "some/path", "").WithTypes(Foo{})
	}, "slash in p3pkg should not be allowed")

	// This should panic, as the go pkg path includes a dot in the wrong place.
	assert.Panics(t, func() {
		NewPackageInfo("blah/foobar.com/some/path", "some.path", "").WithTypes(Foo{})
	}, "invalid go pkg path")

	// This should panic, as the go pkg path includes a leading slash.
	assert.Panics(t, func() {
		NewPackageInfo("/foobar.com/some/path", "some.path", "").WithTypes(Foo{})
	}, "invalid go pkg path")

	// This should panic, as the dirname is relative.
	assert.Panics(t, func() {
		NewPackageInfo("foobar.com/some/path", "some.path", "../someplace").WithTypes(Foo{})
	}, "invalid dirname")

	info := NewPackageInfo("foobar.com/some/path", "some.path", "")
	assert.NotNil(t, info)
}

func TestNameForType(t *testing.T) {
	// The Go package depends on how this test is invoked.
	// Sometimes it is "github.com/tendermint/go-amino/packageinfo_test".
	// Sometimes it is "command-line-arguments"
	// Sometimes it is "command-line-arguments_test"
	gopkg := reflect.TypeOf(Foo{}).PkgPath()
	info := NewPackageInfo(gopkg, "some.path", "").WithTypes(Foo{})

	assert.Equal(t, info.NameForType(reflect.TypeOf(Foo{})), "some.path.Foo")

	typeURL := info.TypeURLForType(reflect.TypeOf(Foo{}))
	assert.False(t, strings.Contains(typeURL[1:], "/"))
	assert.Equal(t, string(typeURL[0]), "/")
}

// If the struct wasn't registered, you can't get a name or type_url for it.
func TestNameForUnexpectedType(t *testing.T) {
	gopkg := reflect.TypeOf(Foo{}).PkgPath()
	info := NewPackageInfo(gopkg, "some.path", "")

	assert.Panics(t, func() {
		info.NameForType(reflect.TypeOf(Foo{}))
	})

	assert.Panics(t, func() {
		info.TypeURLForType(reflect.TypeOf(Foo{}))
	})
}
