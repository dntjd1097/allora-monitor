package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Service는 모니터링 서비스의 HTTP API를 제공합니다
type Service struct {
	monitor *Monitor
	db      *Database
}

// NewService는 새로운 서비스를 생성합니다
func NewService(monitor *Monitor, db *Database) *Service {
	return &Service{
		monitor: monitor,
		db:      db,
	}
}

// HandleGetCompetitions는 경쟁 데이터를 반환하는 핸들러입니다
func (s *Service) HandleGetCompetitions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 최신 데이터 조회
	data, err := s.db.GetLatestCompetitions()
	if err != nil {
		log.Printf("데이터 조회 실패: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	if data == nil {
		// 데이터가 없는 경우
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "No data available"}`))
		return
	}

	// JSON 응답 반환
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleGetCompetitionsByTimeRange는 지정된 시간 범위의 경쟁 데이터를 반환하는 핸들러입니다
func (s *Service) HandleGetCompetitionsByTimeRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 쿼리 파라미터에서 시작/종료 시간 추출
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			http.Error(w, "Invalid start time format. Use RFC3339 format.", http.StatusBadRequest)
			return
		}
	} else {
		// 기본값: 24시간 전
		start = time.Now().Add(-24 * time.Hour)
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			http.Error(w, "Invalid end time format. Use RFC3339 format.", http.StatusBadRequest)
			return
		}
	} else {
		// 기본값: 현재 시간
		end = time.Now()
	}

	// 데이터 조회
	data, err := s.db.GetCompetitionsByTimeRange(start, end)
	if err != nil {
		log.Printf("데이터 조회 실패: %v", err)
		http.Error(w, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	if len(data) == 0 {
		// 데이터가 없는 경우
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "No data available for the specified time range"}`))
		return
	}

	// JSON 응답 반환
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleGetDatabaseStats는 데이터베이스 통계를 반환하는 핸들러입니다
func (s *Service) HandleGetDatabaseStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 데이터베이스 통계 조회
	stats, err := s.db.GetDatabaseStats()
	if err != nil {
		log.Printf("통계 조회 실패: %v", err)
		http.Error(w, "Failed to retrieve statistics", http.StatusInternalServerError)
		return
	}

	// JSON 응답 반환
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleSetDirectURL은 직접 URL을 설정하는 핸들러입니다
func (s *Service) HandleSetDirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 요청 본문 읽기
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// JSON 파싱
	var requestData struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// URL 설정
	s.monitor.SetDirectURL(requestData.URL)

	// 응답 반환
	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Direct URL set to: %s", requestData.URL),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleFetchNow는 즉시 데이터 수집을 요청하는 핸들러입니다
func (s *Service) HandleFetchNow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 모니터링 서비스 실행 상태 확인
	if !s.monitor.IsRunning() {
		http.Error(w, "Monitoring service is not running", http.StatusServiceUnavailable)
		return
	}

	// 데이터 수집 요청
	go s.monitor.collectData()

	// 응답 반환
	response := map[string]string{
		"status":  "success",
		"message": "Data collection requested",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetActiveTopics는 활성 토픽 목록을 반환하는 핸들러입니다
func (s *Service) HandleGetActiveTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 활성 토픽 목록 조회
	activeTopics := s.monitor.GetTopicInferenceStore().GetActiveTopics()

	// 응답 반환
	response := map[string]interface{}{
		"status":        "success",
		"active_topics": activeTopics,
		"count":         len(activeTopics),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetTopicInference는 토픽 추론 데이터를 반환하는 핸들러입니다
func (s *Service) HandleGetTopicInference(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 쿼리 파라미터에서 토픽 ID 추출
	topicID := r.URL.Query().Get("topic_id")
	if topicID == "" {
		http.Error(w, "Missing topic_id parameter", http.StatusBadRequest)
		return
	}

	// 쿼리 파라미터에서 블록 높이 추출 (선택적)
	height := r.URL.Query().Get("height")

	var inference map[string]interface{}
	var err error
	var currentHeight string

	if height != "" {
		// 특정 블록 높이에 대한 토픽 추론 데이터 조회
		inference, err = s.db.GetTopicInferenceByHeight(topicID, height)
		if err != nil {
			log.Printf("토픽 %s의 블록 높이 %s 데이터 조회 실패: %v", topicID, height, err)
			http.Error(w, fmt.Sprintf("Failed to retrieve data for topic %s at height %s", topicID, height), http.StatusInternalServerError)
			return
		}
		currentHeight = height
	} else {
		// 최신 토픽 추론 데이터 조회
		inference, err = s.db.GetLatestTopicInference(topicID)
		if err != nil {
			log.Printf("토픽 %s 데이터 조회 실패: %v", topicID, err)
			http.Error(w, fmt.Sprintf("Failed to retrieve data for topic %s", topicID), http.StatusInternalServerError)
			return
		}
		// 최신 데이터의 블록 높이 가져오기
		if inference != nil && inference["inference_block_height"] != nil {
			currentHeight = inference["inference_block_height"].(string)
		}
	}

	if inference == nil {
		// 데이터가 없는 경우
		if height != "" {
			http.Error(w, fmt.Sprintf("No data available for topic %s at height %s", topicID, height), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("No data available for topic %s", topicID), http.StatusNotFound)
		}
		return
	}

	// 리더보드 데이터 가져오기
	// 토픽 ID에서 경쟁 ID 추출 (competitions_v2 테이블 사용)
	// competitionID, err := s.db.GetCompetitionIDFromTopicID(topicID)
	// if err != nil {
	// 	log.Printf("경쟁 ID 조회 실패: %v", err)
	// 	// 오류가 발생해도 계속 진행
	// }

	// // 리더보드 데이터를 저장할 슬라이스
	var leaderboardEntries []map[string]interface{}

	// synthesis_value 필드에 리더보드 데이터 추가
	if len(leaderboardEntries) > 0 {
		if networkInferences, ok := inference["network_inferences"].(map[string]interface{}); ok {
			// network_inferences가 있는 경우, synthesis_value를 리더보드 데이터로 대체
			networkInferences["synthesis_value"] = leaderboardEntries
		} else {
			// network_inferences가 없는 경우
			inference["network_inferences"] = map[string]interface{}{
				"synthesis_value": leaderboardEntries,
			}
		}
	}

	// 응답 데이터 구성
	responseData := make(map[string]interface{})
	for k, v := range inference {
		// prev_height와 next_height는 최상위 레벨로 이동
		if k == "prev_height" || k == "next_height" {
			continue
		}
		responseData[k] = v
	}

	// 응답 반환
	response := map[string]interface{}{
		"status": "success",
		"data":   responseData,
		"pagination": map[string]interface{}{
			"prev_height":    inference["prev_height"],
			"next_height":    inference["next_height"],
			"current_height": currentHeight,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetAllTopicInferences는 모든 토픽의 추론 데이터를 반환하는 핸들러입니다
func (s *Service) HandleGetAllTopicInferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 활성 토픽 목록 가져오기
	activeTopics := s.monitor.GetTopicInferenceStore().GetActiveTopics()

	// 각 토픽에 대해 가공된 데이터 조회
	inferences := make(map[string]interface{})
	for _, topicID := range activeTopics {
		inference, err := s.db.GetLatestTopicInference(topicID)
		if err != nil {
			log.Printf("토픽 %s 데이터 조회 실패: %v", topicID, err)
			continue
		}

		if inference != nil {
			inferences[topicID] = inference
		}
	}

	// 응답 반환
	response := map[string]interface{}{
		"status":     "success",
		"count":      len(inferences),
		"inferences": inferences,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleForceCollectTopicData는 특정 토픽의 추론 데이터를 강제로 수집하는 핸들러입니다
func (s *Service) HandleForceCollectTopicData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 요청 본문 읽기
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// JSON 파싱
	var requestData struct {
		TopicID string `json:"topic_id"`
	}
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if requestData.TopicID == "" {
		http.Error(w, "Missing topic_id parameter", http.StatusBadRequest)
		return
	}

	// 토픽 추론 데이터 강제 수집
	if err := s.monitor.ForceCollectTopicData(requestData.TopicID); err != nil {
		log.Printf("토픽 %s 데이터 강제 수집 실패: %v", requestData.TopicID, err)
		http.Error(w, fmt.Sprintf("Failed to collect data for topic %s: %v", requestData.TopicID, err), http.StatusInternalServerError)
		return
	}

	// 응답 반환
	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Data collection for topic %s requested", requestData.TopicID),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAddActiveTopic은 활성 토픽을 추가하는 핸들러입니다
func (s *Service) HandleAddActiveTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 요청 본문 읽기
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// JSON 파싱
	var requestData struct {
		TopicID string `json:"topic_id"`
	}
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if requestData.TopicID == "" {
		http.Error(w, "Missing topic_id parameter", http.StatusBadRequest)
		return
	}

	// 활성 토픽 추가
	s.monitor.GetTopicInferenceStore().AddActiveTopic(requestData.TopicID)

	// 응답 반환
	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Topic %s added to active topics", requestData.TopicID),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleRemoveActiveTopic은 활성 토픽을 제거하는 핸들러입니다
func (s *Service) HandleRemoveActiveTopic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 요청 본문 읽기
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// JSON 파싱
	var requestData struct {
		TopicID string `json:"topic_id"`
	}
	if err := json.Unmarshal(body, &requestData); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if requestData.TopicID == "" {
		http.Error(w, "Missing topic_id parameter", http.StatusBadRequest)
		return
	}

	// 활성 토픽 제거
	s.monitor.GetTopicInferenceStore().RemoveActiveTopic(requestData.TopicID)

	// 응답 반환
	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Topic %s removed from active topics", requestData.TopicID),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetTopicStats는 특정 토픽의 통계를 반환하는 핸들러입니다
func (s *Service) HandleGetTopicStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 쿼리 파라미터에서 토픽 ID 추출
	topicID := r.URL.Query().Get("topic_id")
	if topicID == "" {
		http.Error(w, "Missing topic_id parameter", http.StatusBadRequest)
		return
	}

	// 토픽 통계 조회
	stats, err := s.db.GetTopicStats(topicID)
	if err != nil {
		log.Printf("토픽 %s 통계 조회 실패: %v", topicID, err)
		http.Error(w, fmt.Sprintf("Failed to retrieve statistics for topic %s", topicID), http.StatusInternalServerError)
		return
	}

	// 응답 반환
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetTopicBlockHeights는 특정 토픽의 블록 높이 리스트를 반환하는 핸들러입니다
func (s *Service) HandleGetTopicBlockHeights(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 쿼리 파라미터에서 토픽 ID 추출
	topicID := r.URL.Query().Get("topic_id")
	if topicID == "" {
		http.Error(w, "Missing topic_id parameter", http.StatusBadRequest)
		return
	}

	// 페이지네이션 파라미터 추출
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100 // 기본값
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // 기본값
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// 블록 높이 리스트 조회
	result, err := s.db.GetTopicBlockHeights(topicID, limit, offset)
	if err != nil {
		log.Printf("토픽 %s 블록 높이 조회 실패: %v", topicID, err)
		http.Error(w, fmt.Sprintf("Failed to retrieve block heights for topic %s", topicID), http.StatusInternalServerError)
		return
	}

	// 응답 형식 변경
	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"topic_id":      result["topic_id"],
			"total_count":   result["total_count"],
			"limit":         result["limit"],
			"offset":        result["offset"],
			"heights":       result["heights"],
			"heights_count": result["heights_count"],
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetCompetitionsV2는 새로운 형식의 경쟁 데이터를 반환하는 핸들러입니다
func (s *Service) HandleGetCompetitionsV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 쿼리 파라미터에서 active 필터 확인
	activeOnly := r.URL.Query().Get("active") == "true"

	var competitions []CompetitionV2
	var err error

	if activeOnly {
		// 활성 경쟁만 조회
		competitions, err = s.db.GetActiveCompetitionsV2()
	} else {
		// 모든 경쟁 조회
		competitions, err = s.db.GetCompetitionsV2()
	}

	if err != nil {
		log.Printf("경쟁 데이터 조회 실패: %v", err)
		http.Error(w, "Failed to retrieve competition data", http.StatusInternalServerError)
		return
	}

	if len(competitions) == 0 {
		// 데이터가 없는 경우
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": "No competition data available"}`))
		return
	}

	// JSON 응답 반환
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(competitions)
}

// HandleGetCompetitionByTopicID는 토픽 ID에 해당하는 경쟁 정보를 반환하는 핸들러입니다
func (s *Service) HandleGetCompetitionByTopicID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 쿼리 파라미터에서 토픽 ID 추출
	topicID := r.URL.Query().Get("topic_id")
	if topicID == "" {
		http.Error(w, "Missing topic_id parameter", http.StatusBadRequest)
		return
	}

	// 토픽 ID에 해당하는 경쟁 정보 조회
	competition, err := s.db.GetCompetitionByTopicID(topicID)
	if err != nil {
		log.Printf("토픽 ID %s의 경쟁 정보 조회 실패: %v", topicID, err)
		http.Error(w, fmt.Sprintf("Failed to retrieve competition for topic %s: %v", topicID, err), http.StatusInternalServerError)
		return
	}

	// 응답 반환
	response := map[string]interface{}{
		"status":      "success",
		"competition": competition,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
