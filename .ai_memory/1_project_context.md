# 项目上下文

## 仓库状态

- 仓库路径：`d:\Project\anti`
- 当前目录已初始化为 Git 仓库，并已接入远端 `https://github.com/Youkies/Ant-Browser.git`
- 当前本地分支为 `master`，跟踪 `origin/master`

## 现有配置

- `.vscode/settings.json` 中启用了 `"chatgpt.openOnStartup": true`

## 技术栈

- 桌面应用框架：Wails v2（见 `wails.json`）
- 后端语言：Go 1.22（模块名 `ant-chrome`）
- 前端框架：React 18 + TypeScript + Vite
- 前端状态管理：Zustand
- 前端样式体系：Tailwind CSS
- 代理池页面额外使用 `country-flag-icons` 渲染本地 SVG 国旗

## 目录概览

- `backend/`：Go 后端主逻辑、内部模块与测试
- `frontend/`：React 前端界面、页面、共享组件与主题
- `bin/`：运行时二进制与平台相关资源
- `build/`、`publish/`、`bat/`、`scripts/`、`tools/`：构建、发布与辅助脚本

## 已知限制与注意事项

- 当前尚未读取完整业务流程与模块边界，后续修改前仍需按功能点继续定位代码
- 本地保留了仓库外新增的 `.ai_memory/` 目录，用于本项目记忆，不属于远端仓库内容

## 已验证实现约定

- 代理桥接分层：`vmess / vless / trojan / ss` 走 `xray`，`hysteria2 / tuic / anytls` 走 `sing-box`
- 代理池页面的节点 IP 健康结果来自后端 `IPPure` 查询，前端基于该结果渲染国家/地区与节点国旗
