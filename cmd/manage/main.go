package main

import (
	"encoding/json"
	"fmt"
	"os"

	"devon-demo-articles/internal/model"
	"devon-demo-articles/internal/repository"
	"devon-demo-articles/internal/seed"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: manage init|seed|stats")
	}
	command := os.Args[1]
	if command != "init" && command != "seed" && command != "stats" {
		return fmt.Errorf("unknown command %q: use init, seed, or stats", command)
	}
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "data/articles.db"
	}
	r, err := repository.Open(path)
	if err != nil {
		return err
	}
	defer r.Close()
	if command != "stats" {
		if err := r.Migrate(); err != nil {
			return err
		}
	}
	if command == "seed" {
		if err := seed.Run(r); err != nil {
			return err
		}
	}
	if command == "init" {
		fmt.Println("database initialized")
		return nil
	}
	counts := map[string]int64{}
	for _, entry := range []struct {
		name  string
		table any
	}{{"users", &model.User{}}, {"articles", &model.Article{}}, {"comments", &model.Comment{}}, {"favorites", &model.Favorite{}}, {"tags", &model.Tag{}}} {
		var n int64
		if err := r.DB.Model(entry.table).Count(&n).Error; err != nil {
			return err
		}
		counts[entry.name] = n
	}
	return json.NewEncoder(os.Stdout).Encode(counts)
}
