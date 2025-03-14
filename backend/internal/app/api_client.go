package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

// AlloraAPIClient는 Allora API와 통신하는 클라이언트입니다
type AlloraAPIClient struct {
	httpClient *http.Client
	baseURL    string
	debug      bool // 디버깅 모드 활성화 여부
}

// CompetitionsResponse는 경쟁 데이터 응답 구조체입니다
type CompetitionsResponse struct {
	PageProps struct {
		CompetitionsPage struct {
			ActiveAndUpcomingCompetitions []Competition `json:"activeAndUpcomingCompetitions"`
			PastCompetitions              []Competition `json:"pastCompetitions"`
		} `json:"competitionsPage"`
	} `json:"pageProps"`
}

// Competition은 경쟁 정보 구조체입니다
type Competition struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	PreviewImageURL     string    `json:"preview_image_url"`
	Description         *string   `json:"description"`
	DetailedDescription string    `json:"detailed_description"`
	TopicID             int       `json:"topic_id"`
	PrizePool           int       `json:"prize_pool"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	SeasonID            int       `json:"season_id"`
	Tags                []string  `json:"tags"`
}

// LeaderboardResponse는 리더보드 데이터 응답 구조체입니다
type LeaderboardResponse struct {
	RequestID string `json:"request_id"`
	Status    bool   `json:"status"`
	Data      struct {
		Leaderboard       []LeaderboardEntry `json:"leaderboard"`
		ContinuationToken string             `json:"continuation_token"`
	} `json:"data"`
}

// LeaderboardEntry는 리더보드 항목 구조체입니다
type LeaderboardEntry struct {
	Rank          string  `json:"rank"`
	CosmosAddress string  `json:"cosmos_address"`
	Username      string  `json:"username"`
	FirstName     *string `json:"first_name"`
	LastName      *string `json:"last_name"`
	Points        float64 `json:"points"`
	Score         float64 `json:"score"`
	Loss          float64 `json:"loss"`
	IsActive      bool    `json:"is_active"`
}

// NewAlloraAPIClient는 새로운 Allora API 클라이언트를 생성합니다
func NewAlloraAPIClient(baseURL string, timeout time.Duration) *AlloraAPIClient {
	return &AlloraAPIClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		debug:   true, // 디버깅 모드 활성화
	}
}

// SetDebug는 디버깅 모드를 설정합니다
func (c *AlloraAPIClient) SetDebug(debug bool) {
	c.debug = debug
}

// FetchCompetitions는 경쟁 데이터를 가져옵니다
func (c *AlloraAPIClient) FetchCompetitions() (*CompetitionsResponse, error) {
	// 먼저 메인 페이지를 요청하여 최신 빌드 ID를 추출
	buildID, err := c.getBuildID()
	if err != nil {
		return nil, fmt.Errorf("빌드 ID 추출 실패: %w", err)
	}

	if c.debug {
		log.Printf("추출된 빌드 ID: %s", buildID)
	}

	// 추출한 빌드 ID로 데이터 URL 구성
	url := fmt.Sprintf("%s/_next/data/%s/competitions.json", c.baseURL, buildID)

	if c.debug {
		log.Printf("요청 URL: %s", url)
	}

	// HTTP GET 요청
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 응답 오류: %d %s", resp.StatusCode, resp.Status)
	}

	// 응답 본문 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 본문 읽기 실패: %w", err)
	}

	if c.debug {
		// 디버깅 모드에서는 응답 본문을 파일로 저장
		debugFile := "debug_response.json"
		if err := os.WriteFile(debugFile, body, 0644); err != nil {
			log.Printf("디버그 파일 저장 실패: %v", err)
		} else {
			log.Printf("응답 본문이 %s 파일에 저장되었습니다", debugFile)
		}

		// 응답 본문 일부 출력
		if len(body) > 1000 {
			log.Printf("응답 본문 일부: %s...", body[:1000])
		} else {
			log.Printf("응답 본문: %s", body)
		}
	}

	// JSON 디코딩
	var competitionsResp CompetitionsResponse
	if err := json.Unmarshal(body, &competitionsResp); err != nil {
		return nil, fmt.Errorf("JSON 디코딩 실패: %w", err)
	}

	// 디버깅: 데이터 구조 확인
	if c.debug {
		activeCount := len(competitionsResp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions)
		pastCount := len(competitionsResp.PageProps.CompetitionsPage.PastCompetitions)
		log.Printf("활성 경쟁 수: %d, 과거 경쟁 수: %d", activeCount, pastCount)

		if activeCount > 0 {
			firstComp := competitionsResp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions[0]
			log.Printf("첫 번째 활성 경쟁: ID=%d, 이름=%s", firstComp.ID, firstComp.Name)
		}
	}

	return &competitionsResp, nil
}

// getBuildID는 메인 페이지에서 Next.js 빌드 ID를 추출합니다
func (c *AlloraAPIClient) getBuildID() (string, error) {
	// 경쟁 페이지 요청
	competitionsURL := c.baseURL + "/competitions"

	if c.debug {
		log.Printf("경쟁 페이지 URL: %s", competitionsURL)
	}

	resp, err := c.httpClient.Get(competitionsURL)
	if err != nil {
		return "", fmt.Errorf("경쟁 페이지 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("경쟁 페이지 응답 오류: %d %s", resp.StatusCode, resp.Status)
	}

	// 응답 본문 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("경쟁 페이지 본문 읽기 실패: %w", err)
	}

	if c.debug {
		// 디버깅 모드에서는 HTML 일부 출력
		if len(body) > 1000 {
			log.Printf("HTML 일부: %s...", body[:1000])
		}
	}

	// HTML에서 빌드 ID 추출
	// Next.js는 일반적으로 __NEXT_DATA__ 스크립트 태그에 빌드 ID를 포함
	re := regexp.MustCompile(`"buildId":"([^"]+)"`)
	matches := re.FindSubmatch(body)

	if len(matches) < 2 {
		if c.debug {
			// 디버깅 모드에서는 HTML을 파일로 저장
			debugFile := "debug_html.html"
			if err := os.WriteFile(debugFile, body, 0644); err != nil {
				log.Printf("디버그 HTML 파일 저장 실패: %v", err)
			} else {
				log.Printf("HTML이 %s 파일에 저장되었습니다", debugFile)
			}
		}
		return "", fmt.Errorf("빌드 ID를 찾을 수 없습니다")
	}

	return string(matches[1]), nil
}

// DirectFetchCompetitions는 지정된 URL에서 직접 경쟁 데이터를 가져옵니다 (디버깅용)
func (c *AlloraAPIClient) DirectFetchCompetitions(fullURL string) (*CompetitionsResponse, error) {
	if c.debug {
		log.Printf("직접 URL 요청: %s", fullURL)
	}

	// HTTP GET 요청
	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 응답 오류: %d %s", resp.StatusCode, resp.Status)
	}

	// 응답 본문 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 본문 읽기 실패: %w", err)
	}

	if c.debug {
		// 디버깅 모드에서는 응답 본문을 파일로 저장
		debugFile := "debug_direct_response.json"
		if err := os.WriteFile(debugFile, body, 0644); err != nil {
			log.Printf("디버그 파일 저장 실패: %v", err)
		} else {
			log.Printf("응답 본문이 %s 파일에 저장되었습니다", debugFile)
		}

		// 응답 본문 일부 출력
		if len(body) > 1000 {
			log.Printf("응답 본문 일부: %s...", body[:1000])
		} else {
			log.Printf("응답 본문: %s", body)
		}
	}

	// JSON 디코딩
	var competitionsResp CompetitionsResponse
	if err := json.Unmarshal(body, &competitionsResp); err != nil {
		return nil, fmt.Errorf("JSON 디코딩 실패: %w", err)
	}

	// 디버깅: 데이터 구조 확인
	if c.debug {
		activeCount := len(competitionsResp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions)
		pastCount := len(competitionsResp.PageProps.CompetitionsPage.PastCompetitions)
		log.Printf("활성 경쟁 수: %d, 과거 경쟁 수: %d", activeCount, pastCount)

		if activeCount > 0 {
			firstComp := competitionsResp.PageProps.CompetitionsPage.ActiveAndUpcomingCompetitions[0]
			log.Printf("첫 번째 활성 경쟁: ID=%d, 이름=%s", firstComp.ID, firstComp.Name)
		}
	}

	return &competitionsResp, nil
}

// FetchLeaderboard는 특정 경쟁의 리더보드 데이터를 가져옵니다
func (c *AlloraAPIClient) FetchLeaderboard(competitionID string, continuationToken string) (*LeaderboardResponse, error) {
	var url string
	if continuationToken == "" {
		url = fmt.Sprintf("https://forge.allora.network/api/upshot-api-proxy/allora/forge/competition/%s/leaderboard", competitionID)
	} else {
		url = fmt.Sprintf("https://forge.allora.network/api/upshot-api-proxy/allora/forge/competition/%s/leaderboard?continuation_token=%s", competitionID, continuationToken)
	}

	if c.debug {
		log.Printf("리더보드 요청 URL: %s", url)
	}

	// HTTP GET 요청
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("리더보드 API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	// 응답 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("리더보드 API 응답 오류: %d %s", resp.StatusCode, resp.Status)
	}

	// 응답 본문 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("리더보드 응답 본문 읽기 실패: %w", err)
	}

	if c.debug {
		// 디버깅 모드에서는 응답 본문을 파일로 저장
		err = os.WriteFile("leaderboard_response.json", body, 0644)
		if err != nil {
			log.Printf("리더보드 응답 저장 실패: %v", err)
		}
	}

	// JSON 파싱
	var leaderboardResp LeaderboardResponse
	if err := json.Unmarshal(body, &leaderboardResp); err != nil {
		return nil, fmt.Errorf("리더보드 JSON 파싱 실패: %w", err)
	}

	return &leaderboardResp, nil
}
