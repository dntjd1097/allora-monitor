package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/glebarez/go-sqlite" // 순수 Go로 작성된 SQLite 드라이버
	"github.com/golang/snappy"
)

func main() {
	// 데이터베이스 파일 경로
	dbPath := "data/allora-monitor.db"

	// 데이터베이스 연결
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("데이터베이스 연결 실패: %v", err)
	}
	defer db.Close()

	// 테이블 목록 확인
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		log.Fatalf("테이블 목록 조회 실패: %v", err)
	}
	defer rows.Close()

	fmt.Println("데이터베이스 테이블 목록:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatalf("테이블 이름 스캔 실패: %v", err)
		}
		fmt.Printf("- %s\n", tableName)
	}
	fmt.Println()

	// competitions 테이블 데이터 조회
	rows, err = db.Query("SELECT id, timestamp, data FROM competitions ORDER BY timestamp DESC LIMIT 1")
	if err != nil {
		log.Fatalf("데이터 조회 실패: %v", err)
	}
	defer rows.Close()

	fmt.Println("최근 경쟁 데이터:")
	for rows.Next() {
		var id int
		var timestamp string
		var compressedData []byte

		if err := rows.Scan(&id, &timestamp, &compressedData); err != nil {
			log.Fatalf("데이터 스캔 실패: %v", err)
		}

		fmt.Printf("ID: %d, 타임스탬프: %s, 압축 데이터 크기: %d 바이트\n",
			id, timestamp, len(compressedData))

		// 데이터 압축 해제
		jsonData, err := snappy.Decode(nil, compressedData)
		if err != nil {
			log.Fatalf("압축 해제 실패: %v", err)
		}

		fmt.Printf("압축 해제 후 크기: %d 바이트\n", len(jsonData))

		// 압축 해제된 JSON 데이터를 파일로 저장
		outputFile := "data_decoded.json"
		if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
			log.Fatalf("파일 저장 실패: %v", err)
		}
		fmt.Printf("압축 해제된 데이터가 %s 파일에 저장되었습니다\n", outputFile)

		// JSON 데이터 구조 확인
		var result map[string]interface{}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			log.Fatalf("JSON 언마샬링 실패: %v", err)
		}

		// 데이터 구조 출력
		fmt.Println("\n데이터 구조:")
		printStructure(result, 0)
	}

	// 데이터베이스 통계 출력
	var count int
	var totalSize int64
	db.QueryRow("SELECT COUNT(*) FROM competitions").Scan(&count)
	db.QueryRow("SELECT SUM(LENGTH(data)) FROM competitions").Scan(&totalSize)

	fmt.Printf("\n총 레코드 수: %d, 총 데이터 크기: %.2f MB\n",
		count, float64(totalSize)/(1024*1024))
}

// 중첩된 데이터 구조를 출력하는 함수
func printStructure(data interface{}, level int) {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fmt.Printf("%s%s: ", indent, key)

			switch val := value.(type) {
			case map[string]interface{}, []interface{}:
				fmt.Println()
				printStructure(val, level+1)
			default:
				fmt.Printf("%v\n", value)
			}
		}
	case []interface{}:
		if len(v) > 0 {
			fmt.Printf("%s배열 (항목 수: %d)\n", indent, len(v))
			// 첫 번째 항목만 출력
			if len(v) > 0 {
				fmt.Printf("%s첫 번째 항목:\n", indent)
				printStructure(v[0], level+1)
			}
		} else {
			fmt.Printf("%s빈 배열\n", indent)
		}
	default:
		fmt.Printf("%s%v\n", indent, v)
	}
}
