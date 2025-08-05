package testcoverage

import (
	"os"
	"testing"
)

func TestHasCoveragePercent(t *testing.T) {
	dsnURI := "test.db"
	defer os.RemoveAll(dsnURI)
	repo := NewRepo(dsnURI)
	has, err := repo.HasCoveragePercent("hello", "asfafa")
	if err != nil {
		t.Fatal(err)
	}
	if has == true {
		t.Fatal("should not exist")
	}
}

func TestAddCoveragePercent(t *testing.T) {
	dsnURI := "test.db"
	defer os.RemoveAll(dsnURI)
	repo := NewRepo(dsnURI)
	err := repo.AddCoveragePercent("hello", "asfafa", 33.3)
	if err != nil {
		t.Fatal(err)
	}

	coveragePercent, err := repo.GetCoveragePercent("hello", "asfafa")
	if err != nil {
		t.Fatal(err)
	}

	if coveragePercent != 33.3 {
		t.Fatal("invalid coverage")
	}
}
