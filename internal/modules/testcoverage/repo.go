package testcoverage

import (
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

type Coverage struct {
	ID      int64 `xorm:"pk autoincr"`
	Project string
	Hash    string
	Percent float64
}

type Repo struct {
	engine *xorm.Engine
}

func NewRepo(dsnURI string) *Repo {
	if err := os.MkdirAll(filepath.Dir(dsnURI), os.ModePerm); err != nil {
		log.Printf("Error creating directory %s: %v", filepath.Dir(dsnURI), err)
	}
	engine, err := xorm.NewEngine("sqlite", dsnURI)
	if err != nil {
		log.Fatalf("Error creating engine: %v", err)
	}
	err = engine.Sync(new(Coverage))
	if err != nil {
		log.Fatalf("Error syncing engine: %v", err)
	}
	return &Repo{
		engine,
	}
}

func (r *Repo) GetCoveragePercent(projectName string, hash string) (float64, error) {
	var coverage Coverage
	has, err := r.engine.Where("project = ? AND hash = ?", projectName, hash).Get(&coverage)
	if err != nil {
		return 0.0, err
	}
	if !has {
		return 0.0, nil
	}
	return coverage.Percent, nil
}

func (r *Repo) HasCoveragePercent(projectName string, hash string) (bool, error) {
	var coverage Coverage
	has, err := r.engine.Where("project = ? AND hash = ?", projectName, hash).Get(&coverage)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (r *Repo) AddCoveragePercent(projectName string, hash string, procent float64) error {
	_, err := r.engine.Table("coverage").Insert(&Coverage{
		Project: projectName,
		Hash:    hash,
		Percent: procent,
	})
	return err
}
