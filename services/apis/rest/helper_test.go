package rest

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
)

var update = flag.Bool("update", false, "update test data files")

func checkResponseWithTestDataFile(t *testing.T, responseBody []byte, ignoredFields []string) bool {
	_, filename, _, _ := runtime.Caller(1)
	_, testFile := filepath.Split(filename)
	testFile = strings.TrimSuffix(testFile, "_test.go")
	gp := filepath.Join("./test_data", testFile, t.Name()+".json")
	dir, _ := filepath.Split(gp)
	if *update && len(responseBody) > 0 {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				t.Fatalf("failed to directory for reponse file (%s): %s", dir, err)
			}
		}

		t.Logf("update reponse file: %s", gp)
		if err := ioutil.WriteFile(gp, responseBody, 0644); err != nil {
			t.Fatalf("failed to update reponse file (%s): %s", gp, err)
		}
	}

	g, err := ioutil.ReadFile(gp)
	if err != nil && len(responseBody) > 0 {
		t.Fatalf("failed reading testdata: %s", err)
	}

	ignoreCondtions := "last_modified_at|created_at"
	if len(ignoredFields) > 0 {
		for _, value := range ignoredFields {
			ignoreCondtions += "|" + value
		}
	}

	re, _ := regexp.Compile(`(\s*"(` + ignoreCondtions + `)":\s*"?\S*"?,?)`)
	actual := re.ReplaceAllString(string(responseBody), "")
	expected := re.ReplaceAllString(string(g), "")

	if actual != expected {
		diffFile := diff.LineDiff(expected, actual)
		scanner := bufio.NewScanner(strings.NewReader(diffFile))
		line := 1
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "yourstring") {
				break
			}

			line++
		}
		absf, _ := filepath.Abs(".")

		t.Errorf("result not as expected:\n%s:%d:1 \n%s\n", filepath.Join(absf, gp), line, diffFile)

		return false
	}

	return true
}
