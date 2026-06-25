# EffiHub 效率舱

> 高效导航，触手可及

一个现代化的网址导航站，汇集开发工具、设计资源、AI 工具、效率工具等优质网站，帮助用户提升工作效率。带有后台管理系统，支持分类和链接的增删改查。

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat&logo=mysql)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)

## ✨ 功能特性

- 📂 **分类与链接管理** - 支持完整的增删改查操作
- 🔍 **网站信息自动抓取** - 添加链接时自动获取 favicon、标题和描述
- 🎨 **主题切换** - 支持 Auto / Light / Dark 三种模式
- 🔎 **实时搜索** - 支持按名称和描述搜索，Cmd+K 快捷键聚焦
- 📱 **响应式设计** - 适配桌面和移动设备
- 🐳 **Docker 支持** - 提供 Dockerfile 和 docker-compose.yml

## 📸 截图

| 前台导航页 | 后台管理页 |
|:---:|:---:|
| ![首页](https://images.lizhixu.cn/i/2025/12/03/12y4do1.png) | ![后台](https://images.lizhixu.cn/i/2025/12/03/12tlo5e.png) |

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+

### 1. 克隆项目

```bash
git clone https://github.com/your-username/effihub.git
cd effihub
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置数据库连接和管理密码：

```env
# 数据库配置
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=effihub

# 管理后台密码
ADMIN_PASSWORD=your_admin_password

# 图片上传配置（可选）
IMAGE_UPLOAD_API=https://your-img-api
IMAGE_UPLOAD_TOKEN=your_token
```

### 3. 运行

```bash
# 直接运行
go run .

# 或编译后运行
go build -o effihub .
./effihub
```

访问 http://localhost:8080 查看导航页，http://localhost:8080/admin.html 进入管理后台。

## 🐳 Docker 部署

```bash
docker-compose up -d
```

或手动构建：

```bash
docker build -t effihub .
docker run -d -p 8080:8080 --env-file .env effihub
```

## 🛠️ 开发

### 热重载

使用 [Air](https://github.com/air-verse/air) 实现热重载：

```bash
go install github.com/air-verse/air@latest
air
```

### 构建脚本

```bash
# Linux/macOS
./build.sh

# Windows
build.bat

# 构建并推送 Docker 镜像
./build.sh --push v1.0.0

# 多架构构建（amd64/arm64）
./build.sh --push-multi v1.0.0
```

## 📁 项目结构

```
effihub/
├── main.go              # 入口文件，路由注册，CORS 中间件
├── config/
│   ├── config.go        # 应用配置
│   ├── database.go      # 数据库初始化
│   └── ca.pem           # MySQL TLS 证书
├── handlers/
│   ├── auth.go          # 认证相关接口
│   ├── handlers.go      # 分类和链接 CRUD
│   └── favicon.go       # 网站信息抓取
├── models/
│   └── models.go        # 数据模型 + 自动建表
├── static/
│   ├── index.html       # 前台导航页
│   ├── admin.html       # 后台管理页
│   ├── script.js        # 前台交互逻辑
│   └── style.css        # 前台样式
├── Dockerfile
├── docker-compose.yml
└── .env.example         # 环境变量模板
```

## 📡 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/health` | 健康检查 |
| `POST` | `/api/auth/login` | 管理员登录 |
| `GET` | `/api/categories` | 获取所有分类 |
| `POST` | `/api/categories` | 创建分类 |
| `PUT` | `/api/categories/{id}` | 更新分类 |
| `DELETE` | `/api/categories/{id}` | 删除分类（级联删除链接） |
| `GET` | `/api/links` | 获取所有链接 |
| `POST` | `/api/links` | 创建链接 |
| `PUT` | `/api/links/{id}` | 更新链接 |
| `DELETE` | `/api/links/{id}` | 删除链接 |
| `GET` | `/api/favicon?url=...` | 抓取网站信息 |

## ⚙️ 配置项

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `DB_HOST` | MySQL 主机 | `127.0.0.1` |
| `DB_PORT` | MySQL 端口 | `3306` |
| `DB_USER` | 数据库用户 | `root` |
| `DB_PASSWORD` | 数据库密码 | - |
| `DB_NAME` | 数据库名 | `effihub` |
| `ADMIN_PASSWORD` | 管理后台密码 | - |
| `IMAGE_UPLOAD_API` | 图片上传 API | - |
| `IMAGE_UPLOAD_TOKEN` | 上传 Token | - |

## 🧱 技术栈

**后端：** Go 1.21 + 标准库 net/http + MySQL

**前端：** 原生 HTML/CSS/JavaScript，无框架依赖

**依赖：** 仅两个第三方包
- `github.com/go-sql-driver/mysql` - MySQL 驱动
- `github.com/joho/godotenv` - 环境变量加载

## 📄 License

MIT License

## 🙏 致谢

- [Go](https://golang.org/)
- [MySQL](https://www.mysql.com/)
- [Docker](https://www.docker.com/)
