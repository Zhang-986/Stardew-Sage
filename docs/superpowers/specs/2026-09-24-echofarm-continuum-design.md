# EchoFarm Continuum：持续学习与协作分身设计

## 1. 目标

EchoFarm Continuum 把当前“一次教学、单步执行”的 MVP 扩展为一个会随玩家多日成长、能识别玩家正在做什么并主动承担互补工作的 AI 分身。

演示必须让观察者直接看到三件事：

1. Echo 不会因为一次偶然行为就改变性格，而是从多天证据中形成稳定、可追溯的画像；
2. Echo 与玩家同时工作时不会争抢同一个目标，而会依据玩家当下意图选择互补任务；
3. Echo 的画像变化和每次决定都能在游戏内解释，且所有动作仍经过确定性安全校验。

项目仍然不使用 RAG、MCP、Web 聊天页或远程数据库。AI 的价值集中在行为理解、偏好推断、意图识别和情境决策。

## 2. 路线选择

### 2.1 采用：持续学习的协作分身

在现有农场晨间闭环上增加多日记忆、画像合并、玩家实时意图和任务避让。它复用已经稳定的浇水、补水、收获、存箱与失败重规划能力，把工程深度放在长期智能和人机协作上。

### 2.2 未采用：横向扩展大量技能

增加钓鱼、战斗、社交和购物能扩大动作数量，但每类动作都需要新的 SMAPI 适配和安全边界。短期内会形成“功能多但每项很浅”的 Demo，无法强化最独特的玩家建模主张。

### 2.3 未采用：自由生活的生成式 NPC

自由对话、情绪和关系网络视觉效果明显，但会把重点从“学习玩家并成为另一个自己”转向普通生成式 NPC，也更依赖大量内容生成和不可控状态。

## 3. 主演示流程

Demo 使用同一存档的四天轨迹：

1. 第一天晴天，玩家先浇水、再收获，并把产物放入东侧箱子；Echo 得到低置信度候选习惯。
2. 第二天晴天，玩家重复主要顺序和箱子选择；Echo 提升稳定习惯的观测次数与置信度。
3. 第三天下雨，玩家直接收获并存箱；画像把“雨天跳过浇水”解释为情境行为，不把它错误计为任务顺序冲突。
4. 第四天玩家与 Echo 同时进入农场。玩家先对北侧作物浇水，Echo 从最近活动推断玩家正在处理浇水目标，避开该目标集合，转去收获成熟作物并存入已学会的东侧箱子。

演示最后打开“回声记忆”面板，显示画像版本、稳定习惯、证据天数、置信度、当前玩家意图、Echo 的最后决定和对应原因。

离线脚本必须完整复现四天流程；真实游戏只负责把相同契约接到 SMAPI 事件和绘制层。

## 4. 架构

```text
SMAPI 玩家事件 ──> Teaching Episode ──> Eino Trait Extractor
                                             |
                                             v
SQLite Evidence Ledger ──> Deterministic Model Merger ──> Player Model vN
         |                                                   |
         |                                                   v
         |      实时玩家活动 ──> Intent Window ──> Coordination Context
         |                                                   |
         v                                                   v
Execution Ledger <── 安全执行结果 <── C# Bridge <── Eino Runtime Policy
         |
         └──────────────> Explanation API ──> 游戏内“回声记忆”面板
```

边界保持清晰：

- Eino 负责从复杂轨迹中提取候选习惯、推断实时意图，并在多个合法目标中做符合玩家风格的选择；
- Go 领域层负责证据合并、置信度更新、任务避让、幂等和安全约束；
- SQLite 保存原始教学、画像修订和执行账本；
- C# 只采集游戏事件、维护短时玩家活动窗口、执行已校验动作并渲染解释信息。

## 5. 持续学习模型

### 5.1 TraitObservation

模型不再直接返回整份 `PlayerModel`，而是返回本次教学产生的候选观察：

```text
key                 习惯键，例如 task_order、preferred_chest
value               候选值
context             sunny、rainy 或 any
supportingEventIds  本次示范中的证据事件
strength             本次证据强度，范围 0..1
```

所有事件引用在持久化时转换为 `demonstrationId:eventId`，避免不同教学场次的事件 ID 冲突。

### 5.2 TraitMemory

稳定画像中的每条习惯包含：

```text
key
value
context
confidence
observationCount
contradictionCount
firstSeenDay
lastSeenDay
evidenceRefs
```

`PlayerModel` 保留现有运行期快捷字段，同时增加 `traits` 和 `learnedThroughDay`。快捷字段由合并后的有效 Trait 投影产生，运行策略不直接信任模型生成的整份对象。

### 5.3 确定性合并

`ModelMerger` 对 AI 提取结果执行以下规则：

- 同键、同值、同情境的观察增加 `observationCount`，置信度按 `1-(1-old)*(1-strength*0.5)` 增长；
- 同键、不同值且情境相同的观察给旧值增加一次冲突，并把其置信度乘以 `0.8`；
- `rainy` 与 `sunny` 是不同情境，不互相计为冲突；
- 单条证据不能超过 `0.75` 置信度，两次一致观察后才能成为稳定习惯；
- 每条习惯最多保留 12 个最新证据引用，原始示范仍永久保存在 evidence ledger；
- 相同 `demonstrationId` 重放时返回原结果，不重复提高置信度或修订版本。

这样 AI 负责理解“这次行为意味着什么”，代码负责决定“这份证据应当如何改变长期记忆”。

## 6. 实时协作

### 6.1 PlayerActivity

C# 在 Echo 工作期间保留最近 20 秒的玩家语义动作，不上传逐帧移动：

```text
kind
targetId
tick
success
```

`WorldSnapshot` 增加 `recentPlayerActions`。连续重复动作按目标去重，窗口外事件自动丢弃。

### 6.2 CoordinationContext

Go 在调用运行策略前构造：

```text
inferredIntent       watering、harvesting、depositing 或 unknown
playerClaimedTargets 玩家刚刚操作或正在接近的目标
availableGoals       当前仍可完成的目标摘要
modelRevision        本次决定使用的画像版本
```

Eino 使用最近玩家动作、画像和世界状态推断意图。确定性协调器随后过滤 `playerClaimedTargets`，因此即使模型返回了冲突目标，也不能让 Echo 抢占玩家正在处理的作物。

当所有可做目标都被玩家占用时，Echo 返回 `stop_session`，原因是等待玩家完成，而不是反复抢占。

### 6.3 决策频率

一次高层动作完成后才请求下一动作。逐格寻路和动画不调用模型。若玩家活动窗口没有变化且世界快照的可行动目标集合相同，Go 可复用上一轮推断出的意图，但每个高层动作仍产生独立账本记录。

## 7. 执行账本与解释

SQLite 新增三类记录：

- `learning_revisions`：每次教学的输入摘要、提取出的 Trait、合并前后画像版本；
- `echo_sessions`：会话开始/结束、存档、日期和最终状态；
- `decision_records`：世界快照版本、画像版本、玩家意图、避让目标、AI 原始候选、最终动作、执行结果与失败码。

`GET /v1/echo/memory?saveId=...` 返回面向玩家的只读视图：

```text
modelRevision
learnedThroughDay
stableTraits[]
recentLearningChange
activeSession
lastDecision
```

解释内容只能来自已校验的结构化字段。API 不让模型临时编一段人物小传，也不暴露完整原始提示词或密钥。

## 8. 游戏内体验

新增可配置热键 `F9` 打开或关闭“回声记忆”面板。面板包含：

- `Echo vN · 学习到第 D 天`；
- 最多四条稳定习惯及置信度；
- 当前推断：`你正在浇水`；
- 当前分工：`Echo 去收获 crop-south-2，避开你正在处理的 2 株作物`；
- 最近学习变化：某条习惯是增强、减弱还是新增。

面板仅在已有记忆时显示，不暂停游戏。Core 不可用时显示一行离线状态后自动隐藏详细内容，不影响玩家继续操作。

## 9. API 与兼容性

保留已有端点并扩展请求/响应：

- `POST /v1/demonstrations/learn` 返回 `playerModel`、`skill` 和 `learningChange`；
- `POST /v1/echo/next-action` 接受包含 `recentPlayerActions` 的快照；
- `POST /v1/echo/action-result` 在产生下一动作前写入上一动作结果；
- `GET /v1/echo/memory?saveId=...` 提供 HUD 数据。

新增 JSON 字段在 C# 与 Go 两端都为可选，旧的教学 fixture 和旧存档仍可读取。SQLite 使用 `CREATE TABLE IF NOT EXISTS` 做向前迁移，不改写现有三张表。

## 10. 错误与安全策略

- AI 返回未知 Trait key、越界 strength 或不存在的事件证据时，整次学习失败且事务零写入；
- 画像合并和学习修订在同一 SQLite 事务提交；
- 意图推断失败时不伪装为智能，Echo 安全停止并保留上一份画像；
- 模型选择被玩家占用、已消失或不满足前置条件的目标时，Go 拒绝并返回安全停止，不把非法动作交给 C#；
- 重复提交同一教学或动作结果保持幂等；
- HUD 获取失败不影响教学和执行；
- 保存、返回标题或切换存档仍立即取消当前请求和动作。

## 11. 测试策略

### Go

- `ModelMerger` 覆盖一致证据增长、同情境冲突、跨情境不冲突、证据上限和幂等；
- SQLite 覆盖旧数据库迁移、学习事务、修订读取和决策账本；
- Eino 图覆盖非法 Trait、伪造事件引用和模型不可用；
- 协调器覆盖目标避让、全目标占用、过期玩家活动和未知意图；
- HTTP 覆盖多日学习、memory view、动作结果落账和严格 JSON；
- `go test -race ./...` 验证并发安全。

### .NET Bridge

- JSON 契约覆盖新增字段和向后兼容；
- 玩家活动窗口覆盖去重、过期清理和存档隔离；
- EchoSession 覆盖活动上传、停止和过期响应；
- memory presenter 覆盖稳定排序、缺省状态和 Core 离线。

### 跨进程 Demo

`demo/run-continuum-demo.sh` 启动真实 Go 进程，依次提交三天教学并断言：

- 画像修订严格递增；
- 两次晴天一致行为形成稳定习惯；
- 雨天没有削弱晴天浇水顺序；
- 第四天玩家声明北侧浇水目标后，Echo 选择未被占用的收获目标；
- memory view 能解释画像证据和最后一次分工。

Fixture 模型只为可复现测试提供结构化 AI 输出，不在真实模型故障时自动接管。

## 12. 完成标准

- 同一存档至少三次教学能形成可追溯、情境化、置信度演化的长期画像；
- 重复教学不会重复计数，一次矛盾行为不会立即覆盖稳定习惯；
- Echo 能根据最近玩家活动避开玩家正在处理的目标并选择互补工作；
- 每次学习和运行决定都进入 SQLite 账本，可通过 memory API 查询；
- C# Bridge 能采集近期玩家活动并具备游戏内记忆面板的数据与渲染入口；
- 四天跨进程 Demo、Go race/vet、.NET 测试和 Nexus 打包烟测全部通过；
- 不提交游戏程序集、数据库、日志、模型密钥或伪造的可玩 Mod 包。
