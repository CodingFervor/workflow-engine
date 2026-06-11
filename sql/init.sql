-- Workflow Engine - Database Schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenants (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    plan VARCHAR(20) DEFAULT 'free',
    max_workflows INT DEFAULT 100,
    max_executions_per_day BIGINT DEFAULT 10000,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    username VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(100), email VARCHAR(100), phone VARCHAR(30),
    role VARCHAR(20) DEFAULT 'operator',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, username)
);

-- 工作流定义
CREATE TABLE workflows (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    trigger_type VARCHAR(20) DEFAULT 'manual' CHECK (trigger_type IN ('manual','schedule','event','webhook','api')),
    trigger_config JSONB,
    variables JSONB,
    timeout_seconds INT DEFAULT 300,
    max_retries INT DEFAULT 3,
    version INT DEFAULT 1,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft','active','paused','archived')),
    created_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, name)
);

-- 工作流步骤/节点
CREATE TABLE workflow_steps (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('start','end','task','condition','parallel','join','delay','sub_process','http_call','script','email','notification','approval','transform','loop','switch')),
    config JSONB NOT NULL,
    position_x INT DEFAULT 0,
    position_y INT DEFAULT 0,
    sort_order INT DEFAULT 0,
    timeout_seconds INT DEFAULT 60,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_steps_workflow ON workflow_steps(workflow_id);

-- 步骤之间的连线/转换
CREATE TABLE workflow_transitions (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    from_step_id BIGINT NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    to_step_id BIGINT NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    condition_expr TEXT,
    label VARCHAR(200),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_transitions_workflow ON workflow_transitions(workflow_id);

-- 工作流实例 (一次执行)
CREATE TABLE workflow_instances (
    id BIGSERIAL PRIMARY KEY,
    workflow_id BIGINT NOT NULL REFERENCES workflows(id),
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    business_key VARCHAR(200),
    input JSONB,
    output JSONB,
    variables JSONB,
    status VARCHAR(20) DEFAULT 'running' CHECK (status IN ('running','completed','failed','cancelled','paused','timeout')),
    progress INT DEFAULT 0,
    error_message TEXT,
    started_by BIGINT REFERENCES users(id),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP,
    duration_ms BIGINT DEFAULT 0
);
CREATE INDEX idx_instances_workflow ON workflow_instances(workflow_id);
CREATE INDEX idx_instances_tenant ON workflow_instances(tenant_id);
CREATE INDEX idx_instances_status ON workflow_instances(status);
CREATE INDEX idx_instances_time ON workflow_instances(started_at);

-- 步骤实例 (每一步的执行记录)
CREATE TABLE step_instances (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_id BIGINT NOT NULL REFERENCES workflow_steps(id),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending','running','completed','failed','skipped','cancelled','timeout')),
    input JSONB,
    output JSONB,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    duration_ms BIGINT DEFAULT 0,
    executor VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_step_inst_instance ON step_instances(instance_id);
CREATE INDEX idx_step_inst_step ON step_instances(step_id);

-- 审批任务
CREATE TABLE approval_tasks (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_id BIGINT NOT NULL REFERENCES workflow_steps(id),
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    assignee_id BIGINT REFERENCES users(id),
    assignee_type VARCHAR(20) DEFAULT 'user' CHECK (assignee_type IN ('user','role','group','any')),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','transferred','expired')),
    result VARCHAR(20),
    comment TEXT,
    due_at TIMESTAMP,
    resolved_at TIMESTAMP,
    resolved_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_approval_assignee ON approval_tasks(assignee_id);
CREATE INDEX idx_approval_status ON approval_tasks(status);

-- 执行日志
CREATE TABLE execution_logs (
    id BIGSERIAL PRIMARY KEY,
    instance_id BIGINT REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_instance_id BIGINT REFERENCES step_instances(id),
    level VARCHAR(10) NOT NULL CHECK (level IN ('debug','info','warn','error')),
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_exec_logs_instance ON execution_logs(instance_id);
CREATE INDEX idx_exec_logs_time ON execution_logs(created_at);

-- Webhook端点
CREATE TABLE webhooks (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    workflow_id BIGINT NOT NULL REFERENCES workflows(id),
    path VARCHAR(500) NOT NULL,
    method VARCHAR(10) DEFAULT 'POST',
    secret VARCHAR(200),
    enabled BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMP,
    trigger_count BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, path)
);

-- 调度统计 (按小时)
CREATE TABLE workflow_metrics (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL REFERENCES tenants(id),
    workflow_id BIGINT REFERENCES workflows(id),
    hour TIMESTAMP NOT NULL,
    executions BIGINT DEFAULT 0,
    completed BIGINT DEFAULT 0,
    failed BIGINT DEFAULT 0,
    cancelled BIGINT DEFAULT 0,
    avg_duration_ms DECIMAL(10,2),
    max_duration_ms DECIMAL(10,2),
    UNIQUE(tenant_id, workflow_id, hour)
);
CREATE INDEX idx_wf_metrics_hour ON workflow_metrics(hour);
