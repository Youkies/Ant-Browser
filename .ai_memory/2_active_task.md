# 当前任务

## 任务概览

- 当前阶段：应用品牌名已切换为 `Youkies Browser`，运行中的开发窗口标题已确认生效。
- 最近完成：统一了 Wails 配置、前端标题、默认配置、托盘标题、设置默认值、恢复脚本说明、README 与主要打包脚本中的应用名。
- 当前状态：`frontend/npm run build` 与 `go build ./...` 已通过，开发环境已重启，当前窗口标题显示为 `Youkies Browser`。

## 本轮改动

- 运行时显示名称从 `Ant Browser` 切换为 `Youkies Browser`。
- 顶部页面标题回退名称、HTML `<title>`、Launch API 文档页提示文案已同步更新。
- 后端默认配置、备份清单默认应用名、配置测试与备份测试已同步调整。
- Windows NSIS、Linux、macOS 发布脚本中的产品名与产物名已同步调整为 `Youkies Browser` 系列。
- 恢复脚本、构建脚本、发布脚本和仓库 README 中的应用名示例已同步更新。

## 待确认项

- 仓库中仍保留少量 `Ant Chrome Team` / `Ant Chrome` 字符串，当前属于发布者、公司名或内部工具标识，不影响运行中的应用名显示。
- `tools/public-release/publish-public.ps1` 与 `tools/keygen/main.go` 仍残留旧品牌文案，若后续要做完整品牌清理，可继续一并替换。
- `config.yaml` 仍属于本地运行态文件，本轮仅顺带把 `app.name` 对齐为 `Youkies Browser`，未处理其余本地差异。

## 下一步建议

- 如需继续做完整品牌清理，可统一替换发布者、公司名、内部发布脚本与遗留工具中的旧文案。
- 如需固化这次改动，可先检查 `git diff`，再按用户意图决定是否提交。
