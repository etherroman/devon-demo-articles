package seed_test

import (
	"devon-demo-articles/internal/model"
	"devon-demo-articles/internal/repository"
	"devon-demo-articles/internal/seed"
	"golang.org/x/crypto/bcrypt"
	"path/filepath"
	"testing"
)

func TestSeedIsRepeatableAndPasswordsAreHashed(t *testing.T) {
	r, err := repository.Open(filepath.Join(t.TempDir(), "seed.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if err := r.Migrate(); err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 2; run++ {
		if err := seed.Run(r); err != nil {
			t.Fatal(err)
		}
		for _, item := range []struct {
			table any
			want  int64
		}{{&model.User{}, 3}, {&model.Article{}, 5}, {&model.Comment{}, 6}, {&model.Favorite{}, 7}, {&model.Tag{}, 6}} {
			var count int64
			if err := r.DB.Model(item.table).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != item.want {
				t.Fatalf("run %d: want %d got %d", run, item.want, count)
			}
		}
	}
	user, err := r.UserByEmail("hana@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.PasswordHash == "DemoPass123!" || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("DemoPass123!")) != nil {
		t.Fatal("seed password not hashed")
	}
}
