package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dntjd1097/allora-monitor/internal/app"
)

// CORS middleware to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// 명령줄 인자 파싱
	configPath := flag.String("config", "config.json", "설정 파일 경로")
	flag.Parse()

	// 설정 로드
	config, err := app.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("설정 로드 실패: %v", err)
	}

	// 데이터 디렉토리 생성
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		log.Fatalf("데이터 디렉토리 생성 실패: %v", err)
	}

	// 데이터 디렉토리 권한 확인
	dbPath := filepath.Join(config.DataDir, "allora-monitor.db")

	// 데이터베이스 파일이 이미 존재하는 경우 권한 확인
	if _, err := os.Stat(dbPath); err == nil {
		// 파일이 존재하면 쓰기 권한 확인
		file, err := os.OpenFile(dbPath, os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("데이터베이스 파일 쓰기 권한 확인 실패: %v", err)
			log.Printf("데이터베이스 파일 권한을 수정합니다...")

			// 파일 권한 변경 시도
			if err := os.Chmod(dbPath, 0644); err != nil {
				log.Printf("데이터베이스 파일 권한 변경 실패: %v", err)

				// 파일 백업 및 새로 생성
				backupPath := dbPath + ".bak." + time.Now().Format("20060102150405")
				log.Printf("데이터베이스 파일을 백업하고 새로 생성합니다: %s", backupPath)

				if err := os.Rename(dbPath, backupPath); err != nil {
					log.Fatalf("데이터베이스 파일 백업 실패: %v", err)
				}
			} else {
				if file != nil {
					file.Close()
				}
			}
		} else {
			file.Close()
		}
	}

	// 데이터베이스 연결
	db, err := app.NewDatabase(dbPath)
	if err != nil {
		log.Fatalf("데이터베이스 연결 실패: %v", err)
	}
	defer db.Close()

	// API 클라이언트 생성
	apiClient := app.NewAlloraAPIClient(config.AlloraBaseURL, time.Duration(config.APITimeoutSeconds)*time.Second)

	// 모니터링 서비스 생성
	monitor := app.NewMonitor(apiClient, db, config)

	// 서비스 생성
	service := app.NewService(monitor, db)

	// 모니터링 서비스 시작
	if err := monitor.Start(); err != nil {
		log.Fatalf("모니터링 서비스 시작 실패: %v", err)
	}

	// HTTP 서버 설정
	mux := http.NewServeMux()

	// API 라우트 등록
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to Allora Monitor API", "status": "running"}`))
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		monitorStatus := "stopped"
		if monitor.IsRunning() {
			monitorStatus = "running"
		}
		w.Write([]byte(`{"status": "healthy", "monitor_status": "` + monitorStatus + `", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
	})
	mux.HandleFunc("/api/competitions", service.HandleGetCompetitions)
	mux.HandleFunc("/api/competitions/range", service.HandleGetCompetitionsByTimeRange)
	mux.HandleFunc("/api/competitions/v2", service.HandleGetCompetitionsV2)
	mux.HandleFunc("/api/stats", service.HandleGetDatabaseStats)
	// mux.HandleFunc("/api/set-direct-url", service.HandleSetDirectURL)
	// mux.HandleFunc("/api/fetch-now", service.HandleFetchNow)

	// 토픽 추론 데이터 API 엔드포인트 등록
	mux.HandleFunc("/api/topics/active", service.HandleGetActiveTopics)
	mux.HandleFunc("/api/topics/inference", service.HandleGetTopicInference)
	mux.HandleFunc("/api/topics/inferences", service.HandleGetAllTopicInferences)
	// mux.HandleFunc("/api/topics/collect", service.HandleForceCollectTopicData)
	// mux.HandleFunc("/api/topics/add", service.HandleAddActiveTopic)
	// mux.HandleFunc("/api/topics/remove", service.HandleRemoveActiveTopic)
	mux.HandleFunc("/api/topics/stats", service.HandleGetTopicStats)
	mux.HandleFunc("/api/topics/heights", service.HandleGetTopicBlockHeights)

	// Apply CORS middleware
	handler := corsMiddleware(mux)

	// HTTP 서버 생성
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: handler,
	}

	// 서버를 고루틴에서 시작
	go func() {
		log.Printf("HTTP 서버가 포트 %s에서 시작되었습니다", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP 서버 실행 실패: %v", err)
		}
	}()

	// 종료 신호 처리
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("종료 신호를 받았습니다. 서버를 종료합니다...")

	// 모니터링 서비스 중지
	if err := monitor.Stop(); err != nil {
		log.Printf("모니터링 서비스 중지 실패: %v", err)
	}

	// HTTP 서버 종료
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("서버 종료 실패: %v", err)
	}

	log.Println("서버가 정상적으로 종료되었습니다")
}
