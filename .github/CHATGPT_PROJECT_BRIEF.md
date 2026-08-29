# ChatGPT Project Brief

> 本文件只保存长期稳定、仓库级的信息。当前任务、临时分支、SHA、测试状态和执行进度应保存在当前 Pull Request 正文中。

## 1. Project

- 项目名称：GoApps
- GitHub 仓库：`ychenracing/GoApps`
- 默认分支：`master`
- 系统定位：用于学习和演示 Go 语言并发、缓存与区块链概念的多个独立示例集合。
- 项目最终目标：通过可阅读的 Go 实现展示并发图片爬虫、线程安全 LRU 缓存以及多种简单区块链实验；更长期目标未在仓库文档中明确。

## 2. Purpose and Non-Goals

仓库包含三个主要示例域：使用 goroutine 的图片爬虫、受 Java LinkedHashMap 启发的线程安全 LRU 缓存，以及 HTTP、网络、工作量证明和权益证明等区块链实验。

长期非目标未在仓库文档中明确。仓库未定义生产级区块链、通用缓存服务、爬虫平台、发布包或稳定公共 API，不应自行假定这些职责。

## 3. Architecture and Module Boundaries

- `src/KongjieSpider/`：基于 `net/http`、正则表达式和 goroutine 的并发图片爬虫。
- `src/LRUCache/`：LRU 缓存实现和对应 Go 测试。
- `src/BlockChain/HttpServiceBlockChain/`：HTTP 服务形式的区块链示例。
- `src/BlockChain/NetworkBlockChain/`：网络节点形式的区块链示例。
- `src/BlockChain/ProofOfWork/`：工作量证明实验。
- `src/BlockChain/ProofOfStake/`：权益证明实验。
- `images/`：README 使用的演示截图。
- ` main/`：仓库中保留的带前导空格目录及独立 Go 示例；路径名属于现有事实，不得静默更名。

这些目录是彼此独立的学习示例，没有仓库级模块或统一运行编排。各示例源文件是自身行为 Owner；治理文件不得成为第二份可执行配置或协议定义。

## 4. Non-Negotiable Constraints

- 各示例相互独立；修改一个示例时不得假定其他示例共享状态、配置或 API。
- LRU 的容量、访问顺序、并发安全与淘汰语义应由其实现和测试共同约束。
- 爬虫依赖外部网站结构和网络可用性；不得把外部内容或可访问性写成仓库保证。
- 区块链代码是教学示例；不得在没有证据时描述为生产级共识、加密资产或安全系统。
- 仓库没有 `go.mod` 或声明的 Go toolchain 版本，不得猜测模块路径或支持版本。
- 凭据、个人数据和抓取内容不得复制到治理文档、PR 模板或日志中。

## 5. Authoritative Sources

- 项目定位与示例索引：`README.MD`
- 工程约定：`AGENTS.md`
- 并发爬虫：`src/KongjieSpider/`
- LRU 实现与测试：`src/LRUCache/LRU.go`、`src/LRUCache/LRU_test.go`
- 区块链示例：`src/BlockChain/`
- 演示图片：`images/`
- Go 版本、模块、依赖、统一构建、发布和部署权威来源：未在仓库文档中明确

## 6. Standard Commands

- 安装与依赖解析：未在仓库中定义；仓库没有 `go.mod`。
- 构建：未在仓库中定义统一命令。
- 运行：README 说明各目录为独立 Go 示例，但未给出权威命令和环境要求。
- 测试：`src/LRUCache/LRU_test.go` 表明 LRU 示例具有 Go 测试，但仓库未记录权威执行命令或 Go 版本。
- lint、格式检查、静态分析和完整验收命令：未在仓库中定义。

## 7. Important Paths

- `src/KongjieSpider/main/kongjie.go`：并发爬虫入口。
- `src/LRUCache/LRU.go`：LRU 缓存实现。
- `src/LRUCache/LRU_test.go`：LRU 测试。
- `src/BlockChain/HttpServiceBlockChain/`：HTTP 区块链示例。
- `src/BlockChain/NetworkBlockChain/`：网络区块链示例。
- `src/BlockChain/ProofOfWork/`：工作量证明示例。
- `src/BlockChain/ProofOfStake/`：权益证明示例。
- `README.MD`：项目说明。
- `AGENTS.md`：渐进式验证约定。

## 8. CI and Acceptance Entry Points

- 仓库没有 `.github/workflows/`，未定义 GitHub Actions 构建、测试、lint 或发布门。
- 本地验证应遵循 `AGENTS.md` 的影响范围驱动原则。
- Definition of Done 的项目特有部分未在仓库中定义；变更至少应对受影响的独立示例执行可用的构建或测试，并准确记录环境和结果。

## 9. Prohibited Actions

- 不得把教学区块链描述为生产级安全或金融系统。
- 不得猜测 Go 版本、模块路径、依赖或统一运行命令。
- 不得无证据更名或合并独立示例目录，包括带前导空格的 ` main/`。
- 不得在治理文件中记录凭据、个人数据或抓取内容。
- 不得擅自改写 Git 历史或 force push。
- 不得丢弃未知或未提交工作，也不得覆盖无关改动。
- 不得把计划执行写成已验证完成。
- 不得根据旧聊天猜测当前分支、SHA、PR 或 CI 状态。

## 10. Context Loading Protocol

1. 新开发任务可以直接使用自然语言提出，不要求预先填写固定 Prompt。
2. 开始任务时先读取本文件。
3. 搜索与任务相关的开放 PR、分支和 Issue。
4. 如果存在匹配工作，从现有现场原地继续。
5. 当前动态任务状态默认维护在 Pull Request 正文。
6. 不强制普通单 PR 任务创建 Issue。
7. 优先读取目标代码、直接调用者、相关测试和直接相关配置。
8. 只有证据不足、状态冲突或影响范围扩大时才扩大读取。
9. 不默认加载完整仓库、完整聊天、完整日志或全部 GitHub Actions 历史。
10. 长对话交接使用 `conversation-continuity-guard`，但 GitHub 当前现场仍是状态权威来源。

## 11. References

- `README.MD`
- `AGENTS.md`
- `src/KongjieSpider/`
- `src/LRUCache/LRU.go`
- `src/LRUCache/LRU_test.go`
- `src/BlockChain/`
