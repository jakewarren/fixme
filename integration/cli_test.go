// nolint: scopelint,gosec
package integration

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

var binaryName = "fixme"

func fixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), fixture)
}

func writeFixture(t *testing.T, fixture string, content []byte) {
	err := ioutil.WriteFile(fixturePath(t, fixture), content, 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func loadFixture(t *testing.T, fixture string) string {
	content, err := ioutil.ReadFile(fixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}

	return cleanPath(string(content))
}

// since the CI will likely have a different working path we need to massage the output a bit to remove the absolute path
func cleanPath(input string) string {
	pathRE := regexp.MustCompile(`(?m)^.+testdata/test(-author)?\.txt`)
	substitution := "testdata/test.txt"

	return pathRE.ReplaceAllString(input, substitution)
}

func TestCliArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		fixture string
	}{
		{"test.txt", []string{"./testdata/test.txt"}, "test-txt.golden"},
		{"test.txt JSON output", []string{"--json", "./testdata/test.txt"}, "test-txt-json.golden"},
		{"test authors", []string{"./testdata/test-author.txt"}, "test-author-txt.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(path.Join(dir, "bin", binaryName), tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("debug: dir: %s\n", dir)
				fmt.Printf("debug: cmd: %s\n", path.Join(dir, "bin", binaryName))
				fmt.Printf("debug: args: %v\n", tt.args)
				fmt.Printf("debug: output: %s\n", output)
				fmt.Printf("debug: error: %s\n", err)
				t.Fatal(err)
			}

			if *update {
				writeFixture(t, tt.fixture, output)
			}

			actual := cleanPath(string(output))

			expected := loadFixture(t, tt.fixture)

			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("actual = %s, expected = %s", actual, expected)
			}
		})
	}
}

func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}
	makeCmd := exec.Command("make", "build")
	err = makeCmd.Run()
	if err != nil {
		fmt.Printf("could not make binary for %s: %v", binaryName, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}
