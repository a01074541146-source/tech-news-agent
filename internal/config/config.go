package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiAPIKey     string
	TelegramBotToken string
	TelegramChatID   int64
	NewsAPIKey       string
	CronSchedule     string
	MaxNewsArticles  int
	GeminiModel      string
	NewsCategories   []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	chatID, err := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat id: %v", err)
	}

	maxArticles := 20
	if max := os.Getenv("MAX_NEWS_ARTICLES"); max != "" {
		if parsed, err := strconv.Atoi(max); err == nil {
			maxArticles = parsed
		}
	}

	// [수정됨] 환경 변수가 있으면 그걸 쓰고, 없으면 기본값(매일 9시) 사용
	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		cronSchedule = "0 9 * * *" 
	}

	// [수정됨] 환경 변수가 있으면 그걸 쓰고, 없으면 1.5-flash를 기본으로 사용
	geminiModel := os.Getenv("GEMINI_MODEL")
	if geminiModel == "" {
		geminiModel = "gemini-1.5-flash"
	}

	cfg := &Config{
		GeminiAPIKey:     os.Getenv("GEMINI_API_KEY"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   chatID,
		NewsAPIKey:       os.Getenv("NEWS_API_KEY"),
		CronSchedule:     cronSchedule,
		MaxNewsArticles:  maxArticles,
		GeminiModel:      geminiModel,
		NewsCategories:   []string{"technology", "science", "business"},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.GeminiAPIKey == "" { return fmt.Errorf("GEMINI_API_KEY is required") }
	if c.TelegramBotToken == "" { return fmt.Errorf("TELEGRAM_BOT_TOKEN is required") }
	if c.TelegramChatID == 0 { return fmt.Errorf("TELEGRAM_CHAT_ID is required") }
	if c.NewsAPIKey == "" { return fmt.Errorf("NEWS_API_KEY is required") }
	return nil
}
