package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	_ "github.com/glebarez/go-sqlite" // 순수 Go로 작성된 SQLite 드라이버
	"github.com/golang/snappy"
)

// Database 구조체는 SQLite 데이터베이스 연결과 관련 메서드를 제공합니다
type Database struct {
	db    *sql.DB
	debug bool // 디버깅 모드 활성화 여부
}

// NewDatabase는 새로운 데이터베이스 연결을 생성합니다
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	// 데이터베이스 초기화
	if err := initDatabase(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("데이터베이스 초기화 실패: %w", err)
	}

	return &Database{db: db, debug: true}, nil
}

// SetDebug는 디버깅 모드를 설정합니다
func (d *Database) SetDebug(debug bool) {
	d.debug = debug
}

// Close는 데이터베이스 연결을 닫습니다
func (d *Database) Close() error {
	return d.db.Close()
}

// 데이터베이스 테이블 초기화
func initDatabase(db *sql.DB) error {
	// 경쟁 데이터 테이블 생성
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS competitions (
			id INTEGER PRIMARY KEY,
			timestamp TEXT NOT NULL,
			data BLOB NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// 인덱스 생성
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_competitions_timestamp ON competitions(timestamp)
	`)
	if err != nil {
		return err
	}

	// 새로운 경쟁 데이터 테이블 생성 (v2)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS competitions_v2 (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			preview_image_url TEXT,
			description TEXT,
			detailed_description TEXT,
			topic_id INTEGER,
			prize_pool INTEGER,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL,
			season_id INTEGER,
			tags TEXT,
			is_active BOOLEAN NOT NULL,
			timestamp TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// 새로운 경쟁 데이터 인덱스 생성
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_competitions_v2_topic_id ON competitions_v2(topic_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_competitions_v2_is_active ON competitions_v2(is_active)
	`)
	if err != nil {
		return err
	}

	// 리더보드 데이터 테이블 생성
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS leaderboard_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			topic_id TEXT NOT NULL,
			inference_block_height TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			cosmos_address TEXT NOT NULL,
			username TEXT,
			first_name TEXT,
			last_name TEXT,
			rank TEXT,
			points REAL,
			score REAL,
			loss REAL,
			is_active BOOLEAN
		)
	`)
	if err != nil {
		return err
	}

	// 리더보드 데이터 인덱스 생성
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_leaderboard_topic_id ON leaderboard_entries(topic_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_leaderboard_inference_block_height ON leaderboard_entries(inference_block_height)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_leaderboard_timestamp ON leaderboard_entries(timestamp)
	`)
	if err != nil {
		return err
	}

	// 토픽 추론 데이터 테이블 생성
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS topic_inferences (
			id INTEGER PRIMARY KEY,
			topic_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			inference_block_height TEXT NOT NULL,
			loss_block_height TEXT NOT NULL,
			data BLOB NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// 토픽 추론 데이터 인덱스 생성
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_topic_inferences_topic_id ON topic_inferences(topic_id)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_topic_inferences_timestamp ON topic_inferences(timestamp)
	`)

	return err
}

// SaveCompetitions는 경쟁 데이터를 압축하여 저장합니다
// 이미 존재하는 데이터는 업데이트하고, 새로운 데이터는 삽입합니다
func (d *Database) SaveCompetitions(data interface{}) error {
	if d.debug {
		log.Println("SaveCompetitions 시작")
	}

	// 데이터를 JSON으로 마샬링
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("JSON 마샬링 실패: %w", err)
	}

	if d.debug {
		log.Printf("JSON 마샬링 완료: %d 바이트", len(jsonData))

		// 디버깅 모드에서는 마샬링된 JSON을 파일로 저장
		debugFile := "debug_marshaled.json"
		if err := os.WriteFile(debugFile, jsonData, 0644); err != nil {
			log.Printf("디버그 파일 저장 실패: %v", err)
		} else {
			log.Printf("마샬링된 JSON이 %s 파일에 저장되었습니다", debugFile)
		}
	}

	// 데이터 압축
	compressedData := snappy.Encode(nil, jsonData)

	if d.debug {
		log.Printf("압축 완료: %d 바이트 -> %d 바이트 (압축률: %.2f%%)",
			len(jsonData), len(compressedData),
			100.0-(float64(len(compressedData))/float64(len(jsonData))*100.0))
	}

	// 현재 시간
	timestamp := time.Now().Format(time.RFC3339)

	// 가장 최근 레코드 확인
	var latestID int
	var exists bool
	err = d.db.QueryRow("SELECT id FROM competitions ORDER BY id DESC LIMIT 1").Scan(&latestID)
	if err == nil {
		exists = true
		if d.debug {
			log.Printf("기존 레코드 발견: ID=%d", latestID)
		}
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("기존 레코드 확인 실패: %w", err)
	}

	var result sql.Result
	if exists {
		// 기존 레코드 업데이트
		result, err = d.db.Exec(
			"UPDATE competitions SET timestamp = ?, data = ? WHERE id = ?",
			timestamp, compressedData, latestID,
		)
		if err != nil {
			return fmt.Errorf("데이터 업데이트 실패: %w", err)
		}

		if d.debug {
			rowsAffected, _ := result.RowsAffected()
			log.Printf("데이터 업데이트 완료: 영향받은 행 수=%d, ID=%d",
				rowsAffected, latestID)
		}
	} else {
		// 새 레코드 삽입
		result, err = d.db.Exec(
			"INSERT INTO competitions (timestamp, data) VALUES (?, ?)",
			timestamp, compressedData,
		)
		if err != nil {
			return fmt.Errorf("데이터 삽입 실패: %w", err)
		}

		if d.debug {
			rowsAffected, _ := result.RowsAffected()
			lastInsertID, _ := result.LastInsertId()
			log.Printf("새 데이터 삽입 완료: 영향받은 행 수=%d, 마지막 삽입 ID=%d",
				rowsAffected, lastInsertID)
		}
	}

	// 새로운 형식으로 competitions_v2 테이블에 저장
	if err := d.saveCompetitionsV2(data); err != nil {
		return fmt.Errorf("competitions_v2 테이블 저장 실패: %w", err)
	}

	return nil
}

// saveCompetitionsV2는 경쟁 데이터를 새로운 형식으로 저장합니다
func (d *Database) saveCompetitionsV2(data interface{}) error {
	if d.debug {
		log.Println("saveCompetitionsV2 시작")
	}

	// 데이터 타입 확인 및 변환
	competitionsResp, ok := data.(*CompetitionsResponse)
	if !ok {
		// 인터페이스 타입이면 JSON 마샬링/언마샬링을 통해 변환 시도
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("데이터 마샬링 실패: %w", err)
		}

		var resp CompetitionsResponse
		if err := json.Unmarshal(jsonData, &resp); err != nil {
			return fmt.Errorf("데이터 언마샬링 실패: %w", err)
		}

		competitionsResp = &resp
	}

	// 트랜잭션 시작
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 기존 데이터 삭제
	_, err = tx.Exec("DELETE FROM competitions_v2")
	if err != nil {
		return fmt.Errorf("기존 데이터 삭제 실패: %w", err)
	}

	// 현재 시간
	timestamp := time.Now().Format(time.RFC3339)

	// 활성 및 예정 경쟁 저장
	for _, comp := range competitionsResp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions {
		isActive := isCompetitionActive(comp, false)

		// 태그를 JSON 문자열로 변환
		tagsJSON, err := json.Marshal(comp.Tags)
		if err != nil {
			return fmt.Errorf("태그 JSON 변환 실패: %w", err)
		}

		// 설명 처리 (NULL 가능)
		var description *string
		if comp.Description != nil {
			description = comp.Description
		}

		_, err = tx.Exec(
			`INSERT INTO competitions_v2 (
				id, name, preview_image_url, description, detailed_description, 
				topic_id, prize_pool, start_date, end_date, season_id, 
				tags, is_active, timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			comp.ID, comp.Name, comp.PreviewImageURL, description, comp.DetailedDescription,
			comp.TopicID, comp.PrizePool, comp.StartDate.Format(time.RFC3339), comp.EndDate.Format(time.RFC3339), comp.SeasonID,
			string(tagsJSON), isActive, timestamp,
		)
		if err != nil {
			return fmt.Errorf("활성 경쟁 삽입 실패 (ID=%d): %w", comp.ID, err)
		}
	}

	// 과거 경쟁 저장
	for _, comp := range competitionsResp.PageProps.CompetitionsPage.PastCompetitions {
		isActive := isCompetitionActive(comp, true)

		// 태그를 JSON 문자열로 변환
		tagsJSON, err := json.Marshal(comp.Tags)
		if err != nil {
			return fmt.Errorf("태그 JSON 변환 실패: %w", err)
		}

		// 설명 처리 (NULL 가능)
		var description *string
		if comp.Description != nil {
			description = comp.Description
		}

		_, err = tx.Exec(
			`INSERT INTO competitions_v2 (
				id, name, preview_image_url, description, detailed_description, 
				topic_id, prize_pool, start_date, end_date, season_id, 
				tags, is_active, timestamp
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			comp.ID, comp.Name, comp.PreviewImageURL, description, comp.DetailedDescription,
			comp.TopicID, comp.PrizePool, comp.StartDate.Format(time.RFC3339), comp.EndDate.Format(time.RFC3339), comp.SeasonID,
			string(tagsJSON), isActive, timestamp,
		)
		if err != nil {
			return fmt.Errorf("과거 경쟁 삽입 실패 (ID=%d): %w", comp.ID, err)
		}
	}

	// 트랜잭션 커밋
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	if d.debug {
		log.Printf("competitions_v2 테이블 저장 완료: 활성 경쟁 %d개, 과거 경쟁 %d개",
			len(competitionsResp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions),
			len(competitionsResp.PageProps.CompetitionsPage.PastCompetitions))
	}

	return nil
}

// SaveTopicInference는 토픽 추론 데이터를 압축하여 저장합니다
func (d *Database) SaveTopicInference(data map[string]interface{}) error {
	if d.debug {
		log.Println("SaveTopicInference 시작")
	}

	// 필수 필드 확인
	topicID, ok := data["topic_id"].(string)
	if !ok {
		return fmt.Errorf("topic_id 필드가 없거나 문자열이 아닙니다")
	}

	timestamp, ok := data["timestamp"].(string)
	if !ok {
		timestamp = time.Now().Format(time.RFC3339)
	}

	inferenceBlockHeight, ok := data["inference_block_height"].(string)
	if !ok {
		return fmt.Errorf("inference_block_height 필드가 없거나 문자열이 아닙니다")
	}

	lossBlockHeight, ok := data["loss_block_height"].(string)
	if !ok {
		return fmt.Errorf("loss_block_height 필드가 없거나 문자열이 아닙니다")
	}

	// 기존 레코드 확인 - topic_id와 inference_block_height만으로 확인
	var existingID int
	var exists bool
	err := d.db.QueryRow(
		"SELECT id FROM topic_inferences WHERE topic_id = ? AND inference_block_height = ?",
		topicID, inferenceBlockHeight,
	).Scan(&existingID)

	if err == nil {
		exists = true
		if d.debug {
			log.Printf("기존 토픽 추론 레코드 발견: ID=%d", existingID)
		}
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("기존 레코드 확인 실패: %w", err)
	}

	// 데이터를 JSON으로 마샬링
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("JSON 마샬링 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 %s 데이터 JSON 마샬링 완료: %d 바이트", topicID, len(jsonData))
	}

	// 데이터 압축
	compressedData := snappy.Encode(nil, jsonData)

	if d.debug {
		log.Printf("토픽 %s 데이터 압축 완료: %d 바이트 -> %d 바이트 (압축률: %.2f%%)",
			topicID, len(jsonData), len(compressedData),
			100.0-(float64(len(compressedData))/float64(len(jsonData))*100.0))
	}

	var result sql.Result
	if exists {
		// 기존 레코드 업데이트 - loss_block_height도 함께 업데이트
		result, err = d.db.Exec(
			"UPDATE topic_inferences SET timestamp = ?, loss_block_height = ?, data = ? WHERE id = ?",
			timestamp, lossBlockHeight, compressedData, existingID,
		)
		if err != nil {
			return fmt.Errorf("토픽 데이터 업데이트 실패: %w", err)
		}

		if d.debug {
			rowsAffected, _ := result.RowsAffected()
			log.Printf("토픽 %s 데이터 업데이트 완료: 영향받은 행 수=%d, ID=%d",
				topicID, rowsAffected, existingID)
		}
	} else {
		// 새 레코드 삽입
		result, err = d.db.Exec(
			"INSERT INTO topic_inferences (topic_id, timestamp, inference_block_height, loss_block_height, data) VALUES (?, ?, ?, ?, ?)",
			topicID, timestamp, inferenceBlockHeight, lossBlockHeight, compressedData,
		)
		if err != nil {
			return fmt.Errorf("토픽 데이터 저장 실패: %w", err)
		}

		if d.debug {
			// 영향받은 행 수 확인
			rowsAffected, _ := result.RowsAffected()
			lastInsertID, _ := result.LastInsertId()
			log.Printf("토픽 %s 데이터 저장 완료: 영향받은 행 수=%d, 마지막 삽입 ID=%d",
				topicID, rowsAffected, lastInsertID)
		}
	}

	return nil
}

// GetLatestTopicInference는 지정된 토픽의 가장 최근 추론 데이터를 가져옵니다
func (d *Database) GetLatestTopicInference(topicID string) (map[string]interface{}, error) {
	if d.debug {
		log.Printf("GetLatestTopicInference 시작: 토픽 ID=%s", topicID)
	}

	var timestamp string
	var compressedData []byte
	var inferenceBlockHeight string

	// 가장 최근 데이터 조회
	err := d.db.QueryRow(
		"SELECT timestamp, data, inference_block_height FROM topic_inferences WHERE topic_id = ? ORDER BY timestamp DESC LIMIT 1",
		topicID,
	).Scan(&timestamp, &compressedData, &inferenceBlockHeight)

	if err != nil {
		if err == sql.ErrNoRows {
			if d.debug {
				log.Printf("토픽 %s의 데이터가 없습니다", topicID)
			}
			return nil, nil // 데이터가 없는 경우
		}
		return nil, fmt.Errorf("토픽 데이터 조회 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 %s 최근 데이터 조회 완료: 타임스탬프=%s, 압축 데이터 크기=%d 바이트",
			topicID, timestamp, len(compressedData))
	}

	// 이전 블록 높이 조회
	var prevHeight sql.NullString
	err = d.db.QueryRow(
		"SELECT inference_block_height FROM topic_inferences WHERE topic_id = ? AND timestamp < ? ORDER BY timestamp DESC LIMIT 1",
		topicID, timestamp,
	).Scan(&prevHeight)

	// 데이터 압축 해제
	jsonData, err := snappy.Decode(nil, compressedData)
	if err != nil {
		return nil, fmt.Errorf("압축 해제 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 %s 데이터 압축 해제 완료: %d 바이트 -> %d 바이트",
			topicID, len(compressedData), len(jsonData))
	}

	// JSON 언마샬링
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("JSON 언마샬링 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 %s 데이터 JSON 언마샬링 완료", topicID)
	}

	// 데이터 재가공
	result = d.processTopicInferenceData(result)

	// 중복된 topic_id 제거
	// 최상위 topic_id는 유지 (API 응답에서 필요할 수 있음)
	if networkInferences, ok := result["network_inferences"].(map[string]interface{}); ok {
		delete(networkInferences, "topic_id")
		result["network_inferences"] = networkInferences
	}

	// 이전/다음 블록 높이 정보 추가
	if prevHeight.Valid {
		result["prev_height"] = prevHeight.String
	} else {
		result["prev_height"] = nil
	}

	result["next_height"] = nil // 최신 데이터이므로 다음 높이는 없음

	return result, nil
}

// processTopicInferenceData는 토픽 추론 데이터를 재가공합니다
func (d *Database) processTopicInferenceData(data map[string]interface{}) map[string]interface{} {
	// 네트워크 추론 데이터가 없으면 원본 데이터 반환
	networkInferences, ok := data["network_inferences"].(map[string]interface{})
	if !ok {
		return data
	}

	// 이미 synthesis_value가 있는 경우 그대로 사용
	if _, hasSynthesisValue := networkInferences["synthesis_value"].([]interface{}); hasSynthesisValue {
		// synthesis_data 필드가 있다면 제거
		delete(networkInferences, "synthesis_data")
		data["network_inferences"] = networkInferences
		return data
	}

	// 이전 버전 호환성을 위해 synthesis_data가 있으면 synthesis_value로 이동
	if synthesisData, hasSynthesisData := networkInferences["synthesis_data"].([]interface{}); hasSynthesisData {
		networkInferences["synthesis_value"] = synthesisData
		delete(networkInferences, "synthesis_data")
		data["network_inferences"] = networkInferences
		return data
	}

	// 필요한 데이터 추출
	confidenceIntervalRawPercentiles, _ := data["confidence_interval_raw_percentiles"].([]interface{})
	confidenceIntervalValues, _ := data["confidence_interval_values"].([]interface{})

	// 워커 정보 및 값 추출
	infererValues, _ := networkInferences["inferer_values"].([]interface{})
	oneOutInfererValues, _ := networkInferences["one_out_inferer_values"].([]interface{})
	infererWeights, _ := data["inferer_weights"].([]interface{})

	// 워커 맵 생성 (worker 주소를 키로 사용)
	workerMap := make(map[string]map[string]interface{})

	// inferer_values 처리
	for _, iv := range infererValues {
		if ivMap, ok := iv.(map[string]interface{}); ok {
			worker, _ := ivMap["worker"].(string)
			value, _ := ivMap["value"].(string)

			if worker != "" {
				if _, exists := workerMap[worker]; !exists {
					workerMap[worker] = make(map[string]interface{})
				}
				workerMap[worker]["worker"] = worker
				workerMap[worker]["inferer_values"] = value
			}
		}
	}

	// one_out_inferer_values 처리
	for _, oiv := range oneOutInfererValues {
		if oivMap, ok := oiv.(map[string]interface{}); ok {
			worker, _ := oivMap["worker"].(string)
			value, _ := oivMap["value"].(string)

			if worker != "" {
				if _, exists := workerMap[worker]; !exists {
					workerMap[worker] = make(map[string]interface{})
					workerMap[worker]["worker"] = worker
				}
				workerMap[worker]["one_out_inferer_values"] = value
			}
		}
	}

	// inferer_weights 처리
	for _, iw := range infererWeights {
		if iwMap, ok := iw.(map[string]interface{}); ok {
			worker, _ := iwMap["worker"].(string)
			weight, _ := iwMap["weight"].(string)

			if worker != "" {
				if _, exists := workerMap[worker]; !exists {
					workerMap[worker] = make(map[string]interface{})
					workerMap[worker]["worker"] = worker
				}
				workerMap[worker]["weight"] = weight
			}
		}
	}

	// 신뢰 구간 백분위수 계산 및 합성 데이터 생성
	var synthesisValue []map[string]interface{}
	for _, workerData := range workerMap {
		// 신뢰 구간 백분위수 계산
		infererValue, _ := workerData["inferer_values"].(string)
		if infererValue != "" {
			confidenceInterval := calculateConfidenceInterval(
				infererValue,
				confidenceIntervalValues,
				confidenceIntervalRawPercentiles,
			)
			workerData["confidence_interval_raw_percentiles"] = confidenceInterval

			// 신뢰 구간 백분위수를 범위 형식으로 추가 (예: "84.13~97.72")
			if len(confidenceIntervalRawPercentiles) >= 2 && len(confidenceIntervalValues) >= 2 {
				// infererValue를 float64로 변환
				infValue, err := strconv.ParseFloat(infererValue, 64)
				if err == nil {
					// 각 신뢰 구간 값을 float64로 변환
					var ciValues []float64
					for _, v := range confidenceIntervalValues {
						if strVal, ok := v.(string); ok {
							val, err := strconv.ParseFloat(strVal, 64)
							if err == nil {
								ciValues = append(ciValues, val)
							}
						}
					}

					// 신뢰 구간 백분위수 문자열 배열
					var ciPercentiles []string
					for _, p := range confidenceIntervalRawPercentiles {
						if strVal, ok := p.(string); ok {
							ciPercentiles = append(ciPercentiles, strVal)
						}
					}

					// 값이 없으면 기본값 반환
					if len(ciValues) > 0 && len(ciPercentiles) > 0 && len(ciValues) == len(ciPercentiles) {
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
								workerData["confidential_percentiles"] = ciPercentiles[lowerIndex]
							} else {
								// 두 백분위수 사이에 있으면 범위로 표시
								workerData["confidential_percentiles"] = ciPercentiles[lowerIndex] + "~" + ciPercentiles[upperIndex]
							}
						}
					}
				}
			}
		}

		// 합성 데이터에 추가
		synthesisValue = append(synthesisValue, workerData)
	}

	// 네트워크 추론 데이터에 synthesis_value만 추가
	networkInferences["synthesis_value"] = synthesisValue
	data["network_inferences"] = networkInferences

	return data
}

// calculateConfidenceInterval은 신뢰 구간 백분위수를 계산합니다
func calculateConfidenceInterval(infererValue string, confidenceIntervalValues []interface{}, confidenceIntervalRawPercentiles []interface{}) string {
	// inferer_values와 confidence_interval_values를 비교하여 적절한 백분위수 결정
	if len(confidenceIntervalValues) == 0 || len(confidenceIntervalRawPercentiles) == 0 {
		return "50" // 기본값
	}

	// infererValue를 float64로 변환
	infValue, err := strconv.ParseFloat(infererValue, 64)
	if err != nil {
		return "50" // 변환 실패 시 기본값
	}

	// 각 신뢰 구간 값을 float64로 변환
	var ciValues []float64
	for _, v := range confidenceIntervalValues {
		if strVal, ok := v.(string); ok {
			val, err := strconv.ParseFloat(strVal, 64)
			if err == nil {
				ciValues = append(ciValues, val)
			}
		}
	}

	// 신뢰 구간 백분위수 문자열 배열
	var ciPercentiles []string
	for _, p := range confidenceIntervalRawPercentiles {
		if strVal, ok := p.(string); ok {
			ciPercentiles = append(ciPercentiles, strVal)
		}
	}

	// 값이 없으면 기본값 반환
	if len(ciValues) == 0 || len(ciPercentiles) == 0 || len(ciValues) != len(ciPercentiles) {
		return "50"
	}

	// infererValue가 어느 구간에 속하는지 확인
	if infValue <= ciValues[0] {
		return ciPercentiles[0] // 최소값보다 작거나 같으면 최소 백분위수
	}

	if infValue >= ciValues[len(ciValues)-1] {
		return ciPercentiles[len(ciPercentiles)-1] // 최대값보다 크거나 같으면 최대 백분위수
	}

	// 중간 구간 찾기
	for i := 0; i < len(ciValues)-1; i++ {
		if infValue >= ciValues[i] && infValue <= ciValues[i+1] {
			// 더 가까운 값의 백분위수 반환
			if math.Abs(infValue-ciValues[i]) < math.Abs(infValue-ciValues[i+1]) {
				return ciPercentiles[i]
			} else {
				return ciPercentiles[i+1]
			}
		}
	}

	// 기본값
	return "50"
}

// GetTopicInferencesByTimeRange는 지정된 시간 범위의 토픽 추론 데이터를 가져옵니다
func (d *Database) GetTopicInferencesByTimeRange(topicID string, start, end time.Time) ([]map[string]interface{}, error) {
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	if d.debug {
		log.Printf("GetTopicInferencesByTimeRange 시작: 토픽 ID=%s, %s ~ %s", topicID, startStr, endStr)
	}

	rows, err := d.db.Query(
		"SELECT timestamp, data FROM topic_inferences WHERE topic_id = ? AND timestamp BETWEEN ? AND ? ORDER BY timestamp",
		topicID, startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("토픽 데이터 조회 실패: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var timestamp string
		var compressedData []byte

		if err := rows.Scan(&timestamp, &compressedData); err != nil {
			return nil, fmt.Errorf("데이터 스캔 실패: %w", err)
		}

		// 데이터 압축 해제
		jsonData, err := snappy.Decode(nil, compressedData)
		if err != nil {
			return nil, fmt.Errorf("압축 해제 실패: %w", err)
		}

		// JSON 언마샬링
		var result map[string]interface{}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return nil, fmt.Errorf("JSON 언마샬링 실패: %w", err)
		}

		// 데이터 재가공
		result = d.processTopicInferenceData(result)

		// 중복된 topic_id 제거
		if networkInferences, ok := result["network_inferences"].(map[string]interface{}); ok {
			delete(networkInferences, "topic_id")
			result["network_inferences"] = networkInferences
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 처리 중 오류: %w", err)
	}

	if d.debug {
		log.Printf("GetTopicInferencesByTimeRange 완료: 토픽 ID=%s, %d개 결과 반환", topicID, len(results))
	}

	return results, nil
}

// GetLatestCompetitions는 가장 최근의 경쟁 데이터를 가져옵니다
func (d *Database) GetLatestCompetitions() (interface{}, error) {
	if d.debug {
		log.Println("GetLatestCompetitions 시작")
	}

	var timestamp string
	var compressedData []byte

	// 가장 최근 데이터 조회
	err := d.db.QueryRow(
		"SELECT timestamp, data FROM competitions ORDER BY timestamp DESC LIMIT 1",
	).Scan(&timestamp, &compressedData)

	if err != nil {
		if err == sql.ErrNoRows {
			if d.debug {
				log.Println("데이터가 없습니다")
			}
			return nil, nil // 데이터가 없는 경우
		}
		return nil, fmt.Errorf("데이터 조회 실패: %w", err)
	}

	if d.debug {
		log.Printf("최근 데이터 조회 완료: 타임스탬프=%s, 압축 데이터 크기=%d 바이트",
			timestamp, len(compressedData))
	}

	// 데이터 압축 해제
	jsonData, err := snappy.Decode(nil, compressedData)
	if err != nil {
		return nil, fmt.Errorf("압축 해제 실패: %w", err)
	}

	if d.debug {
		log.Printf("압축 해제 완료: %d 바이트 -> %d 바이트",
			len(compressedData), len(jsonData))
	}

	// JSON 언마샬링
	var result interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("JSON 언마샬링 실패: %w", err)
	}

	if d.debug {
		log.Println("JSON 언마샬링 완료")
	}

	return result, nil
}

// GetCompetitionsByTimeRange는 지정된 시간 범위의 경쟁 데이터를 가져옵니다
func (d *Database) GetCompetitionsByTimeRange(start, end time.Time) ([]interface{}, error) {
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)

	if d.debug {
		log.Printf("GetCompetitionsByTimeRange 시작: %s ~ %s", startStr, endStr)
	}

	rows, err := d.db.Query(
		"SELECT timestamp, data FROM competitions WHERE timestamp BETWEEN ? AND ? ORDER BY timestamp",
		startStr, endStr,
	)
	if err != nil {
		return nil, fmt.Errorf("데이터 조회 실패: %w", err)
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var timestamp string
		var compressedData []byte

		if err := rows.Scan(&timestamp, &compressedData); err != nil {
			return nil, fmt.Errorf("데이터 스캔 실패: %w", err)
		}

		// 데이터 압축 해제
		jsonData, err := snappy.Decode(nil, compressedData)
		if err != nil {
			return nil, fmt.Errorf("압축 해제 실패: %w", err)
		}

		// JSON 언마샬링
		var result interface{}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return nil, fmt.Errorf("JSON 언마샬링 실패: %w", err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 처리 중 오류: %w", err)
	}

	if d.debug {
		log.Printf("GetCompetitionsByTimeRange 완료: %d개 결과 반환", len(results))
	}

	return results, nil
}

// PruneOldData는 지정된 기간보다 오래된 데이터를 삭제합니다
func (d *Database) PruneOldData(retentionPeriod time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-retentionPeriod).Format(time.RFC3339)

	if d.debug {
		log.Printf("PruneOldData 시작: 기준 시간=%s", cutoffTime)
	}

	result, err := d.db.Exec(
		"DELETE FROM competitions WHERE timestamp < ?",
		cutoffTime,
	)
	if err != nil {
		return 0, fmt.Errorf("데이터 삭제 실패: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("영향받은 행 수 확인 실패: %w", err)
	}

	if d.debug {
		log.Printf("PruneOldData 완료: %d개 레코드 삭제됨", rowsAffected)
	}

	return rowsAffected, nil
}

// PruneOldTopicData는 지정된 기간보다 오래된 토픽 데이터를 삭제합니다
func (d *Database) PruneOldTopicData(topicID string, retentionPeriod time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-retentionPeriod).Format(time.RFC3339)

	if d.debug {
		log.Printf("PruneOldTopicData 시작: 토픽 ID=%s, 기준 시간=%s", topicID, cutoffTime)
	}

	var result sql.Result
	var err error

	if topicID == "" {
		// 모든 토픽 데이터 정리
		result, err = d.db.Exec(
			"DELETE FROM topic_inferences WHERE timestamp < ?",
			cutoffTime,
		)
	} else {
		// 특정 토픽 데이터만 정리
		result, err = d.db.Exec(
			"DELETE FROM topic_inferences WHERE topic_id = ? AND timestamp < ?",
			topicID, cutoffTime,
		)
	}

	if err != nil {
		return 0, fmt.Errorf("토픽 데이터 삭제 실패: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("영향받은 행 수 확인 실패: %w", err)
	}

	if d.debug {
		if topicID == "" {
			log.Printf("모든 토픽 데이터 정리 완료: %d개 레코드 삭제됨", rowsAffected)
		} else {
			log.Printf("토픽 %s 데이터 정리 완료: %d개 레코드 삭제됨", topicID, rowsAffected)
		}
	}

	return rowsAffected, nil
}

// GetDatabaseStats는 데이터베이스 통계를 반환합니다
func (d *Database) GetDatabaseStats() (map[string]interface{}, error) {
	if d.debug {
		log.Println("GetDatabaseStats 시작")
	}

	var count int
	var oldestTimestamp, newestTimestamp string
	var totalSizeBytes int64

	// 총 레코드 수
	err := d.db.QueryRow("SELECT COUNT(*) FROM competitions").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("레코드 수 조회 실패: %w", err)
	}

	// 가장 오래된/최신 타임스탬프
	err = d.db.QueryRow("SELECT MIN(timestamp) FROM competitions").Scan(&oldestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("최소 타임스탬프 조회 실패: %w", err)
	}

	err = d.db.QueryRow("SELECT MAX(timestamp) FROM competitions").Scan(&newestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("최대 타임스탬프 조회 실패: %w", err)
	}

	// 총 데이터 크기
	err = d.db.QueryRow("SELECT SUM(LENGTH(data)) FROM competitions").Scan(&totalSizeBytes)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("데이터 크기 조회 실패: %w", err)
	}

	// 토픽 데이터 통계
	var topicCount int
	var topicOldestTimestamp, topicNewestTimestamp string
	var topicTotalSizeBytes int64

	// 총 토픽 레코드 수
	err = d.db.QueryRow("SELECT COUNT(*) FROM topic_inferences").Scan(&topicCount)
	if err != nil {
		return nil, fmt.Errorf("토픽 레코드 수 조회 실패: %w", err)
	}

	// 가장 오래된/최신 토픽 타임스탬프
	err = d.db.QueryRow("SELECT MIN(timestamp) FROM topic_inferences").Scan(&topicOldestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("토픽 최소 타임스탬프 조회 실패: %w", err)
	}

	err = d.db.QueryRow("SELECT MAX(timestamp) FROM topic_inferences").Scan(&topicNewestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("토픽 최대 타임스탬프 조회 실패: %w", err)
	}

	// 총 토픽 데이터 크기
	err = d.db.QueryRow("SELECT SUM(LENGTH(data)) FROM topic_inferences").Scan(&topicTotalSizeBytes)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("토픽 데이터 크기 조회 실패: %w", err)
	}

	// 토픽 종류 수
	var uniqueTopicCount int
	err = d.db.QueryRow("SELECT COUNT(DISTINCT topic_id) FROM topic_inferences").Scan(&uniqueTopicCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("고유 토픽 수 조회 실패: %w", err)
	}

	stats := map[string]interface{}{
		"competitions": map[string]interface{}{
			"record_count":     count,
			"oldest_timestamp": oldestTimestamp,
			"newest_timestamp": newestTimestamp,
			"total_size_bytes": totalSizeBytes,
			"total_size_mb":    float64(totalSizeBytes) / (1024 * 1024),
		},
		"topic_inferences": map[string]interface{}{
			"record_count":       topicCount,
			"unique_topic_count": uniqueTopicCount,
			"oldest_timestamp":   topicOldestTimestamp,
			"newest_timestamp":   topicNewestTimestamp,
			"total_size_bytes":   topicTotalSizeBytes,
			"total_size_mb":      float64(topicTotalSizeBytes) / (1024 * 1024),
		},
	}

	if d.debug {
		log.Printf("GetDatabaseStats 완료: 경쟁 레코드 수=%d, 크기=%.2fMB, 토픽 레코드 수=%d, 크기=%.2fMB",
			count, float64(totalSizeBytes)/(1024*1024),
			topicCount, float64(topicTotalSizeBytes)/(1024*1024))
	}

	return stats, nil
}

// GetTopicStats는 특정 토픽의 통계를 반환합니다
func (d *Database) GetTopicStats(topicID string) (map[string]interface{}, error) {
	if d.debug {
		log.Printf("GetTopicStats 시작: 토픽 ID=%s", topicID)
	}

	var count int
	var oldestTimestamp, newestTimestamp string
	var totalSizeBytes int64

	// 총 레코드 수
	err := d.db.QueryRow("SELECT COUNT(*) FROM topic_inferences WHERE topic_id = ?", topicID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("토픽 레코드 수 조회 실패: %w", err)
	}

	// 가장 오래된/최신 타임스탬프
	err = d.db.QueryRow("SELECT MIN(timestamp) FROM topic_inferences WHERE topic_id = ?", topicID).Scan(&oldestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("토픽 최소 타임스탬프 조회 실패: %w", err)
	}

	err = d.db.QueryRow("SELECT MAX(timestamp) FROM topic_inferences WHERE topic_id = ?", topicID).Scan(&newestTimestamp)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("토픽 최대 타임스탬프 조회 실패: %w", err)
	}

	// 총 데이터 크기
	err = d.db.QueryRow("SELECT SUM(LENGTH(data)) FROM topic_inferences WHERE topic_id = ?", topicID).Scan(&totalSizeBytes)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("토픽 데이터 크기 조회 실패: %w", err)
	}

	stats := map[string]interface{}{
		"topic_id":         topicID,
		"record_count":     count,
		"oldest_timestamp": oldestTimestamp,
		"newest_timestamp": newestTimestamp,
		"total_size_bytes": totalSizeBytes,
		"total_size_mb":    float64(totalSizeBytes) / (1024 * 1024),
	}

	if d.debug {
		log.Printf("GetTopicStats 완료: 토픽 ID=%s, 레코드 수=%d, 크기=%.2fMB",
			topicID, count, float64(totalSizeBytes)/(1024*1024))
	}

	return stats, nil
}

// GetTopicBlockHeights는 특정 토픽의 블록 높이 리스트를 반환합니다
func (d *Database) GetTopicBlockHeights(topicID string, limit int, offset int) (map[string]interface{}, error) {
	if d.debug {
		log.Printf("GetTopicBlockHeights 시작: 토픽 ID=%s, 제한=%d, 오프셋=%d", topicID, limit, offset)
	}

	// 기본값 설정
	if limit <= 0 {
		limit = 100 // 기본 제한 값
	}
	if offset < 0 {
		offset = 0
	}

	// 총 레코드 수 조회
	var totalCount int
	err := d.db.QueryRow("SELECT COUNT(*) FROM topic_inferences WHERE topic_id = ?", topicID).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("토픽 레코드 수 조회 실패: %w", err)
	}

	// 블록 높이 리스트 조회 (내림차순)
	rows, err := d.db.Query(
		"SELECT inference_block_height FROM topic_inferences WHERE topic_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?",
		topicID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("블록 높이 조회 실패: %w", err)
	}
	defer rows.Close()

	var heights []string
	for rows.Next() {
		var inferenceBlockHeight string
		if err := rows.Scan(&inferenceBlockHeight); err != nil {
			return nil, fmt.Errorf("데이터 스캔 실패: %w", err)
		}

		heights = append(heights, inferenceBlockHeight)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 처리 중 오류: %w", err)
	}

	result := map[string]interface{}{
		"topic_id":      topicID,
		"total_count":   totalCount,
		"limit":         limit,
		"offset":        offset,
		"heights":       heights,
		"heights_count": len(heights),
	}

	if d.debug {
		log.Printf("GetTopicBlockHeights 완료: 토픽 ID=%s, 조회된 높이 수=%d", topicID, len(heights))
	}

	return result, nil
}

// GetTopicInferenceByHeight는 특정 블록 높이에 대한 토픽 추론 데이터를 가져옵니다
func (d *Database) GetTopicInferenceByHeight(topicID string, height string) (map[string]interface{}, error) {
	if d.debug {
		log.Printf("GetTopicInferenceByHeight 시작: 토픽 ID=%s, 블록 높이=%s", topicID, height)
	}

	// 특정 블록 높이에 대한 데이터 조회
	var timestamp string
	var compressedData []byte
	err := d.db.QueryRow(
		"SELECT timestamp, data FROM topic_inferences WHERE topic_id = ? AND inference_block_height = ? ORDER BY timestamp DESC LIMIT 1",
		topicID, height,
	).Scan(&timestamp, &compressedData)

	if err != nil {
		if err == sql.ErrNoRows {
			if d.debug {
				log.Printf("토픽 %s의 블록 높이 %s에 대한 데이터가 없습니다", topicID, height)
			}
			return nil, nil // 데이터가 없는 경우
		}
		return nil, fmt.Errorf("데이터 조회 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 %s 블록 높이 %s 데이터 조회 완료: 타임스탬프=%s, 압축 데이터 크기=%d 바이트",
			topicID, height, timestamp, len(compressedData))
	}

	// 이전 블록 높이 조회
	var prevHeight sql.NullString
	err = d.db.QueryRow(
		"SELECT inference_block_height FROM topic_inferences WHERE topic_id = ? AND timestamp < ? ORDER BY timestamp DESC LIMIT 1",
		topicID, timestamp,
	).Scan(&prevHeight)

	// 다음 블록 높이 조회
	var nextHeight sql.NullString
	err = d.db.QueryRow(
		"SELECT inference_block_height FROM topic_inferences WHERE topic_id = ? AND timestamp > ? ORDER BY timestamp ASC LIMIT 1",
		topicID, timestamp,
	).Scan(&nextHeight)

	// 데이터 압축 해제
	jsonData, err := snappy.Decode(nil, compressedData)
	if err != nil {
		return nil, fmt.Errorf("압축 해제 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 %s 데이터 압축 해제 완료: %d 바이트 -> %d 바이트",
			topicID, len(compressedData), len(jsonData))
	}

	// JSON 언마샬링
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("JSON 언마샬링 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 %s 데이터 JSON 언마샬링 완료", topicID)
	}

	// 데이터 재가공
	result = d.processTopicInferenceData(result)

	// 중복된 topic_id 제거
	if networkInferences, ok := result["network_inferences"].(map[string]interface{}); ok {
		delete(networkInferences, "topic_id")
		result["network_inferences"] = networkInferences
	}

	// 이전/다음 블록 높이 정보 추가
	if prevHeight.Valid {
		result["prev_height"] = prevHeight.String
	} else {
		result["prev_height"] = nil
	}

	if nextHeight.Valid {
		result["next_height"] = nextHeight.String
	} else {
		result["next_height"] = nil
	}

	// 저장된 리더보드 데이터 가져오기
	leaderboardEntries, err := d.GetLeaderboardEntries(topicID, height)
	if err != nil {
		log.Printf("리더보드 데이터 조회 실패: %v", err)
		// 오류가 발생해도 계속 진행
	} else if len(leaderboardEntries) > 0 {
		// 리더보드 데이터가 있으면 추가
		if d.debug {
			log.Printf("토픽 %s 블록 높이 %s에 대한 리더보드 데이터 %d개 추가", topicID, height, len(leaderboardEntries))
		}

		// network_inferences에 synthesis_value로 리더보드 데이터 추가
		if networkInferences, ok := result["network_inferences"].(map[string]interface{}); ok {
			// network_inferences가 있는 경우, synthesis_value를 리더보드 데이터로 대체
			networkInferences["synthesis_value"] = leaderboardEntries
			result["network_inferences"] = networkInferences
		} else {
			// network_inferences가 없는 경우
			result["network_inferences"] = map[string]interface{}{
				"synthesis_value": leaderboardEntries,
			}
		}
	}

	return result, nil
}

// GetCompetitionIDFromTopicID는 토픽 ID에 해당하는 경쟁 ID를 반환합니다
func (d *Database) GetCompetitionIDFromTopicID(topicID string) (string, error) {
	if d.debug {
		log.Printf("GetCompetitionIDFromTopicID 시작: 토픽 ID=%s", topicID)
	}

	// 먼저 competitions_v2 테이블에서 조회 시도
	query := `
		SELECT id 
		FROM competitions_v2 
		WHERE topic_id = ? 
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var competitionID int
	err := d.db.QueryRow(query, topicID).Scan(&competitionID)
	if err == nil {
		// competitions_v2 테이블에서 찾은 경우
		if d.debug {
			log.Printf("competitions_v2 테이블에서 토픽 ID %s의 경쟁 ID: %d", topicID, competitionID)
		}
		return strconv.Itoa(competitionID), nil
	} else if err != sql.ErrNoRows {
		// 쿼리 실행 중 오류가 발생한 경우
		return "", fmt.Errorf("competitions_v2 테이블에서 경쟁 ID 조회 실패: %w", err)
	}

	// competitions_v2 테이블에서 찾지 못한 경우, 기존 테이블에서 조회
	if d.debug {
		log.Printf("competitions_v2 테이블에서 토픽 ID %s에 해당하는 경쟁 데이터 없음, 기존 테이블에서 조회", topicID)
	}

	// 기존 경쟁 데이터에서 토픽 ID에 해당하는 경쟁 ID 조회
	// 이 부분은 기존 데이터 구조에 따라 다를 수 있음
	// 여기서는 기존 데이터가 JSON 형태로 저장되어 있다고 가정
	latestData, err := d.GetLatestCompetitions()
	if err != nil {
		return "", fmt.Errorf("기존 경쟁 데이터 조회 실패: %w", err)
	}

	if latestData == nil {
		return "", fmt.Errorf("토픽 ID %s에 해당하는 경쟁 데이터가 없습니다", topicID)
	}

	// 데이터 타입 확인 및 변환
	competitionsResp, ok := latestData.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("데이터 타입이 예상과 다릅니다")
	}

	// 데이터 구조 탐색
	pageProps, ok := competitionsResp["pageProps"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("pageProps 필드를 찾을 수 없습니다")
	}

	competitionsPage, ok := pageProps["competitionsPage"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("competitionsPage 필드를 찾을 수 없습니다")
	}

	// 활성 경쟁 확인
	activeCompetitions, ok := competitionsPage["activeAndUpcomingCompetitions"].([]interface{})
	if ok {
		for _, comp := range activeCompetitions {
			competition, ok := comp.(map[string]interface{})
			if !ok {
				continue
			}

			compTopicID, ok := competition["topic_id"].(float64)
			if !ok {
				continue
			}

			if strconv.FormatFloat(compTopicID, 'f', 0, 64) == topicID {
				compID, ok := competition["id"].(float64)
				if ok {
					if d.debug {
						log.Printf("기존 테이블의 활성 경쟁에서 토픽 ID %s의 경쟁 ID: %.0f", topicID, compID)
					}
					return strconv.FormatFloat(compID, 'f', 0, 64), nil
				}
			}
		}
	}

	// 과거 경쟁 확인
	pastCompetitions, ok := competitionsPage["pastCompetitions"].([]interface{})
	if ok {
		for _, comp := range pastCompetitions {
			competition, ok := comp.(map[string]interface{})
			if !ok {
				continue
			}

			compTopicID, ok := competition["topic_id"].(float64)
			if !ok {
				continue
			}

			if strconv.FormatFloat(compTopicID, 'f', 0, 64) == topicID {
				compID, ok := competition["id"].(float64)
				if ok {
					if d.debug {
						log.Printf("기존 테이블의 과거 경쟁에서 토픽 ID %s의 경쟁 ID: %.0f", topicID, compID)
					}
					return strconv.FormatFloat(compID, 'f', 0, 64), nil
				}
			}
		}
	}

	if d.debug {
		log.Printf("토픽 ID %s에 해당하는 경쟁 데이터 없음", topicID)
	}
	return "", fmt.Errorf("토픽 ID %s에 해당하는 경쟁 데이터가 없습니다", topicID)
}

// isCompetitionActive는 경쟁이 활성 상태인지 확인합니다
// 활성 상태 조건:
// 1. topic_id가 0이 아님
// 2. pastCompetitions에 포함되지 않음
// 3. tags가 비어있음
func isCompetitionActive(competition Competition, isPastCompetition bool) bool {
	// 조건 1: topic_id가 0이 아님
	if competition.TopicID == 0 {
		return false
	}

	// 조건 2: pastCompetitions에 포함되지 않음
	if isPastCompetition {
		return false
	}

	// 조건 3: tags가 비어있음
	return len(competition.Tags) == 0
}

// CompetitionV2는 새로운 형식의 경쟁 정보 구조체입니다
type CompetitionV2 struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	PreviewImageURL string  `json:"preview_image_url"`
	Description     *string `json:"description"`
	// DetailedDescription string    `json:"detailed_description"`
	TopicID   int       `json:"topic_id"`
	PrizePool int       `json:"prize_pool"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	SeasonID  int       `json:"season_id"`
	Tags      []string  `json:"tags"`
	IsActive  bool      `json:"is_active"`
}

// GetCompetitionsV2는 새로운 형식의 경쟁 데이터를 가져옵니다
func (d *Database) GetCompetitionsV2() ([]CompetitionV2, error) {
	if d.debug {
		log.Println("GetCompetitionsV2 시작")
	}

	rows, err := d.db.Query(
		`SELECT 
			id, name, preview_image_url, description, 
			topic_id, prize_pool, start_date, end_date, season_id, 
			tags, is_active 
		FROM competitions_v2 
		ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("데이터 조회 실패: %w", err)
	}
	defer rows.Close()

	var competitions []CompetitionV2
	for rows.Next() {
		var comp CompetitionV2
		var startDateStr, endDateStr string
		var tagsJSON string

		err := rows.Scan(
			&comp.ID, &comp.Name, &comp.PreviewImageURL, &comp.Description,
			&comp.TopicID, &comp.PrizePool, &startDateStr, &endDateStr, &comp.SeasonID,
			&tagsJSON, &comp.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("데이터 스캔 실패: %w", err)
		}

		// 날짜 문자열을 시간으로 변환
		comp.StartDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, fmt.Errorf("시작 날짜 파싱 실패: %w", err)
		}

		comp.EndDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, fmt.Errorf("종료 날짜 파싱 실패: %w", err)
		}

		// 태그 JSON 파싱
		if err := json.Unmarshal([]byte(tagsJSON), &comp.Tags); err != nil {
			return nil, fmt.Errorf("태그 JSON 파싱 실패: %w", err)
		}

		competitions = append(competitions, comp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 처리 중 오류: %w", err)
	}

	if d.debug {
		log.Printf("경쟁 데이터 조회 완료: %d개", len(competitions))
	}

	return competitions, nil
}

// GetActiveCompetitionsV2는 활성 상태의 경쟁 데이터만 가져옵니다
func (d *Database) GetActiveCompetitionsV2() ([]CompetitionV2, error) {
	if d.debug {
		log.Println("GetActiveCompetitionsV2 시작")
	}

	rows, err := d.db.Query(
		`SELECT 
			id, name, preview_image_url, description, 
			topic_id, prize_pool, start_date, end_date, season_id, 
			tags, is_active 
		FROM competitions_v2 
		WHERE is_active = 1
		ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("데이터 조회 실패: %w", err)
	}
	defer rows.Close()

	var competitions []CompetitionV2
	for rows.Next() {
		var comp CompetitionV2
		var startDateStr, endDateStr string
		var tagsJSON string

		err := rows.Scan(
			&comp.ID, &comp.Name, &comp.PreviewImageURL, &comp.Description,
			&comp.TopicID, &comp.PrizePool, &startDateStr, &endDateStr, &comp.SeasonID,
			&tagsJSON, &comp.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("데이터 스캔 실패: %w", err)
		}

		// 날짜 문자열을 시간으로 변환
		comp.StartDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, fmt.Errorf("시작 날짜 파싱 실패: %w", err)
		}

		comp.EndDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, fmt.Errorf("종료 날짜 파싱 실패: %w", err)
		}

		// 태그 JSON 파싱
		if err := json.Unmarshal([]byte(tagsJSON), &comp.Tags); err != nil {
			return nil, fmt.Errorf("태그 JSON 파싱 실패: %w", err)
		}

		competitions = append(competitions, comp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 처리 중 오류: %w", err)
	}

	if d.debug {
		log.Printf("활성 경쟁 데이터 조회 완료: %d개", len(competitions))
	}

	return competitions, nil
}

// GetCompetitionByTopicID는 토픽 ID에 해당하는 경쟁 정보를 competitions_v2 테이블에서 가져옵니다
func (d *Database) GetCompetitionByTopicID(topicID string) (*CompetitionV2, error) {
	if d.debug {
		log.Printf("GetCompetitionByTopicID 시작: 토픽 ID=%s", topicID)
	}

	// 토픽 ID를 정수로 변환
	topicIDInt, err := strconv.Atoi(topicID)
	if err != nil {
		return nil, fmt.Errorf("토픽 ID를 정수로 변환 실패: %w", err)
	}

	// competitions_v2 테이블에서 조회
	query := `
		SELECT 
			id, name, preview_image_url, description, 
			topic_id, prize_pool, start_date, end_date, season_id, 
			tags, is_active 
		FROM competitions_v2 
		WHERE topic_id = ? 
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var comp CompetitionV2
	var startDateStr, endDateStr string
	var tagsJSON string

	err = d.db.QueryRow(query, topicIDInt).Scan(
		&comp.ID, &comp.Name, &comp.PreviewImageURL, &comp.Description,
		&comp.TopicID, &comp.PrizePool, &startDateStr, &endDateStr, &comp.SeasonID,
		&tagsJSON, &comp.IsActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			if d.debug {
				log.Printf("토픽 ID %s에 해당하는 경쟁 데이터 없음", topicID)
			}
			return nil, fmt.Errorf("토픽 ID %s에 해당하는 경쟁 데이터가 없습니다", topicID)
		}
		return nil, fmt.Errorf("경쟁 데이터 조회 실패: %w", err)
	}

	// 날짜 문자열을 시간으로 변환
	comp.StartDate, err = time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		return nil, fmt.Errorf("시작 날짜 파싱 실패: %w", err)
	}

	comp.EndDate, err = time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("종료 날짜 파싱 실패: %w", err)
	}

	// 태그 JSON 파싱
	if err := json.Unmarshal([]byte(tagsJSON), &comp.Tags); err != nil {
		return nil, fmt.Errorf("태그 JSON 파싱 실패: %w", err)
	}

	if d.debug {
		log.Printf("토픽 ID %s의 경쟁 정보 조회 완료: ID=%d, 이름=%s", topicID, comp.ID, comp.Name)
	}

	return &comp, nil
}

// SaveLeaderboardEntries는 토픽 ID와 블록 높이에 해당하는 리더보드 데이터를 저장합니다
func (d *Database) SaveLeaderboardEntries(topicID string, blockHeight string, entries []map[string]interface{}) error {
	if d.debug {
		log.Printf("SaveLeaderboardEntries 시작: 토픽 ID=%s, 블록 높이=%s, 항목 수=%d", topicID, blockHeight, len(entries))
	}

	// 현재 시간
	timestamp := time.Now().Format(time.RFC3339)

	// 트랜잭션 시작
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 기존 데이터 삭제 (같은 토픽 ID와 블록 높이에 대한 데이터)
	_, err = tx.Exec(
		"DELETE FROM leaderboard_entries WHERE topic_id = ? AND inference_block_height = ?",
		topicID, blockHeight,
	)
	if err != nil {
		return fmt.Errorf("기존 리더보드 데이터 삭제 실패: %w", err)
	}

	// 새 데이터 삽입
	stmt, err := tx.Prepare(`
		INSERT INTO leaderboard_entries (
			topic_id, inference_block_height, timestamp, 
			cosmos_address, username, first_name, last_name, 
			rank, points, score, loss, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("SQL 준비 실패: %w", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		// 필수 필드 확인
		cosmosAddress, ok := entry["cosmos_address"].(string)
		if !ok || cosmosAddress == "" {
			if d.debug {
				log.Printf("cosmos_address 필드가 없거나 유효하지 않은 항목 건너뜀")
			}
			continue
		}

		// 선택적 필드 (기본값 설정)
		username := getStringValue(entry, "username", "")
		firstName := getStringValue(entry, "first_name", "")
		lastName := getStringValue(entry, "last_name", "")
		rank := getStringValue(entry, "rank", "")

		points := getFloatValue(entry, "points", 0)
		score := getFloatValue(entry, "score", 0)
		loss := getFloatValue(entry, "loss", 0)

		isActive := getBoolValue(entry, "is_active", false)

		_, err = stmt.Exec(
			topicID, blockHeight, timestamp,
			cosmosAddress, username, firstName, lastName,
			rank, points, score, loss, isActive,
		)
		if err != nil {
			return fmt.Errorf("리더보드 항목 삽입 실패: %w", err)
		}
	}

	// 트랜잭션 커밋
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	if d.debug {
		log.Printf("리더보드 데이터 저장 완료: 토픽 ID=%s, 블록 높이=%s, 항목 수=%d", topicID, blockHeight, len(entries))
	}

	return nil
}

// GetLeaderboardEntries는 토픽 ID와 블록 높이에 해당하는 리더보드 데이터를 조회합니다
func (d *Database) GetLeaderboardEntries(topicID string, blockHeight string) ([]map[string]interface{}, error) {
	if d.debug {
		log.Printf("GetLeaderboardEntries 시작: 토픽 ID=%s, 블록 높이=%s", topicID, blockHeight)
	}

	// 쿼리 실행
	rows, err := d.db.Query(`
		SELECT 
			cosmos_address, username, first_name, last_name, 
			rank, points, score, loss, is_active
		FROM leaderboard_entries 
		WHERE topic_id = ? AND inference_block_height = ?
		ORDER BY CAST(rank AS INTEGER)
	`, topicID, blockHeight)
	if err != nil {
		return nil, fmt.Errorf("리더보드 데이터 조회 실패: %w", err)
	}
	defer rows.Close()

	var entries []map[string]interface{}
	for rows.Next() {
		var cosmosAddress, username, firstName, lastName, rank string
		var points, score, loss float64
		var isActive bool

		err := rows.Scan(
			&cosmosAddress, &username, &firstName, &lastName,
			&rank, &points, &score, &loss, &isActive,
		)
		if err != nil {
			return nil, fmt.Errorf("리더보드 데이터 스캔 실패: %w", err)
		}

		entry := map[string]interface{}{
			"cosmos_address": cosmosAddress,
			"username":       username,
			"first_name":     firstName,
			"last_name":      lastName,
			"rank":           rank,
			"points":         points,
			"score":          score,
			"loss":           loss,
			"is_active":      isActive,
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("결과 처리 중 오류: %w", err)
	}

	if d.debug {
		log.Printf("리더보드 데이터 조회 완료: 토픽 ID=%s, 블록 높이=%s, 항목 수=%d", topicID, blockHeight, len(entries))
	}

	return entries, nil
}

// 유틸리티 함수: 맵에서 문자열 값 가져오기
func getStringValue(data map[string]interface{}, key string, defaultValue string) string {
	if value, ok := data[key].(string); ok {
		return value
	}
	return defaultValue
}

// 유틸리티 함수: 맵에서 실수 값 가져오기
func getFloatValue(data map[string]interface{}, key string, defaultValue float64) float64 {
	switch value := data[key].(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case string:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// 유틸리티 함수: 맵에서 불리언 값 가져오기
func getBoolValue(data map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := data[key].(bool); ok {
		return value
	}
	return defaultValue
}
