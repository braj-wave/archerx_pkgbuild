// package_test.go - Tests for package.go
//
// Copyright (c) 2013 The go-alpm Authors
//
// MIT Licensed. See LICENSE for details.

package alpm

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"
	"time"
)

// Auxiliary formatting
const pkginfo_template = `
Name         : {{ .Name }}
Version      : {{ .Version }}
Architecture : {{ .Architecture }}
Description  : {{ .Description }}
URL          : {{ .URL }}
Groups       : {{ .Groups.Slice }}
Licenses     : {{ .Licenses.Slice }}
Dependencies : {{ range .Depends.Slice }}{{ . }} {{ end }}
Provides     : {{ range .Provides.Slice }}{{ . }} {{ end }}
Replaces     : {{ range .Replaces.Slice }}{{ . }} {{ end }}
Conflicts    : {{ range .Conflicts.Slice }}{{ . }} {{ end }}
Packager     : {{ .Packager }}
Build Date   : {{ .PrettyBuildDate }}
Install Date : {{ .PrettyInstallDate }}
Package Size : {{ .Size }} bytes
Install Size : {{ .ISize }} bytes
MD5 Sum      : {{ .MD5Sum }}
SHA256 Sum   : {{ .SHA256Sum }}
Reason       : {{ .Reason }}

Required By  : {{ .ComputeRequiredBy }}
Files        : {{ range .Files }}
               {{ .Name }} {{ .Size }}{{ end }}
`

var pkginfoTemplate *template.Template

type PrettyPackage struct {
	*Package
}

func (p PrettyPackage) PrettyBuildDate() string {
	return p.BuildDate().Format(time.RFC1123)
}

func (p PrettyPackage) PrettyInstallDate() string {
	return p.InstallDate().Format(time.RFC1123)
}

func init() {
	var er error
	pkginfoTemplate, er = template.New("info").Parse(pkginfo_template)
	if er != nil {
		fmt.Printf("couldn't compile template: %s\n", er)
		panic("template parsing error")
	}
}

// Tests package attribute getters.
func TestPkginfo(t *testing.T) {
	h, er := Initialize(root, dbpath)
	defer h.Release()
	if er != nil {
		t.Errorf("Failed at alpm initialization: %s", er)
	}

	t.Log("Printing package information for pacman")
	db, _ := h.LocalDB()

	pkg := db.Pkg("glibc")
	buf := bytes.NewBuffer(nil)
	pkginfoTemplate.Execute(buf, PrettyPackage{pkg.(*Package)})
	t.Logf("%s...", buf.Bytes()[:1024])
	t.Logf("Should ignore %t", pkg.ShouldIgnore())

	pkg = db.Pkg("linux")
	if pkg != nil {
		buf = bytes.NewBuffer(nil)
		pkginfoTemplate.Execute(buf, PrettyPackage{pkg.(*Package)})
		t.Logf("%s...", buf.Bytes()[:1024])
		t.Logf("Should ignore %t", pkg.ShouldIgnore())
	}
}

func TestPkgNoExist(t *testing.T) {
	h, er := Initialize(root, dbpath)
	defer h.Release()
	if er != nil {
		t.Errorf("Failed at alpm initialization: %s", er)
	}

	db, _ := h.LocalDB()

	pkg := db.Pkg("non-existing-package-fa93f4af")
	if pkg != nil {
		t.Errorf("pkg should be nil but got %v", pkg)
	}
}

func TestPkgFiles(t *testing.T) {
	h, er := Initialize(root, dbpath)
	defer h.Release()
	if er != nil {
		t.Errorf("Failed at alpm initialization: %s", er)
	}

	db, _ := h.LocalDB()

	pkg := db.Pkg("glibc")
	_, err := pkg.ContainsFile("etc/locale.gen")
	if err != nil {
		t.Errorf("File should not be nil but got %v", err)
	}
	_, err = pkg.ContainsFile("etc/does-not-exist")
	if err == nil {
		t.Errorf("File should be nil but got %v", err)
	}
}
