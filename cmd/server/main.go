package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"tech-news-agent/internal/config"
	"tech-news-agent/internal/services"

	"github.com/robfig/cron/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5" // 라이브러리 확인!
)

func main() {
	testMode := flag.Bool("test", false, "Run once immediately for testing")
	testConnection := flag.Bool("test-connection", false, "Test connections only")
	flag.Parse()

	logger := log.New(os.Stdout, "[TechNewsAgent] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 Tech News Agent starting...")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	agent, err := services.NewNewsAgent(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create news agent: %v", err)
	}
	defer agent.Close()

	// --- 1. 테스트 모드 로직 (기존 유지) ---
	if *testConnection {
		if err := agent.TestConnection(); err != nil { logger.Fatalf("Failed: %v", err) }
		return
	}

	if *testMode {
		logger.Println("Running in test mode...")
		if err := agent.TestRun(); err != nil { logger.Fatalf("Failed: %v", err) }
		return
	}

	// --- 2. 실시간 메시지 수신 (Polling) 추가 ---
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		logger.Printf("⚠️ Failed to connect Telegram for polling: %v", err)
	} else {
		logger.Println("📡 Telegram Polling started! Try sending '/news' in Telegram.")
		
		go func() {
			u := tgbotapi.NewUpdate(0)
			u.Timeout = 60
			updates := bot.GetUpdatesChan(u)

			for update := range updates {
				if update.Message == nil { continue }

				// 사용자가 메시지를 보내면 대답!
				if update.Message.Text == "/news" || update.Message.Text == "/뉴스" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚀 실시간 테크 뉴스를 분석 중입니다. 잠시만 기다려주세요...")
					bot.Send(msg)

					// 실시간 실행!
					if err := agent.TestRun(); err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ 에러 발생: "+err.Error()))
					}
				}
			}
		}()
	}

	// --- 3. 기존 Cron 스케줄러 (그대로 유지) ---
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))
	_, err = c.AddFunc(cfg.CronSchedule, func() {
		logger.Println("Cron job triggered")
		if err := agent.Run(context.Background()); err != nil {
			logger.Printf("❌ Job failed: %v", err)
		}
	})
	c.Start()

	logger.Println("✅ All systems go! Press Ctrl+C to stop")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down... Goodbye!")
}
