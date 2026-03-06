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

	// 텔레그램 채팅 ID 확인
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chat id: %v", err)
	}

	// 최대 뉴스 기사 수 설정
	maxArticles := 10 // 기본값 하향 조정 (에러 방지)
	if max := os.Getenv("MAX_NEWS_ARTICLES"); max != "" {
		if parsed, err := strconv.Atoi(max); err == nil {
			maxArticles = parsed
		}
	}

	// 크론 스케줄 (기본값: 매일 오전 9시)
	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		cronSchedule = "0 9 * * *" 
	}

	// 제미나이 모델 설정 (404 에러 방지를 위해 가장 표준적인 이름 사용)
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
