# Workflow Engine

## English | [中文](#中文)

A BPMN-based workflow engine built with Go, Gin, PostgreSQL, and Redis.

### Features
- **BPMN 2.0** process model (Start/End/UserTask/ServiceTask/Gateway)
- **Process designer** visual node editor
- **Process deployment and versioning** versioned deployment artifacts
- **Process instance lifecycle** start/pause/resume/terminate
- **User task assignment** assignee, candidate users, candidate groups
- **Multi-instance tasks** sequential and parallel
- **Gateways** exclusive/parallel/inclusive
- **Timers and boundary events** with conditional flows
- **Sub-processes and call activities** hierarchical processes
- **Multi-tenant** data isolation by tenant

### Tech Stack
- Go 1.22 + Gin + PostgreSQL 16 + Redis 7
- Multi-tenant, RBAC, JWT auth
- Docker Compose

### Quick Start
```bash
docker-compose up -d
go run cmd/api/main.go
```

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | /api/v1/auth/login | Login |
| GET | /api/v1/process-definitions | List process definitions |
| POST | /api/v1/process-definitions | Deploy process |
| POST | /api/v1/process-instances | Start process instance |
| GET | /api/v1/process-instances | List instances |
| GET | /api/v1/tasks | List tasks |
| PUT | /api/v1/tasks/:id/complete | Complete task |
| PUT | /api/v1/tasks/:id/claim | Claim task |
| GET | /api/v1/dashboard | Dashboard stats |

---

<a id="中文"></a>
# 工作流引擎

基于 Go + Gin + PostgreSQL + Redis 构建的 BPMN 工作流引擎。

### 功能特性
- **BPMN 2.0** 流程模型（开始/结束/用户任务/服务任务/网关）
- **流程设计器** 可视化节点编辑
- **流程部署和版本管理** 版本化部署产物
- **流程实例生命周期** 启动/暂停/恢复/终止
- **用户任务分配** 受让人/候选用户/候选组
- **多实例任务** 串行和并行
- **网关** 排他/并行/包容
- **定时器和边界事件** 支持条件流
- **子流程和调用活动** 层级化流程
- **多租户** 按租户数据隔离

### 技术栈
- Go 1.22 + Gin + PostgreSQL 16 + Redis 7
- 多租户、RBAC、JWT 认证
- Docker Compose

### 快速开始
```bash
docker-compose up -d
go run cmd/api/main.go
```
