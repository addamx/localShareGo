# COMMON
- 不主动执行npm run build
- 不主动创建测试用例
- 不主动创建文档
- 写入AGENTS的内容需精炼且准确

# AGENT
- skill统一只安装到.agents/skills目录下

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
