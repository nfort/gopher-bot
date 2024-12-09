package testcoverage

import (
	"os"
	"testing"
)

func TestHasCoverageProcent(t *testing.T) {
	dsnURI := "test.db"
	defer os.RemoveAll(dsnURI)
	repo := NewRepo(dsnURI)
	has, err := repo.HasCoverageProcent("hello", "asfafa")
	if err != nil {
		t.Fatal(err)
	}
	if has == true {
		t.Fatal("should not exist")
	}
}

func TestAddCoverageProcent(t *testing.T) {
	dsnURI := "test.db"
	defer os.RemoveAll(dsnURI)
	repo := NewRepo(dsnURI)
	err := repo.AddCoverageProcent("hello", "asfafa", 33.3)
	if err != nil {
		t.Fatal(err)
	}

	coverageProcent, err := repo.GetCoverageProcent("hello", "asfafa")
	if err != nil {
		t.Fatal(err)
	}

	if coverageProcent != 33.3 {
		t.Fatal("invalid coverage")
	}
}
