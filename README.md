# DataGSM SDK for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/tmp-it-play/datagsm-sdk-go.svg)](https://pkg.go.dev/github.com/tmp-it-play/datagsm-sdk-go)

DataGSM의 OpenAPI, OAuth·OIDC, Event를 지원하는 파트너 Go SDK입니다.

## 설치

Go 1.25 이상이 필요합니다.

```sh
go get github.com/tmp-it-play/datagsm-sdk-go@v0.3.0
```

0.x 버전에서는 마이너 버전 업데이트에 호환되지 않는 API 변경이 포함될 수 있습니다. 업데이트 전에 [릴리스 노트](https://github.com/tmp-it-play/datagsm-sdk-go/releases)을 확인하세요.

## 지원 기능

| 기능 | 용도 | 준비 사항 |
|---|---|---|
| OpenAPI | 학생·동아리·프로젝트 조회 및 관리, 급식·학사일정·시간표 조회 | API 키 |
| OAuth·OIDC | 사용자 로그인, 토큰 발급·갱신, 사용자 정보 조회, 토큰 검증 | OAuth 클라이언트 등록 |
| Event | 학생·동아리·프로젝트 변경 알림 수신 및 서명 검증 | 이벤트 구독 등록과 secret 보관 |

사용할 기능에 맞춰 DataGSM 콘솔에서 인증 정보와 구독을 준비하세요.

## OpenAPI

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    datagsm "github.com/tmp-it-play/datagsm-sdk-go"
    "github.com/tmp-it-play/datagsm-sdk-go/openapi"
)

func main() {
    client, err := openapi.NewClient(openapi.Config{
        APIKey: os.Getenv("DATAGSM_API_KEY"),
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    page, err := client.Students.List(context.Background(), openapi.StudentQuery{
        Grade: datagsm.Ptr(1),
        Size: 20,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(page.TotalElements)
}
```

`Config.BaseURL`의 기본값은 `https://openapi.datagsm.kr`입니다. 타임아웃이나 전송 설정은 `Config.HTTPClient`로 지정합니다. 생략하면 타임아웃은 30초입니다. 리다이렉트는 자동으로 따라가지 않습니다. URL은 HTTPS를 사용하고 로컬 루프백 주소에는 HTTP를 허용합니다.

### 조회

```go
students, err := client.Students.List(ctx, openapi.StudentQuery{
    StudentID: datagsm.Ptr(int64(123)),
    IncludeGraduates: datagsm.Ptr(true),
    IncludeWithdrawn: datagsm.Ptr(true),
})

clubs, err := client.Clubs.List(ctx, openapi.ClubQuery{
    ClubStatus: datagsm.Ptr(datagsm.ClubStatusActive),
    ClubType: datagsm.Ptr(datagsm.ClubTypeMajor),
})

projects, err := client.Projects.List(ctx, openapi.ProjectQuery{
    Status: datagsm.Ptr(datagsm.ProjectStatusEnded),
    Category: datagsm.Ptr(datagsm.ProjectCategoryTeam),
})

meals, err := client.NEIS.Meals(ctx, openapi.DateQuery{Date: "2026-11-05"})
schedules, err := client.NEIS.Schedules(ctx, openapi.DateQuery{
    FromDate: "2026-11-01", ToDate: "2026-11-30",
})
timetables, err := client.NEIS.Timetables(ctx, openapi.TimetableQuery{
    Grade: 1, ClassNum: 2, Date: "2026-11-05",
})
```

각 호출의 오류를 확인한 뒤 결과를 사용합니다. 학생·동아리·프로젝트는 목록 API에 ID 필터를 보내 조회합니다. 학생은 DataGSM 계정이 연결된 경우에만 조회됩니다. 졸업·자퇴 포함 여부와 프로젝트 운영 상태는 서버의 기본 필터를 따릅니다.

페이지는 0부터 시작합니다. `Size: 0`이면 서버 기본값을 사용합니다. 학생의 학번·학과 등 nullable 필드는 포인터입니다. NEIS 날짜와 학년도는 문자열이며, 시간표의 학기는 `*string`입니다.

날짜는 `YYYY-MM-DD` 형식입니다. 단일 날짜와 범위를 함께 지정할 수 없습니다. 양 끝을 지정한 범위는 최대 365일입니다. 시간표에는 날짜 또는 범위가 필요합니다.

### 쓰기

```go
student, err := client.Students.Create(ctx, openapi.CreateStudentRequest{
    Name: "홍길동", Sex: datagsm.SexMan, Email: "student@gsm.hs.kr",
    Grade: datagsm.Ptr(1), ClassNum: datagsm.Ptr(1), Number: datagsm.Ptr(1),
})

project, err := client.Projects.Create(ctx, openapi.ProjectRequest{
    Name: "학교 서비스", Description: "학생을 위한 서비스", StartYear: 2026,
    ParticipantIDs: []int64{1, 2},
    Categories: []datagsm.ProjectCategory{datagsm.ProjectCategoryTeam},
})
```

쓰기에는 `student:write`, `club:write`, `project:write` 또는 해당 도메인의 와일드카드 권한을 가진 관리자 API 키가 필요합니다. 이름 길이, 역할별 조건, 참조 대상 등의 검증은 서버에서 수행하고 `APIError`로 반환합니다.

- `Students.Update`는 학생 정보를 교체합니다. 생략한 선택 필드는 비워집니다.
- `Clubs.Update`는 구성원 목록까지 교체합니다. 유지할 부장과 부원도 모두 보냅니다.
- `Projects.Update`에서 `Categories: nil`은 유지, 빈 슬라이스는 삭제입니다. `IconKey`·`DeploymentURL`은 `nil`이면 유지, `Ptr("")`이면 삭제입니다.
- 프로젝트의 `Repositories`, `TechStacks`, `ClubID`는 생략하면 비워집니다.
- 프로젝트 상태는 `Projects.End(ctx, id, endYear)`와 `Projects.Reactivate(ctx, id)`로 변경합니다.
- `Clubs.Delete`와 `Projects.Delete`는 데이터를 영구 삭제합니다.

## OAuth·OIDC

```go
client, err := oauth.NewClient(oauth.Config{
    ClientID: os.Getenv("DATAGSM_CLIENT_ID"),
    ClientSecret: os.Getenv("DATAGSM_CLIENT_SECRET"),
    RedirectURI: "https://your-app.example/callback",
})
if err != nil {
    return err
}
defer client.Close()

attempt, err := client.NewAuthorization(
    oauth.ScopeOpenID, oauth.ScopeAccountRead, oauth.ScopeStudentRead,
)
if err != nil {
    return err
}
```

`attempt`를 사용자별 서버 세션에 저장하고 브라우저를 `attempt.URL`로 보냅니다. 콜백에서는 세션에서 한 번만 꺼내 `attempt.Code(r.URL.Query())`로 `state`와 공급자 오류를 확인합니다. 반환된 코드와 `attempt.CodeVerifier`를 `client.Exchange`에 전달합니다.

`client.JWKS`로 받은 키의 `VerifyIDToken`에 ID 토큰, 예상 issuer, 자신의 Client ID, 저장한 `attempt.Nonce`, 현재 시각을 전달합니다. 검증 후 `client.UserInfo`의 `Subject`가 ID 토큰의 `Subject`와 같은지 확인합니다. [콜백 예제](examples/oauth/flow.go)를 참고하세요.

```go
token, err := client.Refresh(ctx, savedRefreshToken)
serviceToken, err := client.ClientCredentials(ctx, "your-app:profile")
user, err := client.UserInfo(ctx, accessToken)
discovery, err := client.Discovery(ctx)
keys, err := client.JWKS(ctx)
```

기본 인증은 시크릿이 있으면 `client_secret_post`, 없으면 `none`입니다. `AuthMethod: oauth.AuthSecretBasic`으로 HTTP Basic을 선택할 수 있습니다. 토큰 요청 본문은 기본적으로 form 형식이며, `RequestFormat: oauth.RequestFormatJSON`으로 JSON 형식을 선택할 수 있습니다. 기본 인가 서버와 리소스 서버는 각각 `https://oauth.authorization.datagsm.kr`, `https://oauth.resource.datagsm.kr`입니다.

`VerifyAccessToken`은 RS256 서명, `clientId`, 유효 시간과 필요한 권한 범위를 확인합니다. DataGSM 액세스 토큰에는 `iss`·`aud`가 없습니다. 키는 신뢰하는 인가 서버에서 받아 사용하며, 키 갱신과 캐시는 호출자가 관리합니다.

JWT의 `iat`·`nbf`·`exp` 검증에는 기본 5초(`oauth.DefaultClockSkew`)의 허용오차를 적용합니다. `oauth.Config`의 `ClockSkew`로 값을 바꾸면 `client.JWKS`가 반환하는 키에 적용됩니다. 직접 만든 `JWKS`에는 `ClockSkew` 필드를 설정합니다. 음수를 지정하면 허용오차 없이 검증합니다.

토큰과 세션의 저장·만료 처리는 애플리케이션에서 합니다. 갱신 요청은 같은 세션에서 직렬로 처리하고 새 Refresh Token으로 교체합니다. 자동 재시도나 자동 갱신은 하지 않습니다.

`Logout`은 주입한 HTTP 클라이언트의 쿠키 저장소에 있는 SSO 세션에만 적용됩니다. 브라우저의 쿠키나 발급된 토큰은 지우지 않습니다. 외부 서비스의 로그아웃은 자체 세션과 저장한 토큰을 삭제해서 처리합니다.

## Event

```go
received, err := event.Parse(rawBody, signatureHeader, secret)
if err != nil {
    return err
}
switch received.Type {
case event.StudentUpdated:
    changes, err := event.Changes[event.Student](received)
    if err != nil {
        return err
    }
    for _, change := range changes {
        fmt.Println(change.New.ID, change.New.Role)
    }
}
```

`rawBody`는 HTTP 요청 본문 원문 바이트, `signatureHeader`는 `X-DataGSM-Signature` 값입니다. secret은 hex로 디코딩하지 않고 문자열 그대로 사용합니다.

`Changes`는 `index`가 같은 변경 전후 데이터를 짝지어 반환합니다. 동아리·프로젝트 생성은 `Old == nil`, 삭제는 `New == nil`입니다. 학생 졸업·자퇴는 역할 변경으로 전달됩니다. 알려지지 않은 열거형 값도 문자열로 보존합니다.

[HTTP 수신 예제](examples/event/handler.go)는 본문 크기를 제한하고, 검증과 처리가 성공한 경우에만 204를 반환합니다. 애플리케이션은 이벤트 ID로 중복 처리를 막아야 합니다. 서명 검증만으로 재전송을 막을 수는 없습니다. 구독 등록과 수신 URL 검증은 [Event 콘솔](https://www.datagsm.kr/events)에서 합니다.

## 오류 처리

```go
var apiErr *datagsm.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message, apiErr.RetryAfter)
}
```

`Code`에는 OAuth 오류 코드가 들어갑니다. OpenAPI·UserInfo의 공통 오류는 HTTP 상태 코드와 `Message`로 확인합니다. `RetryAfter`는 서버 헤더의 원문입니다.

## 라이선스

[MIT License](LICENSE) — Copyright (c) 2026 황지훈
