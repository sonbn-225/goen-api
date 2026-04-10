# Backend Architecture & Coding Rules (Golang)

Tài liệu này định nghĩa các quy tắc cốt lõi khi phát triển Backend sử dụng stack: **Golang, Gin, PostgreSQL, pgx, sqlc, goose, JWT và Viper.**

Mục tiêu: Đảm bảo code base sạch, an toàn, hiệu năng cao và chuẩn bị tốt cho các luồng xử lý dữ liệu nhạy cảm (như tài chính cá nhân).

---

## 1. Cấu trúc Dự án (Clean Architecture / Standard Layout)

Tuân thủ nghiêm ngặt việc chia tách logic thành các lớp riêng biệt để dễ dàng viết test và mở rộng.

* **`/cmd/api`**: Chứa file `main.go`. Nơi khởi tạo Viper, kết nối Database, khởi chạy Gin server.
* **`/internal/handlers` (Giao tiếp)**: Chứa các HTTP handlers của Gin. Chỉ làm nhiệm vụ nhận request (bind JSON), gọi xuống lớp Service, và trả về HTTP response. KHÔNG viết logic nghiệp vụ ở đây.
* **`/internal/services` (Nghiệp vụ)**: Trái tim của ứng dụng. Chứa logic tính toán (ví dụ: đối soát số dư, kiểm tra hạn mức). Nhận dữ liệu từ Handler và gọi xuống Repository.
* **`/internal/repository` (Lưu trữ)**: Nơi gọi các hàm tương tác với Database do `sqlc` sinh ra.
* **`/pkg` (Tiện ích chung)**: Chứa các module dùng chung độc lập với nghiệp vụ (hàm hash mật khẩu, hàm tạo/xác thực JWT).
* **`/db/query` & `/db/migration`**: Chứa các file SQL thuần cho sqlc và goose.

---

## 2. Quản lý Database (PostgreSQL, pgx, sqlc & goose)

* **Quy tắc 1 (Migration)**: Mọi thay đổi về cấu trúc bảng (schema) BẮT BUỘC phải dùng `goose`. Mỗi lần thay đổi phải tạo file migration mới (`goose create ... sql`) gồm đầy đủ hai phần `+goose Up` và `+goose Down`. Tuyệt đối không can thiệp thủ công vào Database.
* **Quy tắc 2 (Query)**: Không dùng ORM. Mọi truy vấn phải viết bằng raw SQL trong thư mục `/db/query` và chạy lệnh `sqlc generate` để sinh ra code Go type-safe.
* **Quy tắc 3 (Transaction)**: Các thao tác thay đổi dữ liệu liên hoàn (đặc biệt là giao dịch liên quan đến tiền tệ, tài sản) PHẢI được bọc trong Database Transaction của `pgx` (`Begin`, `Commit`, `Rollback`) để đảm bảo tính ACID.

---

## 3. Định tuyến và HTTP (Gin Framework)

* **Quy tắc 1 (Data Binding)**: Sử dụng `c.ShouldBindJSON()` của Gin thay vì tự parse JSON thủ công. Kết hợp với struct tags (ví dụ: `binding:"required,min=1"`) để validate dữ liệu đầu vào ngay tại lớp handler.
* **Quy tắc 2 (Routing)**: Gom nhóm các API (Route Grouping) có cùng tiền tố hoặc cùng cơ chế xác thực.
    ```go
    v1 := router.Group("/api/v1")
    {
        auth := v1.Group("/auth")
        // ... routes không cần đăng nhập
        
        users := v1.Group("/users").Use(RequireAuthMiddleware())
        // ... routes cần đăng nhập
    }
    ```
* **Quy tắc 3 (Response Format)**: Thống nhất một cấu trúc JSON trả về duy nhất cho toàn bộ hệ thống.
    ```json
    {
      "success": true,
      "data": { ... },
      "message": "Thông báo nếu có",
      "error_code": null
    }
    ```

---

## 4. Xác thực và Bảo mật (JWT & Bcrypt)

* **Quy tắc 1 (Mật khẩu)**: Mật khẩu người dùng phải được hash bằng thuật toán `bcrypt` trước khi lưu vào Database. Không có ngoại lệ.
* **Quy tắc 2 (JWT)**: Sử dụng JWT cho xác thực stateless. 
    * Tách biệt làm hai loại token: **Access Token** (tuổi thọ ngắn, ví dụ 15 phút) và **Refresh Token** (tuổi thọ dài, ví dụ 7 ngày, lưu trong DB để có thể revoke khi cần thiết).
    * Tuyệt đối không nhét dữ liệu nhạy cảm (mật khẩu, số dư) vào payload của JWT. Chỉ lưu `user_id` và `role`.
* **Quy tắc 3 (Middleware)**: Mọi API cần bảo vệ phải đi qua một Gin Middleware chuyên biệt để parse và validate JWT trong header `Authorization: Bearer <token>`.

---

## 5. Quản lý Cấu hình (Viper)

* **Quy tắc 1**: Không hardcode (gắn chết) bất kỳ thông tin nào có thể thay đổi giữa các môi trường (Dev, Staging, Prod).
* **Quy tắc 2**: Sử dụng Viper để load file `app.env` hoặc đọc trực tiếp từ Environment Variables. Các giá trị bắt buộc phải có thông qua cấu trúc struct:
    * `DB_SOURCE` (Chuỗi kết nối DB)
    * `SERVER_ADDRESS` (Cổng chạy app, ví dụ `:8080`)
    * `JWT_SECRET` (Khóa bí mật siêu mạnh để ký token)