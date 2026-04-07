# COMMON
- 不主动执行 build
- 不主动创建测试用例
- 不主动创建文档
- 写入 AGENTS 的内容需精炼且准确
- 单个文件的代码不宜过大，尽量不超过 500 行，特别是页面组件

# AGENT
- skill 统一只安装到 `.agents/skills` 目录下

# STRUCTURE
```text
.
├─ main.go                         # Wails 启动入口，只做装配
├─ app.go                          # Wails 绑定入口，只暴露桌面方法
├─ app_runtime.go                  # 根目录运行时桥接，不放具体业务
├─ internal/
│  ├─ apierr/                      # 统一 API 错误模型与错误构造
│  ├─ auth/                        # 会话、token、访问链接
│  ├─ clipboard/                   # 剪贴板轮询、写入、平台实现
│  ├─ config/                      # 运行时配置与路径解析
│  ├─ httpserver/                  # LAN HTTP/SSE 与静态资源服务
│  ├─ network/                     # 设备名、候选 LAN 地址、自检输入
│  ├─ runtimeapp/                  # 应用运行时编排与聚合视图
│  └─ store/                       # 持久化、查询、记录模型
├─ frontend/
│  ├─ src/
│  │  ├─ main.ts                   # bootstrap only
│  │  ├─ app/                      # 环境判断、启动、共享 UI helper
│  │  ├─ pages/                    # 桌面页、Web 页
│  │  ├─ components/
│  │  │  ├─ desktop/               # NaiveDesktop 组件
│  │  │  └─ web/                   # Web 组件
│  │  ├─ hooks/                    # 页面状态与交互编排
│  │  ├─ utils/                    # 纯工具与轻量 UI 资源
│  │  ├─ services/                 # API / Wails 交互
│  │  ├─ types/                    # 前端共享类型
│  │  ├─ styles/                   # 全局样式
│  │  └─ assets/                   # 静态资源
│  ├─ wailsjs/                     # 生成物，不手改
│  └─ dist/                        # 生成物，不手改
├─ docs/
│  ├─ demand/                      # 产品文档
│  └─ design/                      # 技术文档
└─ TODO.md
```
- 新增 Go 业务代码默认落 `internal/<domain>/`，不要再回到根目录平铺 `.go` 文件
- `frontend/src/main.ts` 只保留启动逻辑；页面、组件、状态、工具、API 拆到 `app/pages/components/hooks/utils/services/types`
- 桌面组件进 `components/desktop`，Web 组件进 `components/web`，状态逻辑进 `hooks`，纯工具进 `utils`
- 平台相关实现集中在领域目录下的 `*_windows.go` / `*_other.go`，不要散落

# PITFALLS
- Web 开发态下 `/web` 走 Vite 代理，生产环境继续走 `frontend/dist`
- Web 路由不要静态引入桌面页、`desktopWorkbench` 或 `wailsjs` 依赖；桌面页默认用路由懒加载，避免 Web 端请求 `/wailsjs/*`
- 快捷键录入不要只依赖输入框自身事件；`Alt` 这类修饰键要用窗口级 `keydown` 捕获，并支持“仅修饰键预览、非最终值保存”，否则设置页会表现成按下无反应
- 快捷键前端录入映射和 Windows 注册映射必须同步维护；新增按键时要同时检查 `frontend/src/utils/hotkeys.ts` 与 `internal/desktopshell/hotkey_windows.go`，例如 `` ` `` / `Backquote` / `VK_OEM_3`
- Windows 热键注册必须保证“修饰键 + 1 个主键”的约束；单独 `Alt`、`Ctrl`、`Shift` 只能预览，不能保存成最终热键

# FRONTEND
- 样式优先使用 Tailwind，其次 SCSS
- 浏览器storage使用localForage
- 优先flex，必要时grid
- Desktop 不考虑响应式；Web 端考虑响应式，满足桌面端和手机端

# UI
- 紧凑、简洁
- 操作优先用icon替代文字
