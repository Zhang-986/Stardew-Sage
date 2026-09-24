# EchoFarm Reflective Policy：结果驱动的自我纠错设计

## 1. 目标

EchoFarm Reflective Policy 让分身从“会模仿、会临场重规划”继续演进为“会从结果和玩家纠正中形成经验，并在下一次相似局面发生前主动改变策略”的 AI 玩家。

本阶段必须让演示者直接证明四件事：

1. AI 一次决策可以输出主动作、最多两个备选动作和置信度，而不是只给一个不可解释答案；
2. 动作失败后，AI 会抽象出带适用条件的结构化经验，而不是仅在当前请求里临时换一个动作；
3. 玩家可以在游戏内明确否定 Echo 最近的选择，并用下一次成功操作示范正确做法；
4. 后续相似局面会命中经验，在首次尝试前改变计划，且 F9 面板能展示经验来源与采用原因。

项目继续不使用 RAG、MCP、多 Agent 辩论或向量数据库。长期记忆仍是可审计的领域数据，Eino 负责语义抽象和候选规划，确定性代码负责合并、匹配、校准和安全执行。

## 2. 方案选择

### 2.1 采用：事件触发的混合反思闭环

只在动作失败或玩家明确纠正时调用一次 Reflection Graph。模型将具体事件归纳为有限枚举表达的经验；代码验证经验引用、作用域和动作白名单，再合并进 SQLite。运行时 Action Graph 接收与当前状态匹配的少量经验，并一次输出有序候选，不启动递归思考循环。

该方案把模型调用集中在需要语义判断的位置，同时保持延迟、费用和状态变化可控。

### 2.2 不采用：只扩写现有 Prompt

Prompt 可以提醒模型关注失败，但失败上下文不会跨会话保存，也无法证明下一次决策使用了哪条经验。它只能改善单次回答，不能形成持续学习闭环。

### 2.3 不采用：Planner/Critic 多轮自我讨论

每个动作执行多个模型回合会显著增加延迟和成本，还会产生难以重放的中间状态。对实时农场动作而言收益不足，且会削弱当前“一次高层推理、确定性执行”的清晰边界。

## 3. 玩家体验与主演示

新增一个可配置的“纠正 Echo”按键，默认 `F10`。

五段演示流程：

1. 玩家完成既有多日教学，形成带证据的稳定画像；
2. Echo 在背包已满时执行收获并收到 `inventory_full`，随后重规划为存箱；
3. Reflection Graph 把这次失败归纳为“背包满且存在可用箱子时，收获前优先存箱”，经验以中等置信度写入账本；
4. 新会话再次出现“背包满 + 成熟作物”，Action Graph 在首次执行前选择存箱，并输出收获与停止作为备选，证明经验发生了跨会话迁移；
5. 玩家对 Echo 选择的箱子按 `F10`，然后亲自把物品放入另一个箱子。该成功动作作为高信号纠正提交；下一次相同情境，Echo 使用玩家示范的箱子，F9 面板显示纠正证据、经验置信度和备选策略。

按下 `F10` 后 Echo 暂停行动，只捕获玩家接下来的一个成功农场动作。纠正窗口超过二十秒、切换存档或返回标题时自动取消，不阻塞正常游戏。

## 4. 总体架构

```text
WorldSnapshot + PlayerModel + ApplicableExperiences
                         |
                         v
              Eino Action Proposal Graph
                         |
          primary + alternatives + model confidence
                         |
                         v
        Deterministic Candidate Selector / Safety Gate
                         |
                         v
               C# SMAPI action execution
                         |
              result or explicit correction
                         |
                         v
                Eino Reflection Graph
                         |
       bounded ExperienceObservation with evidence
                         |
                         v
       deterministic merge + SQLite experience ledger
                         |
             next matching decision + F9 view
```

保留现有分层：

- C#/SMAPI 负责事件捕获、纠正窗口、主线程动作执行和 HUD；
- HTTP 契约负责跨进程关联，不传递自由文本命令；
- Eino Action Graph 负责在画像、实时世界和经验之间做语义权衡；
- Eino Reflection Graph 负责从结果或纠正中抽象经验；
- Go policy 负责候选选择、置信度治理、幂等和安全验证；
- SQLite 保存原始证据、经验演进和最终采用情况。

## 5. 决策提案

模型不再直接返回单个 `HighLevelAction`，而是返回：

```text
ActionProposal
  primary             首选高层动作
  alternatives[0..2]  有序备选动作
  modelConfidence     0..1，仅表示模型对当前排序的把握
  uncertaintyCodes    有限枚举：missing_experience、conflicting_evidence、novel_context、ambiguous_target
  appliedExperienceIds 模型声称使用的经验 ID
```

所有动作必须使用当前 `saveId + sessionId + snapshotVersion`。主动作和备选动作不能重复，目标必须存在于当前快照，经验 ID 必须来自本次输入。任何未知字段、枚举或伪造引用都会使整次模型输出无效。

`modelConfidence` 不直接充当安全权限。Go 生成最终 `policyConfidence`：

```text
policyConfidence = clamp(
  modelConfidence
  + 0.10（命中稳定玩家 Trait）
  + 0.15（命中已强化经验）
  - 0.20 × uncertaintyCodes 数量,
  0,
  1
)
```

经验达到两次一致证据或来自一次显式玩家纠正时视为“已强化”。当最终置信度低于 `0.35` 时，策略不执行有副作用动作，返回 `stop_session` 并记录低置信度原因。

Go 按顺序验证 primary 和 alternatives，选择第一个合法且不与玩家目标冲突的动作。这样主候选因世界变化失效时，可以使用同一次推理给出的备选，不需要额外一次模型调用；所有候选都不合法时安全停止。

## 6. 结构化经验

### 6.1 ExperienceObservation

Reflection Graph 只能输出有限结构：

```text
trigger              inventory_full、out_of_water、path_blocked、chest_full、player_correction
context              any、sunny、rainy、storm、snow
whenSignals[]        inventory_full、inventory_has_items、can_empty、raining、target_blocked
avoidAction          要避免的动作类型，可空
preferAction         下次优先动作类型
preferredTargetId    仅玩家纠正可指定，且必须存在于纠正快照
summary              面向玩家的一句话解释
supportingDecisionId 或 correctionId
strength             0..1
```

模型不能生成任意谓词或脚本。`whenSignals`、trigger 和动作都来自白名单，因此运行时匹配完全由 Go 完成。

### 6.2 PolicyExperience

持久化经验包含：

```text
id
saveId
trigger/context/whenSignals
avoidAction/preferAction/preferredTargetId
confidence
observationCount/contradictionCount
firstSeenDay/lastSeenDay
evidenceRefs[]
```

ID 由 Go 根据规范化后的作用域、规避动作和优先动作计算，不接受模型生成 ID。相同经验增加观测次数和置信度；同一作用域下互相冲突的经验降低旧经验置信度。每条最多保存十二个证据引用。

失败反思的初始置信度最高为 `0.65`；显式玩家纠正视为高信号，初始置信度最高为 `0.85`。单次普通失败不会改写长期玩家 Trait，只形成策略经验。

## 7. 玩家纠正协议

`PlayerCorrection` 包含：

```text
id/saveId/sessionId
rejectedDecisionSnapshotVersion
rejectedAction
worldSnapshot
preferredAction
observedAtTick
```

C# 只允许把成功观察到的浇水、收获、补水或存箱动作转换为 `preferredAction`。纠正必须引用当前存档最近一条已持久化 decision，且纠正快照版本必须不早于该 decision。Go 验证目标存在、动作白名单和关联关系后才调用 Reflection Graph。

同一 correction ID 重试返回首次持久化结果，不重复提升经验置信度。Reflection Graph 不可用或输出非法时，纠正原始证据不写入半成品经验，Echo 保持暂停并向玩家显示可重试提示。

## 8. 数据与 API

SQLite 增加：

- `policy_experiences`：当前合并后的经验；
- `experience_revisions`：每次失败/纠正对应的不可变结果，用于幂等与解释；
- `player_corrections`：原始纠正事件。

现有 `decision_records` 增加可选字段：模型主候选、备选候选、模型置信度、策略置信度、命中的经验 ID、最终选中候选序号和不确定性代码。旧 JSON 记录仍可读取。

API 调整：

- `POST /v1/echo/next-action`：响应扩展为 action、confidence、alternatives、appliedExperiences；
- `POST /v1/echo/action-result`：失败结果原子落账后触发反思，再返回下一决策；
- `POST /v1/echo/corrections`：提交玩家显式纠正并返回形成的经验；
- `GET /v1/echo/memory`：增加近期经验、上次决策置信度和备选策略。

HTTP 字段采用向后兼容的可选扩展；旧客户端仍可读取 `action`。

## 9. 失败与降级

- Action Graph 不可用：不使用 fixture 冒充模型，Echo 停止，游戏继续；
- Reflection Graph 不可用：动作结果仍持久化，经验学习失败可重试，不阻塞保存和退出；
- 模型伪造经验引用或目标：拒绝整个输出，不写入长期经验；
- 主候选非法：尝试下一个合法备选并记录 rejected reason；
- 所有候选非法或置信度过低：返回 `stop_session`；
- 同一失败结果、纠正或快照重放：返回 SQLite 中的 canonical outcome；
- 读模型超时：F9 面板显示离线状态，不影响行动客户端。

## 10. 成本与性能约束

- 正常成功动作仍只调用一次 Action Graph；
- 只有新失败或新纠正调用一次 Reflection Graph；
- 每次决策最多注入三条匹配经验，每条只包含结构字段和一句摘要；
- 不保存或回放模型隐藏思维过程；
- 不执行递归规划，不让模型直接查询数据库；
- 超时、取消和模型错误沿用现有 sidecar 边界。

## 11. 测试策略

### Go 单元测试

- ActionProposal 拒绝错误关联、重复候选、未知不确定性和伪造经验引用；
- Candidate Selector 验证主动作、备选降级、低置信度停止和玩家目标冲突；
- Reflection Graph 验证失败/纠正证据、枚举边界和模型不可用；
- Experience Merger 验证强化、冲突衰减、显式纠正权重、证据上限和输入不可变；
- SQLite 验证失败结果、经验修订和纠正的原子提交与幂等；
- Policy 验证经验只在匹配情境注入，并确认重试返回 canonical 决策。

### C# Bridge 测试

- 新增 JSON 契约严格反序列化与向后兼容；
- EchoSession 保存最后决策并在纠正模式暂停；
- 纠正窗口只接受一个成功玩家动作，超时和存档切换会清理；
- 客户端校验 correction 的存档、会话和快照关联。

### 跨进程 Demo

新增五段脚本，断言：

1. 首次满背包失败生成 `inventory_full` 经验；
2. 新会话相同状态在首次执行前选择 `deposit_items`；
3. 响应包含置信度和至少一个备选策略；
4. 玩家纠正目标箱后，下一次选择纠正后的箱子；
5. memory API 能展示失败证据、纠正证据、置信度和采用经验。

## 12. 完成标准

- 从动作结果和显式玩家纠正生成的经验均能跨进程、跨会话持久化；
- 同类状态再次出现时，决策在首次执行前使用相关经验；
- 每个决策能输出主动作、最多两个备选、置信度和经验引用；
- 任何 AI 输出都不能绕过现有目标、资源、天气、体力和玩家占用校验；
- 五段自我纠错 Demo、Go race/vet、.NET 测试和四平台打包 smoke 全部通过；
- 真实模型和真实游戏环境未验证时，文档不得宣称生产上线或 Nexus 可玩发布。
