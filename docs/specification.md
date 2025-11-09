🧠 GoProject — Universal AI Gateway Specification

版本：v1.0  
作者：李博凱  
語言：Golang  
風格：Minimalist & Stable  
目的：提供一個極簡、高效、可自由切換 AI Provider 的終界服務

1. 📘 專案概述

GoProject 是一個以 Go 語言開發的 AI 終界服務（Universal AI Gateway），目標是讓開發者可以透過極簡的方式在本地或雲端快速部署一個能代理各種 AI 服務供應商（OpenAI、Gemini、Claude、Azure OpenAI…）的統一 API。

此專案的設計理念是 「極簡、高速、穩定、零阻力接入」。

開發者只需：
1. 執行 `go run main.go`
2. 修改 `.env` 內的 `PROVIDER_URL`

即可切換底層 AI 供應商。

2. 🎯 專案目標與設計原則

2.1 核心目標
- 提供 單一 API 終界（endpoint），統一對接各種 AI 模型服務。
- 實現 零痛點切換，透過簡單設定即可更換 AI provider。
- 強調 穩定與簡潔，所有功能皆保持「單一職責原則」。
- 可 快速部署、可 擴充多家服務商支援。

2.2 設計原則

原則 | 說明
--- | ---
KISS | Keep It Simple & Stupid — 結構簡單、邏輯清晰。
MVP First | 先確保最小可行功能穩定，後續再擴充。
Provider Agnostic | 不與特定服務商耦合，支援動態切換。
Zero Bug Goal | 嚴格輸入驗證與防錯設計。
可插拔架構 | 透過 interface 註冊新 Provider handler。

3. ⚙️ 系統架構

3.1 架構圖

[Client App / CLI / SDK]  
          │  
          ▼  
 ┌────────────────────────┐  
 │   GoProject Gateway     │  
 │────────────────────────│  
 │ Router (HTTP Server)    │  
 │  ↓                      │  
 │ Handler Interface        │  
 │  ↓                      │  
 │ Provider Handlers        │  
 │   - OpenAIHandler        │  
 │   - GeminiHandler        │  
 │   - ClaudeHandler        │  
 │   - CustomHandler        │  
 │                         │  
 │ Config / Utils / Logger │  
 └────────────────────────┘  
          │  
          ▼  
 [AI Provider APIs]

4. 🧩 功能規格

4.1 API 終界（Endpoints）

Endpoint | Method | 說明
--- | --- | ---
/v1/chat | POST | 傳入使用者對話內容，轉發至目前設定的 AI Provider。
/v1/config | GET | 查詢目前 Gateway 的 Provider 設定與狀態。
/v1/health | GET | 健康檢查用（回傳 service 狀態）。

4.2 Request 格式（統一格式）

```json
{
  "model": "gpt-4-turbo",
  "messages": [
    {"role": "user", "content": "Hello! What can you do?"}
  ],
  "temperature": 0.7,
  "stream": false
}
```

4.3 Response 格式（統一格式）

```json
{
  "id": "gateway-abc123",
  "object": "chat.completion",
  "created": 1731200000,
  "model": "gpt-4-turbo",
  "provider": "openai",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! I'm your AI assistant from GoProject Gateway."
      },
      "finish_reason": "stop"
    }
  ]
}
```

5. 🧠 Provider 模組設計

5.1 Handler Interface 定義

```go
type AIProvider interface {
    SendRequest(ctx context.Context, payload []byte) ([]byte, error)
    ProviderName() string
}
```

5.2 Provider Example — OpenAI

```go
type OpenAIHandler struct {
    APIKey string
    BaseURL string
}

func (o *OpenAIHandler) SendRequest(ctx context.Context, payload []byte) ([]byte, error) {
    req, _ := http.NewRequestWithContext(ctx, "POST", o.BaseURL+"/v1/chat/completions", bytes.NewBuffer(payload))
    req.Header.Set("Authorization", "Bearer "+o.APIKey)
    req.Header.Set("Content-Type", "application/json")

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    return io.ReadAll(res.Body)
}

func (o *OpenAIHandler) ProviderName() string {
    return "openai"
}
```

6. 🧾 設定（Config / Env）

.env 範例

```
PROVIDER=openai
PROVIDER_URL=https://api.openai.com/v1
API_KEY=sk-xxxx
PORT=8080
LOG_LEVEL=info
```

config.go 範例

```go
type Config struct {
    Provider    string
    ProviderURL string
    APIKey      string
    Port        string
}
```

7. 🧱 檔案結構

```
GoProject/
├── main.go
├── router.go
├── config.go
├── handlers/
│   ├── openai.go
│   ├── gemini.go
│   └── claude.go
├── utils/
│   ├── logger.go
│   ├── errors.go
│   └── response.go
├── go.mod
├── go.sum
└── .env
```

8. 🔍 錯誤與例外處理

錯誤代碼 | 說明
--- | ---
400 | 無效的請求格式或參數錯誤
401 | API 金鑰無效或未設定
502 | Provider 回傳錯誤或網路中斷
500 | Gateway 內部錯誤

所有錯誤皆統一包裝為：

```json
{
  "error": {
    "code": 502,
    "message": "Failed to contact provider: timeout"
  }
}
```

9. 🧪 測試與驗證

單元測試（Unit Tests）
- 測試 config 載入。
- 測試 Provider handler 是否能正確處理輸入與回傳。
- 模擬錯誤 API 回應的防護行為。

健康檢查
- `/v1/health` endpoint 回傳：

```json
{ "status": "ok", "uptime": "123s" }
```

10. 🚀 開發與部署

開發模式

```bash
go run main.go
```

建立執行檔

```bash
go build -o gateway
./gateway
```

Docker 部署

```dockerfile
FROM golang:1.22
WORKDIR /app
COPY . .
RUN go build -o gateway .
EXPOSE 8080
CMD ["./gateway"]
```

11. 🔮 未來規劃

階段 | 功能 | 狀態
--- | --- | ---
v1.0 | OpenAI Handler | ✅ 已完成
v1.1 | Gemini Handler | ⏳ 開發中
v1.2 | Claude Handler | ⏳ 規劃中
v1.3 | 多 API Key 輪替機制 | 🔜
v2.0 | Streaming 回傳模式 | 🔜
v2.1 | 前端管理介面（Web Dashboard） | 🔜

12. 📚 授權與貢獻
- 授權：MIT License
- 貢獻：歡迎 PR（需附帶單元測試）
- 維護者：李博凱（Lee Poka）


