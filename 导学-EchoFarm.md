# EchoFarm 项目导学

> 基线：`codex/living-valley-director` 分支，EchoFarm 0.4.0 Reflective Policy。本文只描述仓库中已有实现；真实 Stardew Valley + SMAPI 编译和游戏内冒烟仍待合法游戏环境验证。

## 1. 前置知识（面试高频标注）

| 知识点 | 为何需要 | 在本项目中的位置 | 高频度 |
| --- | --- | --- | --- |
| 分层架构与端口适配 | 解释为什么 AI 不直接修改游戏状态 | `echofarm-core/internal/*`、`stardew-echo-mod/src/EchoFarm.Bridge` | 高 |
| 事件溯源与版本化状态 | 解释多日行为如何形成可追溯画像 | `internal/modeling`、`internal/memory` | 高 |
| 幂等与请求关联 | 防止重复教学、重复动作和过期响应污染状态 | `learning.Service`、`policy.Service`、`ActionSafetyGate` | 高 |
| LLM 结构化输出 | 限制模型只能生成 Trait、意图和白名单动作 | `internal/intelligence` | 高 |
| 状态机与并发控制 | 防止同一帧并发执行多个 Echo 动作 | `EchoSession`、`CoreProcessSupervisor` | 高 |
| 人机协同规划 | 避免 Echo 与玩家争抢同一作物 | `internal/coordination`、`PlayerActivityWindow` | 高 |
| 反思式策略学习 | 解释失败和玩家纠正如何跨会话改变首次决策 | `internal/experience`、`ReflectionGraph`、`CorrectionCapture` | 高 |
| SQLite 事务 | 保证示范、画像、技能和修订原子提交 | `internal/memory/sqlite.go` | 中高 |
| 游戏主线程约束 | 防止异步网络回调直接修改游戏世界 | `StardewGamePort` | 中高 |
| 跨平台 sidecar 发布 | 解释 Go 服务如何随 Mod 分发并托管 | `CoreProcessSupervisor`、`scripts/package-nexus.sh` | 中 |

## 2. 重点亮点与学习顺序（先看这个）

| 亮点标题 | 为什么重要 | 通用技术关键词 | 先看哪些文件 | 建议学习顺序 |
| --- | --- | --- | --- | ---: |
| 受约束的 AI 决策 | 模型负责语义判断，但不能任意操作游戏 | Structured Output、白名单、双重校验 | `internal/intelligence/action_graph.go` → `internal/policy/service.go` → `ActionSafetyGate.cs` | 1 |
| 持续学习与证据合并 | 避免一次异常行为覆盖长期习惯 | Event Sourcing、版本化画像、置信度、幂等 | `internal/intelligence/learning_graph.go` → `internal/modeling/merger.go` → `internal/learning/service.go` | 2 |
| 人机任务协调 | 把“AI 助手”升级为能与玩家并行行动的分身 | Intent Inference、短时窗口、Target Claim | `PlayerActivityWindow.cs` → `internal/coordination/service.go` → `internal/policy/service.go` | 3 |
| 跨进程状态一致性 | 保证 C#、HTTP、Go 和 SQLite 对同一次动作达成一致 | Correlation ID、Snapshot Version、Idempotency | `EchoSession.cs` → `EchoFarmClient.cs` → `internal/httpapi/handler.go` → `internal/memory/sqlite.go` | 4 |
| 可解释运行账本 | 让面试演示能回答“为什么这样做” | Decision Ledger、Projection、Read Model | `internal/memory/sqlite.go` → `internal/memoryview/service.go` → `EchoMemoryPresenter.cs` | 5 |
| 反思经验闭环 | 不只当场重试，而是从失败/F10 纠正中形成可复用策略 | Reflection、Bounded Memory、Confidence Calibration | `reflection_graph.go` → `experience/service.go` → `policy/service.go` → `CorrectionCapture.cs` | 6 |
| 本地服务生命周期 | 玩家只安装 Mod，不需要手动开 Go 服务 | Sidecar、Health Check、Ownership、Graceful Shutdown | `CoreProcessSupervisor.cs` → `SystemCoreProcess.cs` → `ModEntry.cs` | 7 |

## 3. 必备知识点

- [ ] 能说明一次教学从 F7 到 SQLite 提交的完整调用链。
- [ ] 能解释为什么让 LLM 输出 `TraitObservation`，而不是直接覆盖 `PlayerModel`。
- [ ] 能手算一次置信度增长公式，并说明冲突为何按情境隔离。
- [ ] 能解释 `saveId + sessionId + snapshotVersion` 分别解决什么问题。
- [ ] 能说明玩家目标占用为何既写入 Prompt，又必须由确定性代码再次校验。
- [ ] 能画出 `Idle → Recording → Learning → Ready → Acting → AwaitingResult/Correcting` 状态机。
- [ ] 能解释 `ActionProposal` 中模型置信度与 Go 策略置信度的区别。
- [ ] 能说清失败经验、玩家纠正的权重上限和幂等语义。
- [ ] 能解释为什么逐格寻路不调用模型、什么时候才重新推理。
- [ ] 能说明 Echo 独立背包如何避免污染玩家背包，以及箱满时如何保证物品不丢。
- [ ] 能解释本地 sidecar 的启动、健康检查、进程归属和退出清理。
- [ ] 能清楚区分已自动验证、待真实游戏验证和待 Nexus 发布三类事实。

## 4. 推荐阅读（结合仓库）

| 主题 | 通用技术点 | 建议阅读位置 | 预计时间 | 读完能回答什么 |
| --- | --- | --- | ---: | --- |
| 产品闭环 | AI 原生游戏交互 | `README.md`、`docs/superpowers/specs/2026-09-24-echofarm-continuum-design.md` | 20 分钟 | 为什么这不是聊天助手或坐标宏？ |
| 多日学习入口 | 应用服务、幂等 | `echofarm-core/internal/learning/service.go` | 20 分钟 | 重复教学为何不会重复增加置信度？ |
| 画像合并 | 状态演化、证据模型 | `echofarm-core/internal/modeling/merger.go` | 35 分钟 | 如何处理一致证据、矛盾证据和天气上下文？ |
| Eino 编排 | 结构化生成、边界验证 | `echofarm-core/internal/intelligence/learning_graph.go`、`intent_graph.go`、`action_graph.go` | 35 分钟 | 哪些判断交给模型，哪些必须留在代码？ |
| 协作决策 | 时间窗口、资源声明 | `echofarm-core/internal/coordination/service.go` | 20 分钟 | Echo 如何避免与玩家处理同一目标？ |
| 数据一致性 | SQLite 事务、事件账本 | `echofarm-core/internal/memory/sqlite.go` | 40 分钟 | 哪些状态原子提交，动作结果如何与原决策关联？ |
| HTTP 边界 | 严格 JSON、错误映射 | `echofarm-core/internal/httpapi/handler.go` | 20 分钟 | 非法模型输出和服务不可用如何隔离？ |
| 游戏状态机 | 异步状态、单飞请求 | `stardew-echo-mod/src/EchoFarm.Bridge/Runtime/EchoSession.cs` | 25 分钟 | 为什么不会一帧触发多个动作？ |
| 游戏执行 | 主线程队列、寻路、物品安全 | `stardew-echo-mod/src/EchoFarm.Mod/StardewGamePort.cs` | 40 分钟 | Go 的高层动作如何安全落到游戏世界？ |
| 可解释 UI | 读模型、展示投影 | `internal/memoryview/service.go`、`EchoMemoryPresenter.cs`、`EchoMemoryOverlay.cs` | 20 分钟 | 面板内容来自哪里，为什么不让模型自由生成？ |
| 反思式策略 | 失败归因、可审计经验、候选降级 | `reflection_graph.go`、`internal/experience`、`policy/service.go` | 35 分钟 | 为什么 AI 能跨会话避免同类失败？ |
| 发布链路 | 可复现构建、平台矩阵 | `scripts/package-nexus.sh`、`scripts/verify-nexus-package.sh` | 25 分钟 | 如何避免把密钥、数据库和错误平台二进制打进包？ |

## 5. 自学提醒

若某文件或原理看不懂，请继续追问 AI；本技能负责给学习路径与题目，不提供逐行讲解。

建议按调用链阅读，不要先背类名：先运行 `./demo/run-continuum-demo.sh`，再从 HTTP handler 追到 learning/policy 服务，最后分别下钻 Eino、SQLite 和 C# 游戏执行。

## 6. 项目技术定位

这是一个 **AI 应用后端 + 游戏客户端基础设施的交叉项目**：Go/Eino 负责受约束的智能推理与长期记忆，C#/SMAPI 负责实时游戏集成，SQLite 与 HTTP 契约负责跨进程一致性，Shell/CI 负责跨平台交付。

## 7. 核心原理解析

### 7.1 从模型覆盖改为证据增量

问题：让模型每次生成完整画像会使一次偶然操作覆盖已有习惯，也无法解释变化来源。

机制：Eino 只输出本次示范的候选 Trait 和事件证据；`ModelMerger` 用确定性公式增长或衰减置信度，并保留最多十二条最近证据引用。

落点：`learning_graph.go` 定义 AI 输出边界，`merger.go` 负责长期状态演化，`sqlite.go` 把示范、画像、技能和修订放进同一事务。

### 7.2 情境化习惯

问题：雨天不浇水是环境决定，不代表玩家改变了“晴天先浇水”的习惯。

机制：Trait 带有 `any/sunny/rainy/storm/snow` 上下文，只有同键、同情境、不同值才计为矛盾。

落点：四天 Demo 中两次晴天形成稳定顺序，第三天雨天生成独立 Trait，晴天 Trait 的 `contradictionCount` 仍为零。

### 7.3 在线协同而非离线代办

问题：玩家和自动角色同时行动时，静态计划容易重复浇同一株作物，体验像抢控制权。

机制：Bridge 只发送最近二十秒的语义动作；Eino 推断当前意图；Go 将这些目标声明为玩家正在处理，并在模型选择后再次执行硬校验。

落点：`PlayerActivityWindow` 控制短时上下文，`coordination.Service` 形成分工上下文，`policy.Service` 拒绝冲突目标并持久化候选与最终动作。

### 7.4 跨进程一致性

问题：网络超时、重复请求或世界状态变化可能让旧动作落到新状态。

机制：每个动作携带存档、会话和快照三元组；相同快照重复请求直接返回账本中的决定；结果必须与已记录的最终动作完整匹配。

落点：Go 的 `policy.Service` 与 SQLite 决策表负责服务端幂等，C# 的 `EchoFarmClient` 和 `ActionSafetyGate` 再做响应关联与本地状态校验。

### 7.5 AI 与确定性代码的责任分离

问题：模型适合解释语义和权衡，但不适合直接操作内存、执行逐帧寻路或决定安全边界。

机制：LLM 只生成受 JSON 结构约束的 Trait、Intent、ActionProposal 和 ExperienceObservation；Go 验证动作、证据与目标，C# 在主线程执行确定性寻路和资源转移。

落点：学习、意图、动作和反思四个 Eino 图，与 Go policy 校验和 C# safety gate 形成逐层收窄的信任边界。

### 7.6 失败和玩家纠正如何真正改变策略

问题：只在失败后当场换一个动作，下一次仍会重复犯错；如果只记自由文本反思，又无法稳定匹配、安全执行和证明来源。

机制：Reflection Graph 只在新失败或新纠正上运行一次，输出由 trigger、情境、有限信号、规避动作、优先动作和证据 ID 组成的经验。Go 根据语义主键合并、对矛盾经验衰减，并在后续快照中只匹配 Top-3。Action Graph 输出主动作、两个备选和模型置信度，最终置信度、候选选择与安全停止由确定性代码完成。

落点：`run-reflective-demo.sh` 会在两次进程重启后分别验证失败经验和 F10 纠正经验改变下一个会话的首次决策。

## 8. 关键设计决策

| 决策 | 备选 | 取舍 | 风险 | 验证 |
| --- | --- | --- | --- | --- |
| LLM 输出 Trait Patch | LLM 重写整份画像 | 多一次合并逻辑，换来稳定、可审计的长期记忆 | 合并公式可能过于保守 | 多日一致/冲突/情境单测 |
| HTTP 本地 sidecar | 全部逻辑写进 Mod | 增加进程管理，换来 Go/Eino 独立演进与隔离 | 启动和退出时序 | supervisor 测试、健康检查 |
| 单步高层动作 | 一次生成完整动作序列 | 增加模型调用，换来世界变化后的及时重规划 | 延迟与调用成本待测 | 四场景与失败回放 Demo |
| 目标声明后置校验 | 只在 Prompt 中提醒 | 可能保守停止，但模型无法抢占玩家目标 | 玩家活动误判 | 协调器和 policy 单测 |
| 本地 SQLite 账本 | 只存最新画像 | 增加存储量，换来幂等、解释与诊断 | 长期体积待测 | 事务、重放和读取测试 |

## 9. 量化与验证（含待测，建议）

当前可核验证据：Go 全包 race 测试与 vet 通过；.NET Bridge 87 个测试通过；单日、四天 Continuum 和五段反思策略跨进程 Demo 通过；四个平台的 Nexus 包结构烟测通过；macOS arm64 sidecar `/healthz` 实测成功。

上线前建议补测：

- `待测`：真实模型下单次 Trait 提取与动作决策的 P50/P95 延迟及 token 成本；
- `待测`：包含 50、200、1000 个农场目标时的快照体积和决策延迟；
- `待测`：连续 30 个游戏日后的数据库大小、画像收敛速度与冲突率；
- `待测`：真实 Stardew Valley 1.6 + SMAPI 环境中的动作成功率、误抢目标率和异常退出恢复；
- `待测`：Windows、Linux、macOS Intel/Apple Silicon 的安装与 sidecar 启动成功率。
