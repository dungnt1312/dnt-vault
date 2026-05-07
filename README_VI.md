# dnt-vault

Công cụ đồng bộ SSH config, private key và biến môi trường (environment variables) tự host, viết bằng Go. Đồng bộ `~/.ssh/config`, private keys và app secrets giữa các máy thông qua vault server riêng — mã hóa phía client, không phụ thuộc dịch vụ bên thứ ba.

## Cài đặt

Linux / macOS:

```
curl -fsSL https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.sh | sudo bash
```

Windows (PowerShell):

```
irm https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.ps1 -OutFile "$env:TEMP\install.ps1"; & "$env:TEMP\install.ps1"
```

Windows (Bash / Git Bash):

```
curl -fsSL https://raw.githubusercontent.com/dungnt1312/dnt-vault/master/install.sh | bash
```

Sau đó reload PATH:

```
source ~/.bashrc
```

Hoặc tải binary trực tiếp từ [Releases](https://github.com/dungnt1312/dnt-vault/releases).

## Bắt đầu nhanh

### 1. Khởi động Server

```
dnt-vault-server
```

Mặc định chạy trên `0.0.0.0:8443`. Tài khoản mặc định: `admin` / `admin`.

```
PORT=8443
DATA_PATH=~/dnt-vault/data
CONFIG_PATH=~/dnt-vault/config
```

### 2. Khởi tạo Client

```
dnt-vault init
```

Nhập URL server và đặt master password. Config lưu tại `~/.dnt-vault/config.yaml`.

Lệnh này tạo `~/.dnt-vault/ssh-master.key` cho mã hóa SSH. Biến môi trường sử dụng key riêng, khởi tạo qua `dnt-vault env init`.

### 3. Đăng nhập

```
dnt-vault login
```

### 4. Push SSH Config

```
dnt-vault push
```

Push kèm private keys:

```
dnt-vault push --include-keys
```

### 5. Pull trên máy khác

```
dnt-vault init    # cùng server URL + master password
dnt-vault login
dnt-vault pull
```

## Các lệnh CLI

```
dnt-vault init              # Khởi tạo client
dnt-vault login             # Đăng nhập vault
dnt-vault push              # Push SSH config
dnt-vault pull              # Pull SSH config
dnt-vault profile list      # Liệt kê profiles
dnt-vault profile use <name> # Pull và áp dụng profile
dnt-vault list              # Liệt kê profiles (deprecated)
dnt-vault delete <name>     # Xóa profile
dnt-vault upgrade           # Nâng cấp phiên bản mới
dnt-vault version           # Hiển thị thông tin phiên bản
```

## Đồng bộ biến môi trường (Environment Variables)

Khởi tạo mã hóa env:

```bash
dnt-vault env init
```

Push biến môi trường:

```bash
dnt-vault env push myapp/production --file .env.production
```

Pull vào shell hiện tại:

```bash
eval $(dnt-vault env pull myapp/production)
```

Pull ra file:

```bash
dnt-vault env pull myapp/production --output .env
```

Quản lý scopes/variables:

```bash
dnt-vault env list
dnt-vault env list myapp/production
dnt-vault env get myapp/production API_KEY
dnt-vault env set myapp/production API_KEY new-value
dnt-vault env delete myapp/production API_KEY
dnt-vault env delete myapp/production --all
```

## Tính năng

- Mã hóa phía client: AES-256-GCM với PBKDF2 key derivation — server không bao giờ thấy plaintext.
- Đồng bộ private key: Tùy chọn, mã hóa bằng passphrase riêng.
- Phát hiện xung đột: So sánh LCS-based diff trước khi ghi đè config local.
- Tự động backup: Backup có timestamp trước mỗi lần pull.
- Đa profile: Nhiều profile đặt tên cho mỗi user.
- Đa người dùng: Mỗi user có storage mã hóa riêng biệt.
- Giới hạn tốc độ: 5 lần đăng nhập/phút mỗi IP.
- Tắt an toàn: Drain các request đang xử lý khi nhận SIGINT/SIGTERM.

## Cấu hình

Config client tại `~/.dnt-vault/config.yaml`:

```yaml
server:
  url: http://your-server:8443
  tls_verify: true
ssh:
  config_path: ~/.ssh/config
  keys_dir: ~/.ssh
backup:
  enabled: true
  dir: ~/.dnt-vault/backups/ssh
  max_backups: 10
env:
  backup_dir: ~/.dnt-vault/backups/env
encryption:
  ssh_master_key_file: ~/.dnt-vault/ssh-master.key
  env_master_key_file: ~/.dnt-vault/env-master.key
```

Lưu ý migration: `encryption.master_key_file` cũ vẫn được hỗ trợ và được coi là đường dẫn SSH key.

Biến môi trường server:

```
PORT=8443
DATA_PATH=~/dnt-vault/data
CONFIG_PATH=~/dnt-vault/config
```

## Chạy như systemd Service

```bash
sudo tee /etc/systemd/system/dnt-vault.service << 'EOF'
[Unit]
Description=DNT-Vault SSH Config Sync Server
After=network.target

[Service]
Type=simple
Environment="PORT=8443"
ExecStart=/usr/local/bin/dnt-vault-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now dnt-vault
```

## Kiến trúc

```
┌─────────────┐      HTTP/HTTPS       ┌──────────────────┐
│  dnt-vault  │ ────── encrypted ───► │ dnt-vault-server │
│    (CLI)    │ ◄──── data only ────  │   (REST API)     │
└─────────────┘                       └──────────────────┘
```

- `server/`: REST API vault — lưu trữ encrypted blobs, JWT auth, filesystem storage.
- `cli/`: CLI tool — mã hóa local, push/pull qua HTTP.
- `shared/`: Các type dùng chung giữa server và CLI.

## Build từ source

Yêu cầu: Go 1.22+

```bash
make build
# bin/dnt-vault
# bin/dnt-vault-server
```

## Quy trình Release

**Tag & Push** → GitHub Actions tự động build và release binaries cho tất cả platforms.

```bash
# 1. Commit tất cả thay đổi
git add -A && git commit -m "your changes"

# 2. Tag phiên bản mới (trigger CI/CD)
git tag v1.1.3
git push origin master && git push origin --tags

# 3. GitHub Actions tự upload binaries lên release page
#    Không cần build hay upload thủ công
```

**Build thủ công** (không dùng CI/CD):

```bash
make release VERSION=1.1.3
# Upload releases/* lên GitHub thủ công
```

## API

```
POST   /api/v1/auth/login           # Đăng nhập → JWT token
GET    /api/v1/profiles             # Liệt kê profiles       [auth]
GET    /api/v1/profiles/:name       # Lấy dữ liệu profile   [auth]
POST   /api/v1/profiles/:name       # Lưu profile            [auth]
DELETE /api/v1/profiles/:name       # Xóa profile            [auth]

GET    /api/v1/env/scopes                  # Liệt kê env scopes     [auth]
GET    /api/v1/env/scopes/:scope           # Lấy env scope          [auth]
POST   /api/v1/env/scopes/:scope           # Lưu env scope          [auth]
DELETE /api/v1/env/scopes/:scope           # Xóa env scope          [auth]
GET    /api/v1/env/scopes/:scope/:key      # Lấy env variable       [auth]
PUT    /api/v1/env/scopes/:scope/:key      # Đặt env variable       [auth]
DELETE /api/v1/env/scopes/:scope/:key      # Xóa env variable       [auth]
```

## Xử lý sự cố

**Server không khởi động được** — kiểm tra port: `lsof -i :8443`

**Đăng nhập thất bại** — kiểm tra URL trong `~/.dnt-vault/config.yaml`, xác nhận server đang chạy: `curl http://localhost:8443/api/v1/profiles`

**Giải mã thất bại** — master password phải trùng với password đã dùng khi `push`

## Giấy phép

MIT
