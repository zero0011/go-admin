# Go-Admin

通用简易后台管理系统，基于 **Gin + Vue3 + Element Plus**，适用于快速开发。

## 项目结构

```
go-admin/
├── go-admin-template/       # 后端 (Go/Gin)
└── go-admin-template-vue/   # 前端 (Vue3/Vite)
```

## 技术栈

### 后端
- [Gin](https://github.com/gin-gonic/gin) - HTTP 框架
- [Gorm](https://gorm.io) - ORM 框架
- [Casbin](https://github.com/casbin/casbin) - 权限认证
- [JWT](https://github.com/dgrijalva/jwt-go) - 登录认证

### 前端
- Vue 3
- Vite
- Element Plus
- Vue Router
- Pinia

## 功能特点

- 用户管理
- 角色管理
- 权限管理（RBAC）
- 登录认证（JWT）
- Swagger 接口文档

## 快速开始

### 后端

```bash
cd go-admin-template

# 配置数据库
# 编辑 etc/go-admin-template-local.yaml

# 导入数据库
mysql -u root -p < go-admin-template.sql

# 运行
go run go-admin-template.go -f etc/go-admin-template-local.yaml
```

### 前端

```bash
cd go-admin-template-vue

# 安装依赖
npm install

# 开发环境运行
npm run dev

# 生产构建
npm run build
```

## Docker 部署

```bash
cd go-admin-template
docker-compose up -d
```

## License

MIT
