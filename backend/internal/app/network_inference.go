package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	rpcaddress = "allora-rpc.testnet.allora.network"
	apiaddress = "allora-api.testnet.allora.network"
	version    = "v9"
)

// NetworkInference는 Allora 네트워크의 추론 데이터 구조체입니다
type NetworkInference struct {
	NetworkInferences                NetworkInferences `json:"network_inferences"`
	InfererWeights                   []InfererWeight   `json:"inferer_weights"`
	ForecasterWeights                []interface{}     `json:"forecaster_weights"`
	InferenceBlockHeight             string            `json:"inference_block_height"`
	LossBlockHeight                  string            `json:"loss_block_height"`
	ConfidenceIntervalRawPercentiles []string          `json:"confidence_interval_raw_percentiles"`
	ConfidenceIntervalValues         []string          `json:"confidence_interval_values"`
}

type InfererWeight struct {
	Worker string `json:"worker"`
	Weight string `json:"weight"`
}

type NetworkInferences struct {
	TopicID                       string              `json:"topic_id"`
	ReputerRequestNonce           ReputerRequestNonce `json:"reputer_request_nonce"`
	Reputer                       string              `json:"reputer"`
	ExtraData                     interface{}         `json:"extra_data"`
	CombinedValue                 string              `json:"combined_value"`
	InfererValues                 []InfererValue      `json:"inferer_values"`
	ForecasterValues              []interface{}       `json:"forecaster_values"`
	NaiveValue                    string              `json:"naive_value"`
	OneOutInfererValues           []InfererValue      `json:"one_out_inferer_values"`
	OneOutForecasterValues        []interface{}       `json:"one_out_forecaster_values"`
	OneInForecasterValues         []interface{}       `json:"one_in_forecaster_values"`
	OneOutInfererForecasterValues []interface{}       `json:"one_out_inferer_forecaster_values"`
}

type InfererValue struct {
	Worker string `json:"worker"`
	Value  string `json:"value"`
}

type ReputerRequestNonce struct {
	ReputerNonce ReputerNonce `json:"reputer_nonce"`
}

type ReputerNonce struct {
	BlockHeight string `json:"block_height"`
}

// TopicInferenceStore는 토픽별 네트워크 추론 데이터를 저장하는 저장소입니다
type TopicInferenceStore struct {
	inferences     map[string]*NetworkInference // topicID -> NetworkInference
	lastUpdated    map[string]time.Time         // topicID -> 마지막 업데이트 시간
	activeTopics   []string                     // 활성 토픽 ID 목록
	mu             sync.RWMutex
	db             *Database
	monitor        *Monitor // Monitor 인스턴스 추가
	updateInterval time.Duration
	stopChan       chan struct{}
	isRunning      bool
	runningMutex   sync.Mutex
	debug          bool
}

// NewTopicInferenceStore는 새로운 토픽 추론 데이터 저장소를 생성합니다
func NewTopicInferenceStore(db *Database, monitor *Monitor, updateInterval time.Duration) *TopicInferenceStore {
	return &TopicInferenceStore{
		inferences:     make(map[string]*NetworkInference),
		lastUpdated:    make(map[string]time.Time),
		activeTopics:   make([]string, 0),
		db:             db,
		monitor:        monitor,
		updateInterval: updateInterval,
		stopChan:       make(chan struct{}),
		debug:          true,
	}
}

// SetDebug는 디버깅 모드를 설정합니다
func (s *TopicInferenceStore) SetDebug(debug bool) {
	s.debug = debug
}

// SetActiveTopics는 활성 토픽 ID 목록을 설정합니다
func (s *TopicInferenceStore) SetActiveTopics(topicIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeTopics = make([]string, len(topicIDs))
	copy(s.activeTopics, topicIDs)

	if s.debug {
		log.Printf("활성 토픽 설정: %v", s.activeTopics)
	}
}

// AddActiveTopic은 활성 토픽 ID를 추가합니다
func (s *TopicInferenceStore) AddActiveTopic(topicID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 이미 존재하는지 확인
	for _, id := range s.activeTopics {
		if id == topicID {
			return
		}
	}

	s.activeTopics = append(s.activeTopics, topicID)

	if s.debug {
		log.Printf("활성 토픽 추가: %s, 현재 활성 토픽: %v", topicID, s.activeTopics)
	}
}

// RemoveActiveTopic은 활성 토픽 ID를 제거합니다
func (s *TopicInferenceStore) RemoveActiveTopic(topicID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, id := range s.activeTopics {
		if id == topicID {
			s.activeTopics = append(s.activeTopics[:i], s.activeTopics[i+1:]...)
			break
		}
	}

	if s.debug {
		log.Printf("활성 토픽 제거: %s, 현재 활성 토픽: %v", topicID, s.activeTopics)
	}
}

// GetActiveTopics는 활성 토픽 ID 목록을 반환합니다
func (s *TopicInferenceStore) GetActiveTopics() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]string, len(s.activeTopics))
	copy(result, s.activeTopics)
	return result
}

// Start는 토픽 추론 데이터 수집을 시작합니다
func (s *TopicInferenceStore) Start() error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if s.isRunning {
		return fmt.Errorf("토픽 추론 데이터 수집이 이미 실행 중입니다")
	}

	s.isRunning = true

	if s.debug {
		log.Printf("토픽 추론 데이터 수집 시작: 간격=%v", s.updateInterval)
	}

	// 즉시 첫 번째 데이터 수집 실행
	go func() {
		s.collectAllTopicData()

		// 이후 정기적으로 데이터 수집
		ticker := time.NewTicker(s.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if s.debug {
					log.Printf("정기 토픽 추론 데이터 수집 시작 (간격: %v)", s.updateInterval)
				}
				s.collectAllTopicData()
			case <-s.stopChan:
				if s.debug {
					log.Println("토픽 추론 데이터 수집 루프 종료")
				}
				return
			}
		}
	}()

	log.Printf("토픽 추론 데이터 수집이 시작되었습니다. 간격: %v", s.updateInterval)
	return nil
}

// Stop은 토픽 추론 데이터 수집을 중지합니다
func (s *TopicInferenceStore) Stop() error {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()

	if !s.isRunning {
		return fmt.Errorf("토픽 추론 데이터 수집이 실행 중이 아닙니다")
	}

	if s.debug {
		log.Println("토픽 추론 데이터 수집 중지 요청")
	}

	close(s.stopChan)
	s.isRunning = false
	log.Println("토픽 추론 데이터 수집이 중지되었습니다")
	return nil
}

// IsRunning은 토픽 추론 데이터 수집의 실행 상태를 반환합니다
func (s *TopicInferenceStore) IsRunning() bool {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()
	return s.isRunning
}

// collectAllTopicData는 모든 활성 토픽의 추론 데이터를 수집합니다
func (s *TopicInferenceStore) collectAllTopicData() {
	startTime := time.Now()

	if s.debug {
		log.Printf("모든 활성 토픽 데이터 수집 시작: 시간=%s", startTime.Format(time.RFC3339))
	}

	// 활성 토픽 목록 가져오기
	activeTopics := s.GetActiveTopics()

	if len(activeTopics) == 0 {
		if s.debug {
			log.Println("활성 토픽이 없습니다")
		}
		return
	}

	// 각 토픽에 대해 데이터 수집
	for _, topicID := range activeTopics {
		if err := s.collectTopicData(topicID); err != nil {
			log.Printf("토픽 %s 데이터 수집 실패: %v", topicID, err)
		}
	}

	elapsedTime := time.Since(startTime)
	if s.debug {
		log.Printf("모든 활성 토픽 데이터 수집 완료: 소요 시간=%v", elapsedTime)
	}
}

// createSynthesisData는 NetworkInference 데이터로부터 synthesis_data를 생성합니다
func (s *TopicInferenceStore) createSynthesisData(networkInference NetworkInference, topicID string) []map[string]interface{} {
	// 워커 맵 생성 (worker 주소를 키로 사용)
	workerMap := make(map[string]map[string]interface{})

	// inferer_values 처리
	for _, iv := range networkInference.NetworkInferences.InfererValues {
		if iv.Worker != "" {
			if _, exists := workerMap[iv.Worker]; !exists {
				workerMap[iv.Worker] = make(map[string]interface{})
			}
			workerMap[iv.Worker]["worker"] = iv.Worker
			workerMap[iv.Worker]["inferer_values"] = iv.Value
		}
	}

	// one_out_inferer_values 처리
	for _, oiv := range networkInference.NetworkInferences.OneOutInfererValues {
		if oiv.Worker != "" {
			if _, exists := workerMap[oiv.Worker]; !exists {
				workerMap[oiv.Worker] = make(map[string]interface{})
				workerMap[oiv.Worker]["worker"] = oiv.Worker
			}
			workerMap[oiv.Worker]["one_out_inferer_values"] = oiv.Value
		}
	}

	// inferer_weights 처리
	for _, iw := range networkInference.InfererWeights {
		if iw.Worker != "" {
			if _, exists := workerMap[iw.Worker]; !exists {
				workerMap[iw.Worker] = make(map[string]interface{})
				workerMap[iw.Worker]["worker"] = iw.Worker
			}
			workerMap[iw.Worker]["weight"] = iw.Weight
		}
	}

	// 경쟁 ID가 있는 경우에만 리더보드 데이터 가져오기
	if s.db != nil && s.monitor != nil {
		// 토픽 ID에 해당하는 경쟁 ID 가져오기
		competitionID, err := s.db.GetCompetitionIDFromTopicID(topicID)
		if err != nil {
			if s.debug {
				log.Printf("토픽 ID %s에 대한 경쟁 ID 조회 실패: %v", topicID, err)
			}
		} else if competitionID != "" {
			// 리더보드 데이터를 저장할 맵 (cosmos_address -> 리더보드 데이터)
			leaderboardMap := make(map[string]map[string]interface{})

			// 첫 번째 페이지 가져오기
			leaderboardResp, err := s.monitor.apiClient.FetchLeaderboard(competitionID, "")
			if err != nil {
				log.Printf("리더보드 데이터 조회 실패: %v", err)
			} else if leaderboardResp != nil && leaderboardResp.Status {
				// 첫 번째 페이지 데이터 추가
				for _, entry := range leaderboardResp.Data.Leaderboard {
					entryMap := map[string]interface{}{
						"rank":           entry.Rank,
						"cosmos_address": entry.CosmosAddress,
						"username":       entry.Username,
						"points":         entry.Points,
						"score":          entry.Score,
						"loss":           entry.Loss,
						"is_active":      entry.IsActive,
					}

					// FirstName과 LastName은 null일 수 있으므로 조건부로 추가
					if entry.FirstName != nil {
						entryMap["first_name"] = *entry.FirstName
					} else {
						entryMap["first_name"] = ""
					}

					if entry.LastName != nil {
						entryMap["last_name"] = *entry.LastName
					} else {
						entryMap["last_name"] = ""
					}

					// cosmos_address를 키로 사용하여 리더보드 맵에 저장
					leaderboardMap[entry.CosmosAddress] = entryMap
				}

				// continuation_token이 있으면 추가 페이지 가져오기
				continuationToken := leaderboardResp.Data.ContinuationToken
				for continuationToken != "" {
					nextPageResp, err := s.monitor.apiClient.FetchLeaderboard(competitionID, continuationToken)
					if err != nil {
						log.Printf("추가 리더보드 데이터 조회 실패: %v", err)
						break
					}

					if !nextPageResp.Status || len(nextPageResp.Data.Leaderboard) == 0 {
						break
					}

					// 추가 페이지 데이터 추가
					for _, entry := range nextPageResp.Data.Leaderboard {
						entryMap := map[string]interface{}{
							"rank":           entry.Rank,
							"cosmos_address": entry.CosmosAddress,
							"username":       entry.Username,
							"points":         entry.Points,
							"score":          entry.Score,
							"loss":           entry.Loss,
							"is_active":      entry.IsActive,
						}

						// FirstName과 LastName은 null일 수 있으므로 조건부로 추가
						if entry.FirstName != nil {
							entryMap["first_name"] = *entry.FirstName
						} else {
							entryMap["first_name"] = ""
						}

						if entry.LastName != nil {
							entryMap["last_name"] = *entry.LastName
						} else {
							entryMap["last_name"] = ""
						}

						// cosmos_address를 키로 사용하여 리더보드 맵에 저장
						leaderboardMap[entry.CosmosAddress] = entryMap
					}

					// 다음 페이지를 위한 토큰 업데이트
					continuationToken = nextPageResp.Data.ContinuationToken
				}
			}

			// 워커 데이터에 해당하는 리더보드 데이터 추가
			for worker, workerData := range workerMap {
				// worker 주소가 cosmos_address와 일치하는 리더보드 데이터 찾기
				if leaderboardEntry, exists := leaderboardMap[worker]; exists {
					// 리더보드 데이터를 워커 데이터에 추가
					workerData["leaderboard"] = leaderboardEntry
				}
			}
		}
	}

	// 신뢰 구간 백분위수 계산 및 합성 데이터 생성
	var synthesisValue []map[string]interface{}
	for _, workerData := range workerMap {
		// 신뢰 구간 백분위수 계산은 데이터베이스 조회 시 수행

		// 신뢰 구간 백분위수를 범위 형식으로 추가 (예: "84.13~97.72")
		if len(networkInference.ConfidenceIntervalRawPercentiles) >= 2 && len(networkInference.ConfidenceIntervalValues) >= 2 {
			infererValue, _ := workerData["inferer_values"].(string)
			if infererValue != "" {
				// infererValue를 float64로 변환
				infValue, err := strconv.ParseFloat(infererValue, 64)
				if err == nil {
					// 각 신뢰 구간 값을 float64로 변환
					var ciValues []float64
					for _, v := range networkInference.ConfidenceIntervalValues {
						val, err := strconv.ParseFloat(v, 64)
						if err == nil {
							ciValues = append(ciValues, val)
						}
					}

					// 값이 없으면 기본값 반환
					if len(ciValues) > 0 && len(networkInference.ConfidenceIntervalRawPercentiles) > 0 &&
						len(ciValues) == len(networkInference.ConfidenceIntervalRawPercentiles) {
						// 가장 가까운 두 개의 백분위수 찾기
						lowerIndex := -1
						upperIndex := -1

						// infererValue가 어느 구간에 속하는지 확인
						if infValue <= ciValues[0] {
							// 최소값보다 작거나 같으면 첫 번째 백분위수만 사용
							lowerIndex = 0
							upperIndex = 0
						} else if infValue >= ciValues[len(ciValues)-1] {
							// 최대값보다 크거나 같으면 마지막 백분위수만 사용
							lowerIndex = len(ciValues) - 1
							upperIndex = len(ciValues) - 1
						} else {
							// 중간 구간 찾기
							for i := 0; i < len(ciValues)-1; i++ {
								if infValue >= ciValues[i] && infValue <= ciValues[i+1] {
									lowerIndex = i
									upperIndex = i + 1
									break
								}
							}
						}

						if lowerIndex != -1 && upperIndex != -1 {
							// 두 백분위수가 같으면 하나만 표시
							if lowerIndex == upperIndex {
								workerData["confidential_percentiles"] = networkInference.ConfidenceIntervalRawPercentiles[lowerIndex]
							} else {
								// 두 백분위수 사이에 있으면 범위로 표시
								workerData["confidential_percentiles"] = networkInference.ConfidenceIntervalRawPercentiles[lowerIndex] + "~" +
									networkInference.ConfidenceIntervalRawPercentiles[upperIndex]
							}
						}
					}
				}
			}
		}

		synthesisValue = append(synthesisValue, workerData)
	}

	return synthesisValue
}

// getBlockTimestamp fetches the timestamp for a specific block height
func (s *TopicInferenceStore) getBlockTimestamp(blockHeight string) (string, error) {
	url := fmt.Sprintf("https://%s/cosmos/base/tendermint/v1beta1/blocks/%s", apiaddress, blockHeight)

	if s.debug {
		log.Printf("Block API 요청 URL: %s", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("블록 API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("블록 API 응답 오류: %d %s", resp.StatusCode, resp.Status)
	}

	// 응답 본문 디코딩
	var blockResponse struct {
		Block struct {
			Header struct {
				Time string `json:"time"`
			} `json:"header"`
		} `json:"block"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&blockResponse); err != nil {
		return "", fmt.Errorf("블록 JSON 디코딩 실패: %w", err)
	}

	if blockResponse.Block.Header.Time == "" {
		return "", fmt.Errorf("블록 타임스탬프 정보가 없습니다")
	}

	return blockResponse.Block.Header.Time, nil
}

// collectTopicData는 지정된 토픽의 추론 데이터를 수집합니다
func (s *TopicInferenceStore) collectTopicData(topicID string) error {
	if s.debug {
		log.Printf("토픽 %s 데이터 수집 시작", topicID)
	}

	// API에서 최신 네트워크 추론 데이터 가져오기
	url := fmt.Sprintf("https://%s/emissions/%s/latest_network_inferences/%s", apiaddress, version, topicID)

	if s.debug {
		log.Printf("API 요청 URL: %s", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 응답 오류: %d %s", resp.StatusCode, resp.Status)
	}

	// 응답 본문 디코딩
	var networkInference NetworkInference
	if err := json.NewDecoder(resp.Body).Decode(&networkInference); err != nil {
		return fmt.Errorf("JSON 디코딩 실패: %w", err)
	}

	// 기존 데이터와 비교하여 변경 여부 확인
	s.mu.Lock()
	defer s.mu.Unlock()

	existingData, exists := s.inferences[topicID]

	// 기존 데이터가 있고, inference_block_height가 동일하면 저장하지 않음
	// loss_block_height가 다르면 업데이트 진행
	if exists && existingData.InferenceBlockHeight == networkInference.InferenceBlockHeight {
		// loss_block_height가 다르면 업데이트 진행
		if existingData.LossBlockHeight != networkInference.LossBlockHeight {
			if s.debug {
				log.Printf("토픽 %s: loss_block_height 변경 감지 (기존=%s, 신규=%s), 업데이트 진행",
					topicID, existingData.LossBlockHeight, networkInference.LossBlockHeight)
			}
		} else {
			// 둘 다 동일하면 업데이트 하지 않음
			if s.debug {
				log.Printf("토픽 %s: 변경 없음 (inference_block_height=%s, loss_block_height=%s)",
					topicID, networkInference.InferenceBlockHeight, networkInference.LossBlockHeight)
			}
			return nil
		}
	}

	// 데이터 저장
	s.inferences[topicID] = &networkInference
	s.lastUpdated[topicID] = time.Now()

	if s.debug {
		log.Printf("토픽 %s: 새 데이터 저장 (inference_block_height=%s, loss_block_height=%s)",
			topicID, networkInference.InferenceBlockHeight, networkInference.LossBlockHeight)
	}

	// 데이터베이스에 저장
	if s.db != nil {
		// 데이터 가공 및 중복 필드 제거를 위한 처리
		// 먼저 필요한 데이터를 추출하여 synthesis_value 생성
		synthesisValue := s.createSynthesisData(networkInference, topicID)

		// 블록 타임스탬프 가져오기
		blockTimestamp := ""
		if networkInference.InferenceBlockHeight != "" {
			timestamp, err := s.getBlockTimestamp(networkInference.InferenceBlockHeight)
			if err != nil {
				if s.debug {
					log.Printf("블록 타임스탬프 조회 실패: %v, 현재 시간 사용", err)
				}
				blockTimestamp = time.Now().Format(time.RFC3339)
			} else {
				blockTimestamp = timestamp
			}
		} else {
			blockTimestamp = time.Now().Format(time.RFC3339)
		}

		// 저장할 데이터 구성
		storeData := map[string]interface{}{
			"topic_id":  topicID,
			"timestamp": blockTimestamp,
			"network_inferences": map[string]interface{}{
				"reputer_request_nonce":             networkInference.NetworkInferences.ReputerRequestNonce,
				"reputer":                           networkInference.NetworkInferences.Reputer,
				"extra_data":                        networkInference.NetworkInferences.ExtraData,
				"combined_value":                    networkInference.NetworkInferences.CombinedValue,
				"naive_value":                       networkInference.NetworkInferences.NaiveValue,
				"forecaster_values":                 networkInference.NetworkInferences.ForecasterValues,
				"one_out_forecaster_values":         networkInference.NetworkInferences.OneOutForecasterValues,
				"one_in_forecaster_values":          networkInference.NetworkInferences.OneInForecasterValues,
				"one_out_inferer_forecaster_values": networkInference.NetworkInferences.OneOutInfererForecasterValues,
				"synthesis_value":                   synthesisValue,
			},
			"inference_block_height":              networkInference.InferenceBlockHeight,
			"loss_block_height":                   networkInference.LossBlockHeight,
			"confidence_interval_raw_percentiles": networkInference.ConfidenceIntervalRawPercentiles,
			"confidence_interval_values":          networkInference.ConfidenceIntervalValues,
		}

		if err := s.db.SaveTopicInference(storeData); err != nil {
			log.Printf("토픽 %s 데이터 저장 실패: %v", topicID, err)
		}
	}

	return nil
}

// GetTopicInference는 지정된 토픽의 최신 추론 데이터를 반환합니다
func (s *TopicInferenceStore) GetTopicInference(topicID string) (*NetworkInference, time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inference, exists := s.inferences[topicID]
	if !exists {
		return nil, time.Time{}, false
	}

	lastUpdated := s.lastUpdated[topicID]
	return inference, lastUpdated, true
}

// GetAllTopicInferences는 모든 토픽의 최신 추론 데이터를 반환합니다
func (s *TopicInferenceStore) GetAllTopicInferences() map[string]*NetworkInference {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*NetworkInference, len(s.inferences))
	for topicID, inference := range s.inferences {
		result[topicID] = inference
	}

	return result
}

// ForceCollectTopicData는 지정된 토픽의 추론 데이터를 강제로 수집합니다
func (s *TopicInferenceStore) ForceCollectTopicData(topicID string) error {
	return s.collectTopicData(topicID)
}
