package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/issadarkthing/spirit/internal"
)

const (
	testDir = "../lib"
	libDir  = "../lib"
)

var _ internal.Scope = (*internal.Spirit)(nil)

func TestSpirit_Bind(t *testing.T) {
	sl := internal.NewSpirit()

	tests := []struct {
		name    string
		symbol  string
		ns      string
		wantErr bool
	}{
		{
			name:    "CrossNamespaceBindingValidation",
			symbol:  "core/not",
			ns:      "user",
			wantErr: true,
		},
		{
			name:    "BindingInCurrentNS",
			symbol:  "hello",
			ns:      "user",
			wantErr: false,
		},
		{
			name:    "UserBinding",
			symbol:  "user/hello",
			ns:      "user",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sl.Bind(tt.symbol, internal.Nil{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Bind() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSpirit_Resolve(t *testing.T) {
	sl := internal.NewSpirit()
	sl.Bind("pi", internal.Number(3.1412))

	tests := []struct {
		name    string
		symbol  string
		wantErr bool
	}{
		{
			name:   "CoreBinding",
			symbol: "core/impl?",
		},
		{
			name:    "UserBinding",
			symbol:  "hello",
			wantErr: true,
		},
		{
			name:    "MissingUserBinding",
			symbol:  "hello",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sl.Resolve(tt.symbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSpirit(t *testing.T) {
	if testing.Short() {
		return
	}

	t.Parallel()

	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}

	for _, fi := range files {
		if !strings.HasSuffix(fi.Name(), "_test.st") {
			continue
		}

		t.Run(fi.Name(), func(t *testing.T) {
			testFile(t, filepath.Join(testDir, fi.Name()))
		})
	}
}

func testFile(t *testing.T, file string) {
	fh, err := os.Open(file)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer fh.Close()

	sl, err := initspirit()
	if err != nil {
		t.Fatalf("failed to init spirit: %v", err)
	}

	_, err = sl.ReadEval(fh)
	if err != nil {
		t.Errorf("execution failed for '%s': %v", file, err)
	}
}

func initspirit() (*internal.Spirit, error) {
	di, err := ioutil.ReadDir(testDir)
	if err != nil {
		return nil, err
	}

	sl := internal.NewSpirit()
	for _, fi := range di {
		if !strings.HasSuffix(fi.Name(), ".st") ||
			strings.HasSuffix(fi.Name(), "_test.st") {
			continue
		}

		fh, err := os.Open(filepath.Join(testDir, fi.Name()))
		if err != nil {
			return nil, err
		}
		defer fh.Close()

		if _, err := sl.ReadEval(fh); err != nil {
			return nil, err
		}
	}

	return sl, nil
}

func BenchmarkBinaryTree(b *testing.B) {

	f, err := os.Open("../sample/tree.st")
	if err != nil {
		b.Error(err)
	}
	defer f.Close()

	spirit, err := initspirit()
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_, err := spirit.ReadEval(f)
		if err != nil {
			b.Error(err)
		}
	}

}

