package spectool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGoPackageNode_ImportFQDN(t *testing.T) {
	// module/a/b/c/
	module := GoMod{
		Name: "module",
		Path: "",
	}

	a := &GoPackage{
		Parent:      nil,
		Name:        "a",
		FilePath:    "",
		SubPackages: nil,
		ModFile:     module,
	}

	b := &GoPackage{
		Parent:      a,
		Name:        "b",
		FilePath:    "",
		SubPackages: nil,
		ModFile:     module,
	}
	a.SubPackages = append(a.SubPackages, b)

	c := &GoPackage{
		Parent:      b,
		Name:        "c",
		FilePath:    "",
		SubPackages: nil,
		ModFile:     module,
	}
	b.SubPackages = append(b.SubPackages, c)

	assert.Equal(t, "module/a/b/c", c.ImportPath())
	assert.Equal(t, "module/a/b", b.ImportPath())
	assert.Equal(t, "module/a", a.ImportPath())
}

func TestGoPackageNode_Root(t *testing.T) {
	// module/a/b/c/
	module := GoMod{
		Name: "module",
		Path: "",
	}

	a := &GoPackage{
		Parent:      nil,
		Name:        "a",
		FilePath:    "",
		SubPackages: nil,
		ModFile:     module,
	}

	b := &GoPackage{
		Parent:      a,
		Name:        "b",
		FilePath:    "",
		SubPackages: nil,
		ModFile:     module,
	}
	a.SubPackages = append(a.SubPackages, b)

	c := &GoPackage{
		Parent:      b,
		Name:        "c",
		FilePath:    "",
		SubPackages: nil,
		ModFile:     module,
	}
	b.SubPackages = append(b.SubPackages, c)

	assert.Equal(t, a, c.Root())
	assert.Equal(t, a, b.Root())
	assert.Equal(t, a, a.Root())
}
