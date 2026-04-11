# Assistant 项目分析文档

> 更新时间: 2026-04-16
> 分析路径: /home/shiyi/share/github/assistant

---

## 一、项目概述

**项目名称**: Assistant
**技术栈**: Go 1.22+ / Gin / MySQL / SQLc / Swagger
**定位**: 本地文档检索 + 多层记忆的个人 AI 助理

---

## 二、核心功能

### 2.1 Tars AI 助理

个人 AI 助手，核心能力：

- **多层记忆系统**: ShortTerm (64条环形缓冲) → Memdoc (Markdown 文档) → Recall (历史召回)
- **本地 Wiki 搜索**: Grep + LLM Rerank，支持中英文混合检索
- **知识图谱**: 实体抽取 + 关系建立 + 知识页面生成
- **会话管理**: Summary / Context / PendingTasks 跟踪

### 2.2 System Prompt 指令

Tars 被配置为：

- 复杂问题先推理后回答
- 只回答被问到的问题，不主动扩展
- 不相关上下文必须忽略
- 不知道就说不知道，不编造
- 优先用 bullet lists，不用表格

### 2.3 Wiki 搜索特点

- **直接 grep 模式**: 不自动抽取 entities/concepts/summaries，只做关键词搜索
- **LLM Rerank**: 对 grep 结果打分 0-10，只用 >= 5.0 的结果
- **内容缓存**: 30s TTL，按文件 ModTime 失效
- **句子对齐**: snippet 扩展到句子边界

---

## 三、架构亮点

### 3.1 并行上下文构建

4 个数据源（Memory, Session, Wiki, Knowledge）并行获取，减少等待时间。

### 3.2 Token 限制

上下文上限 12,000 tokens，按优先级截断（Conversation 先于 System/Memory 截断）。

### 3.3 单事务写入

`IntegrateMessage` 一次事务写入 entity + relation + knowledge page，任一失败全部 rollback。

### 3.4 LLM 任务优先级

主请求绕过 llmSemaphore，后台任务（摘要/抽取/刷新）排队，容量 20。

---

## 四、待改进项

 | 类别     | 问题                     | 建议                                |
 | ------   | ------                   | ------                              |
 | 测试     | 覆盖率极低               | 增加单元测试                        |
 | Wiki     | Grep 为纯子串匹配        | 考虑 BM25 或 embedding 语义搜索     |
 | 召回     | Recall 和 Knowledge 分离 | 可考虑合并检索                      |
 | 可观测性 | 无 Prometheus metrics    | 添加关键指标                        |
 | 可观测性 | 无链路追踪               | 引入 OpenTelemetry                  |
 | 配置     | config.yaml 包含敏感信息 | 使用 config.example.yaml + 环境变量 |

---

## 五、项目结构

```
assistant/
├── cmd/assistant/                    # 入口点
├── internal/
│   ├── app/
│   │   ├── modules/
│   │   │   ├── health/
│   │   │   ├── sys_user/
│   │   │   ├── sys_server/
│   │   │   ├── tars/                 # AI 助理核心
│   │   │   │   ├── handler.go        # 消息处理中枢
│   │   │   │   ├── interfaces.go
│   │   │   │   ├── module.go
│   │   │   │   ├── memory/
│   │   │   │   │   ├── memory.go     # MemoryService
│   │   │   │   │   ├── short_term.go # 环形缓冲
│   │   │   │   ├── memdoc.go         # 记忆文档
│   │   │   │   ├── recall.go         # 历史召回
│   │   │   │   └── types.go
│   │   │   │   └── knowledge/
│   │   │   │       ├── manager.go    # 知识管理器
│   │   │   │       ├── extractor.go  # 实体抽取
│   │   │   │       ├── integrator.go # 知识集成
│   │   │   │       ├── session.go    # 会话管理
│   │   │   │       └── types.go
│   │   │   └── wiki/                 # 本地 wiki
│   │   │       ├── index.go          # GrepContent + 缓存
│   │   │       ├── reranker.go       # LLM 重排序
│   │   │       └── types.go
│   │   └── repo/                     # sqlc 数据访问层
│   └── bootstrap/psl/
├── pkg/
│   ├── channel/feishu/               # 飞书通道
│   ├── llm/                          # LLM 抽象
│   │   └── providers/                # 8+ Provider
│   ├── middleware/                   # JWT + 限流
│   ├── cache/                        # TTL 缓存
│   └── workpool/                     # 并发控制
├── db/schema/                        # 表结构
└── docs/                             # 文档
```

---

## 六、启动流程

```
InitConfig() → InitLog() → InitDB() → InitDisLocker()
→ MigrateDB() → EnsureLocalSysServerRegistered()
→ app.Run()
```

`app.Run()` 中各 Module 依次注册到 Gin Router，Channel 启动监听。

---

## 七、接口清单

### 7.1 公开接口

 | 方法   | 路径          | 说明     |
 | ------ | ------        | ------   |
 | POST   | `/auth/login` | 用户登录 |
 | GET    | `/health`     | 健康检查 |

### 7.2 认证接口

 | 模块       | 方法   | 路径                                  | 说明       |
 | ------     | ------ | ------                                | ------     |
 | sys_user   | *      | `/api/v1/sys_user/*`                  | 用户管理   |
 | sys_server | *      | `/api/v1/sys_server/*`                | 服务器管理 |
 | feishu     | POST   | `/api/v1/feishu/send`                 | 发送消息   |
 | chat       | GET    | `/api/v1/chat/memory-doc/:session_id` | 记忆文档   |
 | chat       | GET    | `/api/v1/chat/messages/:session_id`   | 消息列表   |
 | chat       | GET    | `/api/v1/chat/session/:session_id`    | 会话状态   |
 | chat       | GET    | `/api/v1/chat/entities/:session_id`   | 知识实体   |
 | chat       | GET    | `/api/v1/chat/knowledge/:session_id`  | 知识页面   |
 | chat       | GET    | `/api/v1/chat/recalls/:session_id`    | 召回记录   |

