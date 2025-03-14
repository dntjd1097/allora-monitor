package app

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dntjd1097/allora-monitor/pkg/utils"
)

// Monitor는 Allora 경쟁 데이터를 모니터링하는 서비스입니다
type Monitor struct {
	apiClient           *AlloraAPIClient
	db                  *Database
	config              *Config
	ticker              *time.Ticker
	stopChan            chan struct{}
	isRunning           bool
	runningMutex        sync.Mutex
	directURL           string               // 직접 사용할 URL (디버깅용)
	debug               bool                 // 디버깅 모드 활성화 여부
	topicInferenceStore *TopicInferenceStore // 토픽 추론 데이터 저장소
}

// NewMonitor는 새로운 모니터링 서비스를 생성합니다
func NewMonitor(apiClient *AlloraAPIClient, db *Database, config *Config) *Monitor {
	monitor := &Monitor{
		apiClient: apiClient,
		db:        db,
		config:    config,
		stopChan:  make(chan struct{}),
		debug:     true, // 디버깅 모드 활성화
	}

	// 토픽 추론 데이터 저장소 생성 (1분 간격으로 업데이트)
	monitor.topicInferenceStore = NewTopicInferenceStore(db, monitor, 1*time.Minute)

	return monitor
}

// SetDebug는 디버깅 모드를 설정합니다
func (m *Monitor) SetDebug(debug bool) {
	m.debug = debug
	m.apiClient.SetDebug(debug)
	m.db.SetDebug(debug)
	m.topicInferenceStore.SetDebug(debug)
}

// SetDirectURL은 직접 사용할 URL을 설정합니다 (디버깅용)
func (m *Monitor) SetDirectURL(url string) {
	m.directURL = url
	if m.debug {
		log.Printf("직접 URL 설정: %s", url)
	}
}

// Start는 모니터링 서비스를 시작합니다
func (m *Monitor) Start() error {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("모니터링 서비스가 이미 실행 중입니다")
	}

	// 모니터링 간격 설정
	interval := time.Duration(m.config.MonitoringIntervalMinutes) * time.Minute
	m.ticker = time.NewTicker(interval)
	m.isRunning = true

	if m.debug {
		log.Printf("모니터링 서비스 시작: 간격=%v", interval)
	}

	// 토픽 추론 데이터 저장소 시작
	if err := m.topicInferenceStore.Start(); err != nil {
		log.Printf("토픽 추론 데이터 저장소 시작 실패: %v", err)
	}

	// 즉시 첫 번째 데이터 수집 실행
	go func() {
		m.collectData()

		// 이후 정기적으로 데이터 수집
		for {
			select {
			case <-m.ticker.C:
				if m.debug {
					log.Printf("정기 데이터 수집 시작 (간격: %v)", interval)
				}
				m.collectData()
			case <-m.stopChan:
				m.ticker.Stop()
				if m.debug {
					log.Println("모니터링 루프 종료")
				}
				return
			}
		}
	}()

	log.Printf("모니터링 서비스가 시작되었습니다. 간격: %v", interval)
	return nil
}

// Stop은 모니터링 서비스를 중지합니다
func (m *Monitor) Stop() error {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()

	if !m.isRunning {
		return fmt.Errorf("모니터링 서비스가 실행 중이 아닙니다")
	}

	if m.debug {
		log.Println("모니터링 서비스 중지 요청")
	}

	// 토픽 추론 데이터 저장소 중지
	if err := m.topicInferenceStore.Stop(); err != nil {
		log.Printf("토픽 추론 데이터 저장소 중지 실패: %v", err)
	}

	close(m.stopChan)
	m.isRunning = false
	log.Println("모니터링 서비스가 중지되었습니다")
	return nil
}

// IsRunning은 모니터링 서비스의 실행 상태를 반환합니다
func (m *Monitor) IsRunning() bool {
	m.runningMutex.Lock()
	defer m.runningMutex.Unlock()
	return m.isRunning
}

// collectData는 Allora API에서 경쟁 데이터를 수집하고 데이터베이스에 저장합니다
func (m *Monitor) collectData() {
	startTime := time.Now()
	logMsg := utils.LogMessage("INFO", "데이터 수집 시작")
	log.Println(logMsg)

	if m.debug {
		log.Printf("데이터 수집 시작: 시간=%s", startTime.Format(time.RFC3339))
	}

	var resp *CompetitionsResponse
	var err error

	// API에서 데이터 가져오기
	if m.directURL != "" {
		// 직접 URL 사용 (디버깅용)
		log.Printf("직접 URL 사용: %s", m.directURL)
		resp, err = m.apiClient.DirectFetchCompetitions(m.directURL)
	} else {
		// 자동 URL 탐색
		if m.debug {
			log.Println("자동 URL 탐색 사용")
		}
		resp, err = m.apiClient.FetchCompetitions()
	}

	if err != nil {
		log.Printf("데이터 가져오기 실패: %v", err)
		return
	}

	if m.debug {
		// 응답 데이터 확인
		activeCount := len(resp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions)
		pastCount := len(resp.PageProps.CompetitionsPage.PastCompetitions)
		log.Printf("응답 데이터 확인: 활성 경쟁=%d, 과거 경쟁=%d", activeCount, pastCount)
	}

	// 데이터베이스에 저장
	if err := m.db.SaveCompetitions(resp); err != nil {
		log.Printf("데이터 저장 실패: %v", err)
		return
	}

	if m.debug {
		log.Println("데이터베이스 저장 완료")
	}

	// 활성 토픽 ID 목록 추출 및 설정
	activeTopicIDs := m.extractActiveTopicIDs(resp)
	m.topicInferenceStore.SetActiveTopics(activeTopicIDs)

	if m.debug {
		log.Printf("활성 토픽 ID 목록 설정: %v", activeTopicIDs)
	}

	// 오래된 데이터 정리 (설정된 보존 기간보다 오래된 데이터)
	retentionPeriod := time.Duration(m.config.DataRetentionDays) * 24 * time.Hour
	rowsDeleted, err := m.db.PruneOldData(retentionPeriod)
	if err != nil {
		log.Printf("오래된 데이터 정리 실패: %v", err)
	} else if rowsDeleted > 0 {
		log.Printf("%d개의 오래된 레코드가 정리되었습니다", rowsDeleted)
	}

	// 오래된 토픽 데이터 정리
	topicRowsDeleted, err := m.db.PruneOldTopicData("", retentionPeriod)
	if err != nil {
		log.Printf("오래된 토픽 데이터 정리 실패: %v", err)
	} else if topicRowsDeleted > 0 {
		log.Printf("%d개의 오래된 토픽 데이터가 정리되었습니다", topicRowsDeleted)
	}

	// 데이터베이스 통계 로깅
	if stats, err := m.db.GetDatabaseStats(); err == nil {
		compStats := stats["competitions"].(map[string]interface{})
		topicStats := stats["topic_inferences"].(map[string]interface{})

		log.Printf("DB 통계: 경쟁 레코드=%d (%.2fMB), 토픽 레코드=%d (%.2fMB)",
			compStats["record_count"], compStats["total_size_mb"],
			topicStats["record_count"], topicStats["total_size_mb"])
	}

	elapsedTime := time.Since(startTime)
	if m.debug {
		log.Printf("데이터 수집 완료: 소요 시간=%v", elapsedTime)
	}

	logMsg = utils.LogMessage("INFO", "데이터 수집 완료")
	log.Println(logMsg)
}

// extractActiveTopicIDs는 경쟁 데이터에서 활성 토픽 ID 목록을 추출합니다
func (m *Monitor) extractActiveTopicIDs(resp *CompetitionsResponse) []string {
	if resp == nil {
		return []string{}
	}

	// 중복 방지를 위한 맵 사용
	topicIDMap := make(map[string]bool)

	// 활성 및 예정된 경쟁에서 토픽 ID 추출
	for _, comp := range resp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions {
		// TopicID가 0인 경우 제외
		if comp.TopicID == 0 {
			if m.debug {
				log.Printf("TopicID가 0인 경쟁 제외: ID=%d", comp.TopicID)
			}
			continue
		}

		topicIDStr := fmt.Sprintf("%d", comp.TopicID)
		topicIDMap[topicIDStr] = true

		if m.debug {
			log.Printf("활성 토픽 발견: ID=%s", topicIDStr)
		}
	}

	// 맵에서 슬라이스로 변환
	result := make([]string, 0, len(topicIDMap))
	for topicID := range topicIDMap {
		result = append(result, topicID)
	}

	return result
}

// GetTopicInferenceStore는 토픽 추론 데이터 저장소를 반환합니다
func (m *Monitor) GetTopicInferenceStore() *TopicInferenceStore {
	return m.topicInferenceStore
}

// ForceCollectTopicData는 지정된 토픽의 추론 데이터를 강제로 수집합니다
func (m *Monitor) ForceCollectTopicData(topicID string) error {
	return m.topicInferenceStore.ForceCollectTopicData(topicID)
}
