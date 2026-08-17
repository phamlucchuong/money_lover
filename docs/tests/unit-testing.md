# Hướng dẫn viết Unit Test trong Go

Tutorial về unit test trong Go, dùng các thư viện phổ biến: `testify`, `mockery`. Pattern chính là **table-driven test** — một hàm test với nhiều subtest.

## Mục lục

1. [Convention đặt tên](#1-convention-đặt-tên)
2. [Table-driven test pattern](#2-table-driven-test-pattern)
3. [testify/assert vs require](#3-testifyassert-vs-require)
4. [testify/mock — Argument matching](#4-testifymock--argument-matching)
5. [testify/mock — Return values](#5-testifymock--return-values)
6. [testify/mock — Verification](#6-testifymock--verification)
7. [Mockery — Auto-generate mock](#7-mockery--auto-generate-mock)
8. [t.Parallel()](#8-tparallel)
9. [Lỗi thường gặp](#9-lỗi-thường-gặp)
10. [DO / DON'T](#10-do--dont)
11. [Commands](#11-commands)

---

## 1. Convention đặt tên

```go
func TestService_Create(t *testing.T) {     // function test
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) { // subtest
            // ...
        })
    }
}
```

**Pattern:**
- `Test{Struct}_{Method}` — function test
- `t.Run(name, ...)` — subtest với `name` mô tả scenario

**Tên subtest mô tả case, không phải số:**

```go
// Tốt
t.Run("success", ...)
t.Run("email already exists", ...)
t.Run("repo error on get", ...)

// Xấu
t.Run("case_1", ...)
t.Run("test 2", ...)
```

---

## 2. Table-driven test pattern

Cấu trúc chuẩn cho unit test trong Go: **một struct slice chứa các test case**, mỗi case tự setup mock và expect.

### Cấu trúc cơ bản

```go
func TestService_Create(t *testing.T) {
    testCases := []struct {
        name          string
        input         Input
        mockSetup     func(repo *MockRepository)
        expectedOutput Output
        expectedError error
    }{
        { name: "success", ... },
        { name: "error case", ... },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            tc.mockSetup(mockRepo)

            svc := NewService(mockRepo)

            output, err := svc.Create(ctx, tc.input)

            // assertion
        })
    }
}
```

### Ví dụ đầy đủ

```go
func TestService_Create(t *testing.T) {
    ctx := context.Background()
    mockUserID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

    validReq := CreateUserRequest{
        Email:    "test@example.com",
        Name:     "Test User",
        Password: "password123",
    }

    testCases := []struct {
        name          string
        req           CreateUserRequest
        mockSetup     func(repo *MockRepository)
        expectedResp  *UserResponse
        expectedError error
    }{
        {
            name: "success",
            req:  validReq,
            mockSetup: func(repo *MockRepository) {
                repo.EXPECT().GetByEmailAndDeletedAtIsNull(mock.Anything, validReq.Email).
                    Return(nil, gorm.ErrRecordNotFound)

                repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *User) bool {
                    return u.Email == validReq.Email &&
                        u.Name == validReq.Name &&
                        u.Password == validReq.Password
                })).
                    RunAndReturn(func(ctx context.Context, u *User) error {
                        u.ID = mockUserID
                        return nil
                    })
            },
            expectedResp: &UserResponse{
                ID:    mockUserID,
                Name:  validReq.Name,
                Email: validReq.Email,
            },
            expectedError: nil,
        },
        {
            name: "email already exists",
            req:  validReq,
            mockSetup: func(repo *MockRepository) {
                repo.EXPECT().GetByEmailAndDeletedAtIsNull(mock.Anything, validReq.Email).
                    Return(&User{Email: validReq.Email}, nil)
            },
            expectedResp:  nil,
            expectedError: ErrUserAlreadyExists,
        },
        {
            name: "repo error on get",
            req:  validReq,
            mockSetup: func(repo *MockRepository) {
                repo.EXPECT().GetByEmailAndDeletedAtIsNull(mock.Anything, validReq.Email).
                    Return(nil, errors.New("connection refused"))
            },
            expectedResp:  nil,
            expectedError: errors.New("check existing user: connection refused"),
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            tc.mockSetup(mockRepo)

            svc := NewService(mockRepo)

            resp, err := svc.Create(ctx, tc.req)

            if tc.expectedError != nil {
                assert.Error(t, err)
                assert.Equal(t, tc.expectedError.Error(), err.Error())
                assert.Nil(t, resp)
            } else {
                require.NoError(t, err)
                require.NotNil(t, resp)
                assert.Equal(t, tc.expectedResp.ID, resp.ID)
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Lợi ích

- **DRY**: setup logic dùng chung, chỉ khác input/expectation
- **Dễ đọc**: mỗi case là 1 block self-contained
- **Dễ mở rộng**: thêm case = thêm 1 entry, không sửa logic
- **Subtest isolation**: 1 case fail không ảnh hưởng case khác

---

## 3. testify/assert vs require

### `assert` — non-fatal

Test vẫn chạy tiếp nếu fail. Dùng để xem hết test fail ở đâu.

```go
assert.Equal(t, expected, actual)
assert.NotNil(t, obj)
assert.Nil(t, obj)
assert.True(t, cond)
assert.Error(t, err)
assert.NoError(t, err)
```

### `require` — fatal

Test **dừng ngay** nếu fail. Dùng cho precondition.

```go
require.NoError(t, err)    // fail → dừng test
require.NotNil(t, resp)   // fail → dừng test
```

### Quy tắc

| Tình huống | Dùng |
|---|---|
| Precondition (nếu fail → các check sau vô nghĩa) | `require` |
| Verification (muốn xem hết) | `assert` |
| Setup lỗi | `require` |
| Assert giá trị cuối | `assert` |

```go
resp, err := svc.Create(ctx, req)
require.NoError(t, err)        // fail fast
require.NotNil(t, resp)        // nil → các check sau vô nghĩa
assert.Equal(t, exp.ID, resp.ID)
assert.Equal(t, exp.Email, resp.Email)
```

---

## 4. testify/mock — Argument matching

### `mock.Anything`

Match bất kỳ giá trị nào. Dùng cho `ctx` (mỗi lần gọi context khác nhau):

```go
repo.EXPECT().GetByEmail(mock.Anything, "test@example.com")
                        ↑ ctx        ↑ email phải match chính xác
```

### Giá trị cụ thể

Match chính xác:

```go
repo.EXPECT().GetByEmail(mock.Anything, "test@example.com")
```

### `mock.AnythingOfType`

Match kiểu:

```go
repo.EXPECT().GetByUserID(mock.Anything, mock.AnythingOfType("uuid.UUID"))
```

### `mock.MatchedBy`

Match bằng predicate. Dùng để verify argument fields:

```go
repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *User) bool {
    return u.Email == "test@example.com" && u.Name == "Test"
}))
```

### Lỗi thường gặp

Khi arg mismatch:

```
FAIL: GetByEmail(ctx, "test@example.com")
Should be called. Arguments do not match:
  0: PASS: ctx, "test@example.com"
  1: FAIL:  ctx, "different@example.com"
```

**Tip:** arg trong test phải **khớp** với arg service gọi. Dùng `mock.Anything` cho field không quan tâm.

---

## 5. testify/mock — Return values

### `.Return(...)`

Định nghĩa giá trị mock trả về. Số lượng arg khớp method signature:

```go
// Method return (*User, error)
repo.EXPECT().GetByEmail(...).Return(&User{}, nil)

// Method return error
repo.EXPECT().Create(...).Return(nil)

// Method return ([]*User, error)
repo.EXPECT().GetAll(...).Return([]*User{}, nil)
```

### `.RunAndReturn(func)`

Khi cần **mutate input** (vd set ID) hoặc **side effect**:

```go
repo.EXPECT().Create(mock.Anything, mock.Anything).
    RunAndReturn(func(ctx context.Context, u *User) error {
        u.ID = uuid.New()  // mutate input, giả lập DB set ID
        return nil
    })
```

### `.Once`, `.Twice`, `.Times(n)`

Giới hạn số lần gọi. Mặc định = bất kỳ.

```go
repo.EXPECT().Create(...).Return(nil).Once()
```

---

## 6. testify/mock — Verification

### `mockRepo.AssertExpectations(t)`

Verify mọi `EXPECT()` được gọi đúng. **Luôn gọi cuối test**.

```go
mockRepo.AssertExpectations(t)
```

### Command chạy test

```bash
go test -v -run TestService_Create
```

Output khi pass:

```
=== RUN   TestService_Create
=== PAUSE TestService_Create
=== CONT  TestService_Create
=== RUN   TestService_Create/success
--- PASS: TestService_Create (0.00s)
    --- PASS: TestService_Create/success (0.00s)
```

---

## 7. Mockery — Auto-generate mock

### Khi nào cần

Khi test cần mock một interface mà bạn không muốn viết tay. Mockery tự generate mock struct implement tất cả methods của interface.

### Cài

```bash
go install github.com/vektra/mockery/v2@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Workflow

**Bước 1:** Định nghĩa interface trong code:

```go
package user

//go:generate mockery

type Repository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}
```

**Bước 2:** Config `.mockery.yaml` ở root project:

```yaml
with-expecter: true
mockname: "Mock{{.InterfaceName}}"
outpkg: "{{.PackageName}}mocks"
filename: "mock_{{.InterfaceNameLower}}.go"
dir: "{{.InterfaceDirRelative}}/mocks"

packages:
  your/module/path/internal/feature/user:
    interfaces:
      Repository:
```

**Bước 3:** Generate:

```bash
go generate ./...
```

Sẽ tạo `internal/feature/user/mocks/mock_repository.go` (~400 dòng) với:
- `MockRepository` struct
- `EXPECT()` method (do `with-expecter: true`)
- Mỗi method có `*_Call` type cho fluent API

**Bước 4:** Dùng trong test:

```go
import yourmocks "your/module/path/internal/feature/user/mocks"

func TestService(t *testing.T) {
    mockRepo := new(yourmocks.MockRepository)
    mockRepo.EXPECT().Create(...).Return(nil)
    // ...
}
```

**Lưu ý:** Import path là `.../mocks` NHƯNG package name là `usermocks` (do `outpkg: "{{.PackageName}}mocks"`). Bắt buộc dùng alias khi import.

### Sau khi refactor interface

```bash
rm internal/feature/user/mocks/mock_*.go
go generate ./...
```

### Template variables trong `.mockery.yaml`

| Variable | Giá trị |
|---|---|
| `{{.InterfaceName}}` | Tên interface (vd `Repository`) |
| `{{.InterfaceNameLower}}` | Tên viết thường (vd `repository`) |
| `{{.PackageName}}` | Tên package (vd `user`) |
| `{{.InterfaceDirRelative}}` | Path tương đối (vd `internal/feature/user`) |

---

## 8. t.Parallel()

Cho phép test chạy song song — nhanh hơn nhưng có ràng buộc.

```go
func TestService_Create(t *testing.T) {
    t.Parallel()  // parallel với test khác

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()  // subtest parallel
            // ...
        })
    }
}
```

### Khi nào dùng

✅ Test mock (mỗi test tạo mock riêng, không share state)
✅ Test không touch global state
✅ Test nhanh

### Khi nào cẩn thận

⚠️ Test dùng global DB (conflict giữa các test)
⚠️ Test modify global config
⚠️ Test có thứ tự dependency

---

## 9. Lỗi thường gặp

### Panic: mock expectation không match

```
PANIC: assert: mock: I don't know what to return
because the method call was unexpected
```

**Nguyên nhân:** Test gọi method mà mock chưa setup.

**Fix:** Thêm `EXPECT()` tương ứng trong `mockSetup`.

### FAIL: argument mismatch

```
FAIL: GetByEmail(ctx, "test@example.com")
Should be called. Arguments do not match:
  0: PASS: ctx, "test@example.com"
  1: FAIL:  ctx, "different@example.com"
```

**Nguyên nhân:** Test setup expect arg A, nhưng service gọi arg B.

**Fix:** Khớp arg, hoặc dùng `mock.Anything`.

### FAIL: expectation không được gọi

```
FAIL: GetByEmailAndDeletedAtIsNull(ctx, "test@example.com")
Should be called exactly once, but was called 0 times
```

**Nguyên nhân:** Service không gọi method mà test expect.

**Fix:** Check code service — có thể return sớm.

### Compile error: undefined mock

```
./service_test.go:42:30: undefined: usermocks.MockRepository
```

**Nguyên nhân:** Mock chưa generate hoặc import sai.

**Fix:**
```bash
go generate ./...
```

Verify import:
```go
import usermocks "your/module/path/internal/feature/user/mocks"
```

---

## 10. DO / DON'T

### DO

1. **Đặt `require` trước, `assert` sau:**
   ```go
   require.NoError(t, err)         // fail fast
   assert.Equal(t, exp, actual)    // xem hết fail
   ```

2. **Test public API, không test private internal:**
   ```go
   package user_test  // external test
   ```

3. **Mock setup trong test case, không ở global:**
   - Mỗi test case tự setup → intent rõ ràng

4. **Cleanup với `t.Cleanup`:**
   ```go
   db := setupTestDB(t)
   t.Cleanup(func() { db.Close() })
   ```

5. **Tên case mô tả scenario:**
   ```go
   name: "email already exists"  // tốt
   ```

### DON'T

1. **Test nhiều thứ trong 1 test case** → khó debug khi fail.

2. **Mock những gì không external:**
   ```go
   // ❌ Mock function thuần pure
   // ✅ Mock DB, HTTP, time, random
   ```

3. **Sleep trong test:**
   ```go
   // ❌ time.Sleep(1 * time.Second)
   // ✅ Mock clock hoặc wait condition
   ```

4. **Skip test không có lý do:**
   ```go
   t.Skip("TODO: fix after refactor")  // OK
   t.Skip()                            // Xấu
   ```

5. **Duplicate code khi thêm test case:**
   - Thêm entry vào `testCases` slice, đừng duplicate `t.Run`

---

## 11. Commands

```bash
# Tất cả test
go test ./...

# Verbose
go test -v ./internal/feature/user/

# Filter theo tên function
go test -v -run TestService_Create ./internal/feature/user/

# Filter subtest cụ thể
go test -v -run "TestService_Create/success" ./internal/feature/user/

# Race detector
go test -race ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Chạy nhiều lần (detect flaky)
go test -count=3 ./...
```

### Pattern highlight khi debug

```bash
# Xem chi tiết test fail
go test -v -run TestService_Create

# Bypass cache, luôn chạy lại
go test -v -count=1 -run TestService_Create
```

---

## Tham khảo

- [testify](https://github.com/stretchr/testify)
- [mockery](https://vektra.github.io/mockery/)
- [Go testing](https://go.dev/doc/tutorial/add-a-test)
- [Table-driven tests (Dave Cheney)](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
