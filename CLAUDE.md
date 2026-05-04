# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

语言音乐多人共享屏幕共享平台，允许用户创建实时房间，进行**语音聊天、同步播放音乐/合唱、多人屏幕共享**。后端使用 Go 语言 Gin 框架，数据库 MySQL，即时通信基于 WebSocket。

**开发规范严格遵循《阿里巴巴 Java 开发手册》（对应 Go 语言实现及通用编程规约）**。

## Tech Stack

- Language: Go 1.21+ (Go Modules required)
- Web Framework: Gin
- Database: MySQL 8.0+
- Cache/Session: Redis (JWT blacklist, online status)
- Protocols: REST + WebSocket
- Auth: JWT (Access + Refresh Token, dual token mechanism)
- File Storage: Local/OSS (avatars, music files)
- Real-time Media: WebRTC (signaling via WebSocket)

## Common Commands

```bash
# Initialize Go module (first time only)
go mod init github.com/yourorg/livemix

# Download dependencies
go mod tidy

# Build
go build ./...

# Run application
go run main.go

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./...

# Run tests with coverage
go test -cover ./...

# Format code (mandatory before commit)
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run
```

## Directory Structure

```
server/
├── main.go
├── config/
│   └── config.go
├── internal/
│   ├── handler/     # HTTP/WS handlers (parameter binding, validation, call service)
│   ├── service/     # Business logic layer, composes repository calls
│   ├── repository/  # Data access layer (DAO)
│   ├── model/       # Data structures mapped to DB tables (PO objects)
│   ├── middleware/  # JWT, CORS, rate limiting, XSS filtering
│   ├── ws/          # WebSocket management (hub/room/user connections)
│   └── dto/         # Data transfer objects (request/response structs)
├── pkg/
│   ├── errcode/     # Unified error codes (Alibaba error code standard)
│   ├── jwt/         # JWT generation and parsing
│   ├── validator/   # Parameter validator
│   └── util/        # Utilities (encryption, masking, etc.)
├── migrations/      # SQL migration files (timestamp named)
└── go.mod
```

## Architecture Notes

### API Design

- All API prefix: `/api/v1`
- Auth: `Authorization: Bearer <JWT>`
- Response format:
  ```json
  {"code": 200, "message": "success", "data": {}}
  ```
- Error codes: 5-digit numbers with module prefix (e.g., 20000=success, 40001=param error, 40003=auth failed), managed in `pkg/errcode`

### WebSocket Events

Events use JSON format: `{"event": "event_name", "data": {}}`

Key events:
- `voice_join`/`voice_leave` - Voice status changes
- `music_sync` - Music sync (progress, track change)
- `screen_share_start`/`screen_share_stop` - Screen sharing notifications
- `chat_message` - New message
- `participant_update` - Participant list changes
- `webrtc_offer`/`webrtc_answer`/`webrtc_ice_candidate` - WebRTC signaling

Server uses Hub pattern for room connections, auth on connect, heartbeat every 15 seconds, concurrent-safe with channel or sync.Map.

### Middleware Registration Order

日志追踪 -> 限流 -> 认证 -> 授权 -> 参数校验 -> handler

## Database Conventions

- Table names: lowercase, underscore-separated, no reserved words
- Required fields: `id`, `created_at`; `updated_at` for frequently updated tables
- Status fields: `tinyint` with enum value comments
- Index naming: `pk_` (primary), `uk_` (unique), `idx_` (normal) - Alibaba standard
- No physical foreign keys - relationships enforced at application layer
- Charset: `utf8mb4`

## Development Standards (Alibaba Java Development Handbook - Go Adaptation)

### Naming

- Package names: lowercase, no underscores/camelCase, no vague names like `base`, `common`
- Struct names: CamelCase, export with capital first letter; JSON tags use lowerCamelCase
- Variable names: camelCase; special terms like ID, URL, JWT keep uppercase or capitalize first letter
- Constants: MixedCaps (Go style), enums use `iota`
- Function names: camelCase, export with capital first letter

### Code Format

- Use `gofmt` - no manual alignment
- Import groups: standard library, third-party, internal packages (separated by blank lines)
- Max line length: 120 columns
- Spaces around operators required

### Comments

- All exported functions, structs, constants must have godoc-format comments
- Complex business logic needs inline comments explaining "why" not "what"
- Comments in Chinese, code identifiers in English

### Control Flow

- Max nesting: 3 levels (if/for), use guard clauses for early return
- No complex expressions in conditions - extract to meaningful variables/functions
- `switch` must have `default` branch even if empty
- No `defer` in loops (resource leak risk) - use anonymous functions or manual close

### Error Handling

- Never ignore `error` return values - handle or explicitly comment for intentional ignore
- No `panic` for business exceptions - return errors, catch at top level
- Resource cleanup (files, locks) in `defer`, check nil or ignore error
- DB transactions must use `defer` for rollback/commit, handle every error branch

### Security

- SQL injection: Never concatenate SQL strings - use parameterized queries (`?` placeholders)
- XSS: HTML-escape user input (chat content) before storage and output
- CSRF: Validate Referer or use CSRF Token for state-changing endpoints (POST/PUT/DELETE)
- Passwords: Salted hash (bcrypt or argon2) - no plaintext or simple hash
- Sensitive data masking: No passwords/tokens in logs; mask emails in external display
- Anti-replay: Combine token expiry with one-time nonce for critical operations

### Database Operations

- All DB operations must use `context.Context` for timeout
- Single record not found: return `sql.ErrNoRows`, service layer converts to business error code
- Batch operations: IN list max 1000 items, batch if exceeded
- Pagination queries must include ORDER BY
- No `SELECT *` - explicitly list required fields

### Unit Testing

- Core service layer coverage: 80%+
- Test files: `*_test.go`, use `testify` assertions and `sqlmock` for DB mocking
- Test cases: normal, boundary, exception scenarios
- Naming: `TestUserService_Register_Success`

### Logging

- Use structured logging (logrus or zap) - no `fmt.Println`
- Levels: DEBUG/INFO/WARN/ERROR, production default INFO
- Request logs must include traceID (inject via middleware)
- Exception logs must include full stack trace (`debug.Stack()`)

### Security Release

- Config files (config.yaml) must not be in version control - use `.gitignore`, provide `config.example.yaml`
- Secrets/DB passwords must use environment variables or config center - no hardcoding
- Migration files must be synced with code, idempotent

## Development Workflow

**分支策略：** 每次开发必须基于子分支最新进度开发，禁止直接在 main 分支开发。

**提交前检查清单（必须全部通过）：**
1. 编译检查：`go build ./...` 成功无报错
2. 语法检查：`go fmt ./...` 无格式问题
3. 静态检查：`golangci-lint run` 无警告（如已安装）
4. 单元测试：`go test ./...` 全部通过
5. 代码规范：符合《阿里巴巴 Java 开发手册》Go 适配版
6. 性能检查：无明显的性能问题（N+1 查询、内存泄漏风险等）

**提交流程：**
```bash
# 1. 确保在子分支
git checkout feature/xxx

# 2. 拉取最新代码
git pull origin feature/xxx

# 3. 执行检查
go fmt ./...
go build ./...
go test ./...

# 4. 提交并推送
git add .
git commit -m "feat: 描述"
git push origin feature/xxx
```
