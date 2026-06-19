package cmd

import (
	"path/filepath"
	"testing"
)

func TestDeriveTexName(t *testing.T) {
	cases := map[string]string{
		"cv_data.json":          "cv.tex",
		"cv_data_qa.json":       "cv_qa.tex",
		"cv_data_sysadmin.json": "cv_sysadmin.tex",
		"dir/cv_data_pm.json":   filepath.Join("dir", "cv_pm.tex"),
	}
	for in, want := range cases {
		if got := deriveTexName(in); got != want {
			t.Errorf("deriveTexName(%q) = %q, want %q", in, got, want)
		}
	}
}
