# EchoFarm：会学会你玩法的 AI 化身

EchoFarm 让玩家通过正常游玩，训练出一个能进入《星露谷物语》世界、与玩家并行行动的“另一个自己”。它学习的是目标、顺序和取舍，不是昨日坐标或按键宏。

当前已经完成第一条 Go + Eino 核心闭环：

```text
示范事件 -> 行为分段 -> Eino 学习图 -> 玩家画像 + 技能
                                        |
当前农场 -> Eino 决策图 -> 安全校验 -> 下一动作
                         ^          |
                         |--失败重规划
```

## 为什么 AI 不可替代

- 从连续操作中理解“照料全部作物”等意图；
- 把一次示范编译为可迁移到新布局的技能；
- 以有证据的画像保留玩家习惯，而不是生成一段人物小传；
- 在雨天、新作物、空水壶或路线阻塞时根据实时状态重规划。

模型只决定高层动作。白名单、目标存在性、天气、工具和体力检查由确定性代码把关。项目不使用 RAG，也不依赖 Web 聊天界面。

## 快速演示

需要 Go 1.24+、`curl` 和 `jq`：

```bash
./demo/run-core-demo.sh
```

脚本会用不联网的 fixture 模型验证完整管线，并输出：

1. 从晨间农活示范形成的玩家画像与技能；
2. 雨天且布局变化时，对一个全新成熟作物选择 `harvest_target`；
3. 晴天水壶为空时，先选择 `refill_can`。

使用真实 OpenAI-compatible 模型：

```bash
cd echofarm-core
export ECHOFARM_MODEL_MODE=openai
export ECHOFARM_MODEL_BASE_URL=https://your-endpoint/v1
export ECHOFARM_MODEL_API_KEY=your-key
export ECHOFARM_MODEL_NAME=your-model
go run ./cmd/echofarm
```

详细配置见 [echofarm-core/README.md](echofarm-core/README.md)，产品设计见 [EchoFarm 设计](docs/superpowers/specs/2026-09-23-echofarm-player-model-design.md)。

## 目录

```text
contracts/         C# 与 Go 共用的 JSON 契约
echofarm-core/     Go + Eino AI 大脑与 SQLite 记忆
demo/              可复现的教学和变化环境样例
docs/echofarm/     游戏桥接协议
stardew-echo-mod/  下一阶段的薄 SMAPI 传感器/执行器
```

`aurora-admin/`、`aurora-mcp/` 和 `aurora-ui/` 是仓库原有实验代码，不属于 EchoFarm 主链路；第一阶段暂时保留，后续再迁移或清理。

## 当前进度

- [x] 跨语言结构化契约与动作白名单
- [x] 示范轨迹压缩与行为分段
- [x] Eino 学习、决策和失败重规划图
- [x] 按存档隔离的 SQLite 玩家记忆
- [x] localhost API 与离线可复现 Demo
- [x] 可测试的 .NET 桥接核心、行为采集器和 Go HTTP 客户端
- [x] SMAPI 适配代码、Echo 半透明渲染及浇水/补水执行骨架
- [ ] 在安装 Stardew Valley + SMAPI 的机器上完成编译与游戏内冒烟
- [ ] 补齐 Echo 独立背包下的收获和存箱执行

## 安全边界

- 服务默认只监听 `127.0.0.1`；
- 仓库不保存有效模型密钥；
- 模型不能修改金币、好感度、存档或生成任意物品；
- 模型不可用时只停止 Echo，不影响游戏保存和退出。

## License

[MIT](LICENSE)
