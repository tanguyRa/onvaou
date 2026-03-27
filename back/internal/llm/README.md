# llm

```tree
llm/
├── README.md
├── anthropic.go
│   ├── type anthropicProvider {apiKey: string, model: string, client: *http.Client}
│   ├── type anthropicMessage {Role: string, Content: string}
│   ├── type anthropicRequest {Model: string, MaxTokens: int, System: string, Messages: []anthropicMessage, Stream: bool}
│   ├── type anthropicResponse {Content: []struct { Type string `json:"type"` Text string `json:"text"` }, StopReason: string}
│   ├── type anthropicStreamEvent {Type: string, Index: int, Delta: *struct { Type string `json:"type"` Text string `json:"text"` }}
│   ├── func getSharedHTTPClient() *http.Client
│   ├── func newAnthropicProvider(apiKey, model string) *anthropicProvider
│   ├── func (*anthropicProvider) Name() string
│   ├── func (*anthropicProvider) Chat(ctx context.Context, systemPrompt string, messages []Message) (string, error)
│   └── func (*anthropicProvider) ChatStream(ctx context.Context, systemPrompt string, messages []Message) (<-chan StreamChunk, error)
├── anthropic_test.go
│   ├── type testTransport {url: string}
│   ├── func TestAnthropicProvider_Name(t *testing.T)
│   ├── func TestAnthropicProvider_DefaultModel(t *testing.T)
│   ├── func TestAnthropicProvider_Chat_Success(t *testing.T)
│   ├── func TestAnthropicProvider_Chat_APIError(t *testing.T)
│   ├── func TestAnthropicProvider_Chat_EmptyResponse(t *testing.T)
│   ├── func TestAnthropicProvider_ChatStream_Success(t *testing.T)
│   ├── func TestAnthropicProvider_ChatStream_APIError(t *testing.T)
│   ├── func TestAnthropicProvider_ChatStream_ContextCancellation(t *testing.T)
│   └── func (*testTransport) RoundTrip(req *http.Request) (*http.Response, error)
├── gemini.go
│   ├── type geminiProvider {apiKey: string, model: string, client: *http.Client}
│   ├── type geminiPart {Text: string}
│   ├── type geminiContent {Role: string, Parts: []geminiPart}
│   ├── type geminiRequest {SystemInstruction: *geminiContent, Contents: []geminiContent}
│   ├── type geminiCandidate {Content: geminiContent}
│   ├── type geminiResponse {Candidates: []geminiCandidate}
│   ├── func newGeminiProvider(apiKey, model string) *geminiProvider
│   ├── func (*geminiProvider) Name() string
│   ├── func (*geminiProvider) Chat(ctx context.Context, systemPrompt string, messages []Message) (string, error)
│   └── func (*geminiProvider) ChatStream(ctx context.Context, systemPrompt string, messages []Message) (<-chan StreamChunk, error)
├── llm.go
│   ├── type Message {Role: string, Content: string}
│   ├── type StreamChunk {Content: string, Done: bool, Error: error}
│   ├── type Provider {Chat: (ctx context.Context, systemPrompt string, messages []Message) (string, error), ChatStream: (ctx context.Context, systemPrompt string, messages []Message) (<-chan StreamChunk, error), Name: () string}
│   ├── type Config {APIKey: string, Model: string}
│   └── func NewProvider(adapter, apiKey, model string) (Provider, error)
├── llm_test.go
│   ├── func TestMockProvider_Name(t *testing.T)
│   ├── func TestMockProvider_Chat(t *testing.T)
│   ├── func TestMockProvider_Chat_Error(t *testing.T)
│   ├── func TestMockProvider_ChatStream(t *testing.T)
│   ├── func TestMockProvider_ChatStream_Error(t *testing.T)
│   ├── func TestMockProvider_ChatStream_ContextCancellation(t *testing.T)
│   ├── func TestMockProvider_Reset(t *testing.T)
│   ├── func TestMessage_Structure(t *testing.T)
│   └── func TestStreamChunk_Structure(t *testing.T)
└── mock.go
    ├── type MockProvider {name: string, ChatResponse: string, ChatError: error, StreamChunks: []string, StreamError: error, ChatCalls: []MockChatCall, StreamCalls: []MockStreamCall}
    ├── type MockChatCall {SystemPrompt: string, Messages: []Message}
    ├── type MockStreamCall {SystemPrompt: string, Messages: []Message}
    ├── func NewMockProvider() *MockProvider
    ├── func (*MockProvider) Name() string
    ├── func (*MockProvider) Chat(ctx context.Context, systemPrompt string, messages []Message) (string, error)
    ├── func (*MockProvider) ChatStream(ctx context.Context, systemPrompt string, messages []Message) (<-chan StreamChunk, error)
    └── func (*MockProvider) Reset()
```
