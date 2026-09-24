# EchoFarm：会学会你玩法的 AI 化身

EchoFarm 让玩家通过正常游玩，训练出一个能进入《星露谷物语》世界、与玩家并行行动的“另一个自己”。它学习的是目标、顺序和取舍，不是昨日坐标或按键宏。

当前已经完成 Go + Eino 的持续学习、协作与反思闭环：

```text
多日示范 -> Eino Trait 提取 -> 确定性证据合并 -> 版本化玩家画像
                                                |
砍树/采矿/矿层/钓鱼 -> 语义活动分类 -> 可追溯生活方式画像
                                                |
实时玩家动作 -> 意图推断 -> 目标避让 -> Eino 决策 -> 双重安全校验
                                                |
                      游戏内记忆面板 <- SQLite 决策与执行账本
                                                |
          模型调用 -> 用途/延迟/token/失败账本 -> 会话预算 -> 安全停止
                                                |
失败/玩家 F10 纠正 -> Eino 反思 -> 结构化经验 -> 下次提前改变策略
                                                |
实际执行结果 -> 幂等反馈账本 -> 有效置信度 -> 强化有效经验/冷却反复失效经验
```

## 为什么 AI 不可替代

- 从连续操作中理解“照料全部作物”等意图；
- 把一次示范编译为可迁移到新布局的技能；
- 以有证据的画像保留玩家习惯，而不是生成一段人物小传；
- 在雨天、新作物、空水壶或路线阻塞时根据实时状态重规划。
- 从多天证据中区分稳定习惯与偶然选择，并保留每项结论的来源；
- 识别玩家正在处理的目标，主动承担互补工作而不是与玩家抢活。
- 将动作失败先写入可恢复的 SQLite 反思任务，再抽象为可复用策略经验；模型暂时失败或进程重启也不会丢掉待学习内容；
- 玩家可按 F10 否定当前决策，用下一次成功操作教会 Echo 更好的选择。
- 只对本次真正采用并执行的经验记录结果反馈，用可审计的成功/矛盾计数调整后续排序；路线变化等环境噪声保持中性，避免错误惩罚 AI 经验。
- 每次 AI 调用都有本地 request ID、用途、延迟、失败分类和 provider 上报 token；会话预算耗尽时直接返回可审计的安全停止。

模型只决定高层动作。白名单、目标存在性、天气、工具和体力检查由确定性代码把关。项目不使用 RAG，也不依赖 Web 聊天界面。

## 快速演示

需要 Go 1.24.1+、`curl` 和 `jq`：

```bash
./demo/run-core-demo.sh
```

脚本会用不联网的 fixture 模型验证完整管线，并输出：

1. 从晨间农活示范形成的玩家画像与技能；
2. 雨天且布局变化时，对一个全新成熟作物选择 `harvest_target`；
3. 晴天水壶为空时，先选择 `refill_can`；
4. 收获因 Echo 背包已满而失败时，转去玩家教过的箱子执行 `deposit_items`。

更完整的四天成长与协作演示：

```bash
./demo/run-continuum-demo.sh
```

它会连续提交两次晴天教学和一次雨天教学，验证画像置信度与情境记忆，再模拟玩家正在浇北侧作物，证明 Echo 会避开玩家目标并接手南侧收获任务。游戏内按 F9 可查看画像与最后一次分工理由。

反思与玩家纠正演示：

```bash
./demo/run-reflective-demo.sh
```

它会启动真实 Go 进程并两次重启，证明“首次满背包收获失败 -> 形成经验 -> 下次提前存箱 -> 玩家纠正箱子 -> 再下次优先新箱子 -> 成功结果提高该经验的有效置信度”的跨会话学习链。反馈是追加式、幂等且与规范动作结果同事务提交的，不依赖模型给自己打分。

扩展玩法分类与画像演示：

```bash
./demo/run-activity-learning-demo.sh
```

它会提交两天版本化语义事件，覆盖砍树、破石、矿层迁移和钓鱼，验证 AI 将这些事件归纳为稳定的活动顺序、资源偏好、下矿与钓鱼习惯，并证明未认证玩法只参与学习、不会进入可执行动作白名单。

使用真实 OpenAI-compatible 模型：

```bash
cd echofarm-core
export ECHOFARM_MODEL_MODE=openai
export ECHOFARM_MODEL_BASE_URL=https://your-endpoint/v1
export ECHOFARM_MODEL_API_KEY=your-key
export ECHOFARM_MODEL_NAME=your-model
export ECHOFARM_MAX_MODEL_CALLS_PER_SESSION=32
export ECHOFARM_MAX_REPORTED_TOKENS_PER_SESSION=100000
go run ./cmd/echofarm
```

用量账本只保存请求用途、状态、延迟和 provider 明确返回的 token 数，不保存 prompt、response、API Key 或 provider 错误正文。如果 provider 不返回 token，F9 会显示 `unknown`，不会伪造估算；项目也不计算货币成本，因为模型价格是外部可变配置。

详细配置见 [echofarm-core/README.md](echofarm-core/README.md)，产品设计见 [EchoFarm 设计](docs/superpowers/specs/2026-09-23-echofarm-player-model-design.md)、[Reflective Policy 设计](docs/superpowers/specs/2026-09-24-echofarm-reflective-policy-design.md) 和 [经验有效性反馈设计](docs/superpowers/specs/2026-09-24-echofarm-experience-feedback-design.md)。

## Nexus Mods 打包

正式包按 Windows x64、Linux x64、macOS Intel、macOS Apple Silicon 分开发布，每个压缩包都内置对应的 Go/Eino 服务端，玩家不需要安装 Go。打包命令、发布文案与核对清单见 [release/nexus/README.md](release/nexus/README.md)。

第一个生产候选版的明确范围是 Windows x64 + Stardew Valley 1.6 + SMAPI 4.1+ + 单人模式。首次启动默认为无付费调用的 `fixture` 演示模式，F9 会明确显示 `DEMO`；收获能力在完成原生游戏语义认证前默认关闭。

Windows 仓库所有者可直接执行：

```powershell
$game = "C:\Program Files (x86)\Steam\steamapps\common\Stardew Valley"
.\scripts\windows\Install-EchoFarm.ps1 -Doctor -GamePath $game
.\scripts\windows\Install-EchoFarm.ps1 -Build -GamePath $game
.\scripts\windows\Install-EchoFarm.ps1 -Install -GamePath $game -PackagePath .\dist\windows\EchoFarm
```

实机验收按 [Windows 冒烟清单](docs/echofarm/windows-smoke-checklist.md) 执行。构建脚本会同时产出 SHA-256 和 `EchoFarm.evidence.json`，游戏内项目在真实观察前一律保持 `pending`。

```bash
./scripts/package-nexus.sh --version 0.4.0 --game-path "/path/to/Stardew Valley"
```

仓库不会打包游戏程序集、API Key、数据库或日志。首次创建 Nexus 页面后，将分配到的 mod ID 通过 `--nexus-mod-id` 注入发布包的 SMAPI 更新键。

## 目录

```text
contracts/         C# 与 Go 共用的 JSON 契约
echofarm-core/     Go + Eino AI 大脑与 SQLite 记忆
demo/              可复现的教学和变化环境样例
docs/echofarm/     游戏桥接协议
stardew-echo-mod/  SMAPI 传感器、受控执行器与游戏内记忆面板
```

`aurora-admin/`、`aurora-mcp/` 和 `aurora-ui/` 是仓库原有实验代码，不属于 EchoFarm 主链路；第一阶段暂时保留，后续再迁移或清理。

## 当前进度

- [x] 跨语言结构化契约与动作白名单
- [x] 示范轨迹压缩与行为分段
- [x] Eino 学习、决策和失败重规划图
- [x] 按存档隔离的 SQLite 玩家记忆
- [x] localhost API 与离线可复现 Demo
- [x] 可测试的 .NET 桥接核心、行为采集器和 Go HTTP 客户端
- [x] SMAPI 适配代码、Echo 半透明渲染及逐格寻路
- [x] Echo 独立背包、成熟作物收获和目标箱存放闭环
- [x] 背包满转存、箱子满安全停止的 AI 失败重规划
- [x] 多日证据合并、情境化画像与幂等学习修订
- [x] 实时玩家意图推断、目标避让和协作分工
- [x] SQLite 决策账本、解释 API 与 F9 游戏内记忆面板
- [x] 失败反思、Top-3 情境经验匹配和跨会话主动避错
- [x] 带置信度和最多两个备选的可解释动作提案
- [x] F10 显式玩家纠正、20 秒捕获窗口与幂等经验落账
- [x] 规范动作结果反馈、经验有效置信度排序与反复失效经验冷却
- [x] 收获默认关闭的跨语言能力门禁与执行前二次校验
- [x] 首启配置诊断、端口冲突识别与 F9 运行状态面板
- [x] Windows doctor/build/install/uninstall 工作流与自动证据 JSON
- [x] SQLite AI 调用账本、provider token 统计、原子会话预算与 F9 用量面板
- [x] 版本化语义活动协议、多帧分类器及砍树/采矿/矿层/钓鱼画像闭环
- [ ] 在安装 Stardew Valley + SMAPI 的机器上完成编译与游戏内冒烟

## 安全边界

- 服务默认只监听 `127.0.0.1`；
- 仓库不保存有效模型密钥；
- 模型不能修改金币、好感度、存档或生成任意物品；
- 未认证的收获动作在 Go 策略层和 C# 游戏执行层都被拒绝；
- 砍树、采矿、矿层迁移和钓鱼当前为 `learn-only`，观察能力不会隐式获得游戏修改权限；
- 模型不可用时只停止 Echo，不影响游戏保存和退出。

## License

[MIT](LICENSE)
