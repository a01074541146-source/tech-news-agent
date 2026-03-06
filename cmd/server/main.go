package main

import (
	"context"
	"flag"
	"log"
	"net/http" // 추가됨: Render Health Check용
	"os"
	"os/signal"
	"syscall"
	"tech-news-agent/internal/config"
	"tech-news-agent/internal/services"

	"github.com/robfig/cron/v3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// --- 0. Render용 Health Check 서버 (추가됨) ---
	// 이 5줄이 없으면 Render가 10분 뒤에 서버를 강제로 꺼버립니다.
	go func() {
		port := os.Getenv("PORT")
		if port == "" { port = "10000" }
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Printf("[HealthCheck] Listening on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("⚠️ Health Check Server Error: %v", err)
		}
	}()

	// Command line flags
	testMode := flag.Bool("test", false, "Run once immediately for testing")
	testConnection := flag.Bool("test-connection", false, "Test connections only")
	flag.Parse()

	// Initialize logger
	logger := log.New(os.Stdout, "[TechNewsAgent] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 Tech News Agent starting...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Create news agent
	agent, err := services.NewNewsAgent(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to create news agent: %v", err)
	}
	defer agent.Close()

	// 1. 테스트 연결 모드
	if *testConnection {
		if err := agent.TestConnection(); err != nil { logger.Fatalf("Failed: %v", err) }
		return
	}

	// 2. 테스트 모드 (한 번 실행 후 종료 - 로컬 확인용)
	if *testMode {
		logger.Println("Running in test mode...")
		if err := agent.TestRun(); err != nil { logger.Fatalf("Failed: %v", err) }
		return
	}

	// 3. 실시간 메시지 수신 (Polling)
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		logger.Printf("⚠️ Telegram Polling connection failed: %v", err)
	} else {
		logger.Println("📡 Telegram Polling started! You can use /news now.")
		go func() {
			u := tgbotapi.NewUpdate(0)
			u.Timeout = 60
			updates := bot.GetUpdatesChan(u)
			for update := range updates {
				if update.Message == nil { continue }
				if update.Message.Text == "/news" || update.Message.Text == "/뉴스" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚀 실시간 뉴스를 분석하고 있습니다. 잠시만 기다려주세요!")
					bot.Send(msg)
					if err := agent.TestRun(); err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ 분석 에러: "+err.Error()))
					}
				}
			}
		}()
	}

	// 4. 기존 정기 스케줄러 (Cron)
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(logger)))
	_, err = c.AddFunc(cfg.CronSchedule, func() {
		logger.Println("Cron job triggered")
		if err := agent.Run(context.Background()); err != nil {
			logger.Printf("❌ Job failed: %v", err)
		}
	})
	c.Start()

	logger.Println("✅ Bot is fully operational and waiting for messages!")
	
	// 종료 시그널 대기
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Println("Shutting down gracefully...")
	c.Stop()
}
