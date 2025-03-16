# FinWise - 智能财务管理系统 API

FinWise（原WalletWise）是一个功能完善的移动记账应用后端API系统，基于Go语言和Beego框架开发，为用户提供全面的个人财务管理解决方案。


## 🌟 功能特点

### 📊 用户管理
- 安全的用户注册与登录系统，支持JWT令牌认证
- 个人资料管理与定制
- 多级密码保护与重置机制

### 💰 账单记录
- 高效便捷的收支记录功能
- 多维度筛选：按日期、类别、金额范围
- 详细的账单描述与分类关联
- 批量导入导出功能

### 📝 预算管理
- 创建总体月度预算
- 细分分类预算限额设置
- 实时预算使用进度跟踪
- 自定义预算告警阈值

### 🏷️ 分类管理
- 灵活的收支分类定制
- 分类图标与色彩个性化
- 智能分类建议系统

### 📈 数据分析
- 丰富的图表可视化统计
- 收支趋势分析
- 消费习惯评估
- 分类占比透视
- 月度/季度/年度财务报告

## 🛠️ 技术栈

- **后端框架**: Go + Beego
- **数据存储**: MySQL
- **认证机制**: JWT
- **API设计**: RESTful
- **文档工具**: Swagger
- **部署方案**: Docker + Docker Compose

## 📊 系统架构

```
┌───────────────┐     ┌────────────────┐     ┌──────────────┐
│  移动端应用    │────▶│   API 网关      │────▶│  认证中间件   │
└───────────────┘     └────────────────┘     └──────────────┘
                                │
                                ▼
┌───────────┐     ┌────────────────────────────┐     ┌───────────┐
│  数据库    │◀───▶│   业务逻辑层 (Controllers)  │────▶│  缓存系统  │
└───────────┘     └────────────────────────────┘     └───────────┘
                                │
                                ▼
                      ┌─────────────────────┐
                      │  数据模型 (Models)   │
                      └─────────────────────┘
```

## 🚀 快速开始

### 前置条件

- Go 1.17+
- MySQL 5.7+
- Docker (可选)

### 安装部署

1. **克隆仓库**

```bash
git clone https://github.com/agions/finwise.git
cd finwise
```

2. **安装依赖**

```bash
go mod tidy
```

3. **创建配置**

创建MySQL数据库:

```sql
CREATE DATABASE finwise CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

修改`conf/app.conf`配置:

```ini
dbuser = your_username
dbpassword = your_password
dbhost = localhost
dbport = 3306
dbname = finwise
```

4. **启动服务**

```bash
go run main.go
```

或使用Beego工具:

```bash
bee run
```

### 🐳 Docker部署

1. 构建镜像

```bash
docker build -t finwise:latest .
```

2. 使用Docker Compose启动

```bash
docker-compose up -d
```

## 📚 API文档

项目启动后，访问以下地址查看Swagger API文档:

```
http://localhost:8080/swagger/
```

### API示例

#### 用户注册

```
POST /api/user/register

请求体:
{
  "username": "user123",
  "email": "user@example.com",
  "password": "secure_password",
  "phone": "13800138000"
}

响应:
{
  "code": 200,
  "message": "成功",
  "data": {
    "user": {
      "id": 1,
      "username": "user123",
      "email": "user@example.com",
      "phone": "13800138000"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 创建账单

```
POST /api/bills

请求头:
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

请求体:
{
  "category_id": 1,
  "amount": 99.50,
  "type": "expense",
  "date": "2023-03-15",
  "description": "超市购物"
}

响应:
{
  "code": 200,
  "message": "成功",
  "data": {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "amount": 99.50,
    "type": "expense",
    "date": "2023-03-15T00:00:00Z",
    "description": "超市购物",
    "category_name": "购物",
    "category_icon": "shopping"
  }
}
```

## 🔒 安全特性

- JWT令牌身份验证
- 密码加密存储(bcrypt)
- API请求限流保护
- SQL注入防护
- 跨域资源共享(CORS)配置

## 📋 开发路线图

- [ ] 多语言支持
- [ ] 社交账号登录集成
- [ ] AI智能消费分析
- [ ] 债务跟踪管理
- [ ] 定期报告邮件推送
- [ ] WebSocket实时通知

## 🤝 贡献指南

1. Fork项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

## 📄 许可证

本项目采用MIT许可证 - 详情见[LICENSE](LICENSE)文件

## 📞 联系方式

- 项目维护者: [Agions](https://github.com/agions)
- 项目仓库: [https://github.com/agions/finwise](https://github.com/agions/finwise)
- 问题反馈: [提交Issue](https://github.com/agions/finwise/issues)

---

**FinWise** - 让财务管理变得简单而智能 💡💰