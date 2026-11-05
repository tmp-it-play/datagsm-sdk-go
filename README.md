# DataGSM OpenAPI SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/jihoonwjj/datagsm-openapi-sdk-go.svg)](https://pkg.go.dev/github.com/jihoonwjj/datagsm-openapi-sdk-go)

[DataGSM OpenAPI](https://docs.datagsm.kr/)를 위한 파트너 Go SDK입니다.  
학생·동아리·프로젝트·급식/학사 데이터를 Go 코드에서 간편하게 조회할 수 있습니다.

## 요구사항

- Go 1.21 이상
- DataGSM API 키 ([발급 방법](https://datagsm-front-client.vercel.app))

## 설치

```bash
go get github.com/jihoonwjj/datagsm-openapi-sdk-go
```

## 빠른 시작

```go
package main

import (
    "context"
    "fmt"
    "log"

    datagsm "github.com/jihoonwjj/datagsm-openapi-sdk-go"
)

func main() {
    client, err := datagsm.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // 학생 목록 조회
    resp, err := client.Students().List(ctx, datagsm.StudentQuery{
        Grade: datagsm.Ptr(1),
        Page:  0,
        Size:  20,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("학생 수: %d명\n", resp.TotalElements)
    for _, s := range resp.Students {
        fmt.Printf("  %s (%d학년 %d반)\n", s.Name, s.Grade, s.ClassNum)
    }
}
```

## 클라이언트 설정

```go
client, err := datagsm.NewClient("your-api-key",
    // 커스텀 베이스 URL (기본값: https://openapi.datagsm.kr)
    datagsm.WithBaseURL("https://openapi.datagsm.kr"),
    // 타임아웃 설정 (기본값: 30초)
    datagsm.WithTimeout(10 * time.Second),
    // 커스텀 http.Client 주입
    datagsm.WithHTTPClient(myHTTPClient),
)
```

## API 예시

### 학생

```go
// 목록 조회 (필터·페이징·정렬)
resp, err := client.Students().List(ctx, datagsm.StudentQuery{
    Grade:    datagsm.Ptr(2),
    Major:    (*datagsm.Major)(datagsm.Ptr(string(datagsm.MajorAI))),
    SortBy:   datagsm.StudentSortByName,
    SortDir:  datagsm.SortDirASC,
    Page:     0,
    Size:     50,
})

// 단건 조회
student, err := client.Students().Get(ctx, 123)
```

### 동아리

```go
// 목록 조회
resp, err := client.Clubs().List(ctx, datagsm.ClubQuery{
    ClubType: (*datagsm.ClubType)(datagsm.Ptr(string(datagsm.ClubTypeMajor))),
    Status:   (*datagsm.ClubStatus)(datagsm.Ptr(string(datagsm.ClubStatusActive))),
})

// 상세 조회 (리더·참여자 포함)
club, err := client.Clubs().Get(ctx, 42)
```

### 프로젝트

```go
resp, err := client.Projects().List(ctx, datagsm.ProjectQuery{Page: 0, Size: 20})
proj, err := client.Projects().Get(ctx, 1)
```

### NEIS (급식·학사일정·시간표)

```go
// 오늘 급식
today := time.Now()
meals, err := client.NEIS().Meals(ctx, datagsm.MealQuery{Date: &today})

// 날짜 범위 학사일정
from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
to   := time.Date(2026, 4, 30, 0, 0, 0, 0, time.Local)
schedules, err := client.NEIS().Schedules(ctx, datagsm.ScheduleQuery{
    FromDate: &from,
    ToDate:   &to,
})

// 시간표 (Grade, ClassNum 필수)
timetables, err := client.NEIS().Timetables(ctx, datagsm.TimetableQuery{
    Grade:    1,
    ClassNum: 2,
    Date:     &today,
})
```

## 에러 처리

```go
resp, err := client.Students().List(ctx, query)
if err != nil {
    var apiErr *datagsm.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.StatusCode {
        case 401:
            fmt.Println("API 키가 유효하지 않거나 만료되었습니다")
        case 403:
            fmt.Println("권한이 없습니다")
        case 429:
            fmt.Println("요청 속도 제한 초과")
        default:
            fmt.Printf("API 오류 %d: %s\n", apiErr.StatusCode, apiErr.Message)
        }
    }
}
```

| 상태 코드 | 의미 |
|-----------|------|
| 400 | 잘못된 요청 파라미터 |
| 401 | 유효하지 않거나 만료된 API 키 |
| 403 | 권한 부족 |
| 429 | 요청 속도 제한 초과 |
| 5xx | 서버 오류 |

## Ptr 헬퍼

선택적 필드에 포인터를 인라인으로 생성할 수 있습니다:

```go
datagsm.Ptr(1)          // *int
datagsm.Ptr("hello")    // *string
datagsm.Ptr(true)       // *bool
```

## 라이선스

MIT License — Copyright (c) 2026 황지훈
