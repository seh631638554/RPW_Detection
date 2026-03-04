package httpserver

import (
	dao "RPW_Detection/Dao"
	models "RPW_Detection/Models"
	"RPW_Detection/db"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

func TestLoginMySQLIntegration(t *testing.T) {
	dsn, maxOpenConns, maxIdleConns, connLifetime := mustLoadMySQLFromTestYAML(t)
	gdb, err := db.New(dsn, maxOpenConns, maxIdleConns, connLifetime)
	if err != nil {
		t.Fatalf("connect mysql failed: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close(gdb)
	})

	if err := gdb.Table("users").Limit(1).Find(&[]models.User{}).Error; err != nil {
		t.Fatalf("query users table failed (make sure table exists): %v", err)
	}

	username := "login_test_user_" + strings.ReplaceAll(time.Now().Format("20060102150405.000000"), ".", "")
	rawPassword := "pass123456"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}
	userToInsert := &models.User{
		Username:     username,
		PasswordHash: string(passwordHash),
		Status:       1,
	}
	// user := &models.User{
	// 	Username:     username,
	// 	PasswordHash: "123",
	// 	Status:       1,
	// }
	if err := gdb.Create(userToInsert).Error; err != nil {
		t.Fatalf("create test user failed: %v", err)
	}
	t.Cleanup(func() {
		gdb.Where("username = ?", username).Delete(&models.User{})
	})

	auth := NewAuthHandler(dao.New(gdb), &Config{
		JWT: JWTConfig{
			SecretKey:  "test-secret",
			ExpireTime: time.Hour,
		},
	})
	r := gin.New()
	r.POST("/api/v1/auth/login", auth.handleLogin)

	w := httptest.NewRecorder()
	reqBody := `{"username":"` + username + `","password":"` + "rawPassword" + `"}`
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"token"`)
	assert.Contains(t, w.Body.String(), username)

	w2 := httptest.NewRecorder()
	badReqBody := `{"username":"` + username + `","password":"bad-password"}`
	req2, _ := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(badReqBody))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusUnauthorized, w2.Code)
	assert.Contains(t, w2.Body.String(), "用户名或密码错误")
}

type testYAMLConfig struct {
	Database struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		Database     string `yaml:"database"`
		MaxOpenConns int    `yaml:"maxOpenConns"`
		MaxIdleConns int    `yaml:"maxIdleConns"`
		ConnLifetime string `yaml:"connLifetime"`
	} `yaml:"database"`
}

func mustLoadMySQLFromTestYAML(t *testing.T) (string, int, int, time.Duration) {
	t.Helper()

	ymlPath := filepath.Join("..", "config", "test.yml")
	raw, err := os.ReadFile(ymlPath)
	if err != nil {
		t.Fatalf("read %s failed: %v", ymlPath, err)
	}

	var cfg testYAMLConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse %s failed: %v", ymlPath, err)
	}

	if cfg.Database.Host == "" || cfg.Database.Port == 0 || cfg.Database.Username == "" || cfg.Database.Database == "" {
		t.Fatalf("database config is incomplete in %s", ymlPath)
	}

	connLifetime := time.Minute
	if cfg.Database.ConnLifetime != "" {
		if d, err := time.ParseDuration(cfg.Database.ConnLifetime); err == nil {
			connLifetime = d
		}
	}
	maxOpenConns := cfg.Database.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 5
	}
	maxIdleConns := cfg.Database.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 2
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
	print(dsn)
	return dsn, maxOpenConns, maxIdleConns, connLifetime
}
