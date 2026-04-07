# v1 OCR 接入方案

## Summary
- 首版只做桌面端 OCR，不做 Web 端交互和浏览器侧识别。
- 识别对象限定为“已经有本地路径的图片文件项”，覆盖两条链路：
  - 本机直接复制进剪贴板历史的图片文件
  - 网络同步到桌面后、用户执行接收并落地到本地路径的图片文件
- 识别方式采用 `Go/Wails 主程序 + Windows OCR helper 进程`。
- 识别结果用于：
  - 文件项全文搜索
  - 列表显示 OCR 摘要文本
  - 详情页查看完整识别文本并一键复制
- 失败策略为静默非阻断：OCR 失败不影响文件项入库、展示、接收、复制和同步。

## Key Changes
### 1. OCR 引擎与进程模型
- 新增一个随应用分发的 Windows helper，可执行文件放在 Windows 构建产物旁边，由主程序按命令行调用。
- helper 使用 Windows 系统 OCR：
  - `Windows.Media.Ocr.OcrEngine.TryCreateFromUserProfileLanguages()`
  - 不在应用内做语言切换 UI
  - 默认目标语言预期为用户机器已安装并启用 `zh-CN` / `en-US` 的用户配置语言
- helper 输入固定为本地图片路径，输出固定为 JSON 到 stdout：
  - `status`: `ready | failed | unsupported`
  - `text`: 完整识别文本
  - `preview`: 截断后的摘要文本
- 主程序中的 OCR worker 单并发串行执行，避免同时识别多张图导致 UI 抖动和 CPU 突刺。
- 单次 OCR 调用设置固定超时；超时视为 `failed`，不重试、不通知。

### 2. 后端数据模型与持久化
- 在 `internal/ocr/` 新增 OCR 领域服务，负责：
  - 判断文件项是否可识别
  - 调度 helper
  - 回写 store
- 扩展 `internal/store` 的文件元数据模型，在 `ClipboardFileMeta` 下新增可选 `ocr` 字段：
  - `status`
  - `text`
  - `preview`
  - `updatedAt`
- 不新增迁移框架；沿用当前 JSON store 的向后兼容方式，旧数据没有 `ocr` 字段时按未识别处理。
- `internal/store` 新增显式更新方法，例如 `UpdateClipboardFileOCR(itemID, ocrMeta)`，不要把 OCR 回写混进已有传输状态更新方法。
- 文件项标准化逻辑保持“`content` 仍为空”；OCR 文本不复用 `content`，避免被 `normalizeClipboardItemRecord` 清空。
- 列表搜索逻辑改为同时匹配：
  - `preview`
  - `content`
  - `fileMeta.ocr.text`
- 对同一路径、同一条目重复触发时：
  - 若 OCR 已 `ready` 且本地路径未变化，则跳过
  - 文件接收后本地路径发生变化时，先清空旧 OCR，再重新排队识别

### 3. 触发时机与数据流
- 本机剪贴板图片文件链路：
  - `internal/clipboard/service.go` 在 `tickFile()` 成功保存文件项后，如果是图片且有本地路径，异步投递 OCR 任务
- 桌面接收同步图片链路：
  - `app.go` 的 `ReceiveClipboardFile()` 成功后，如果条目为图片且 `localPath` 已落地，异步投递 OCR 任务
- 首版不在以下场景触发 OCR：
  - `metadata_only`
  - `receiving`
  - 非图片文件
  - Web 端仅浏览/上传但未落地为桌面本地路径的场景
- 应用启动时不做全量历史回补扫描；首版只对新进入或新落地的图片执行 OCR，避免启动时长和额外 I/O。

### 4. 前端类型与桌面 UI
- 扩展前端 `ClipboardFileMeta` 类型，增加 `ocr` 可选字段，与后端保持一致。
- 桌面列表文件卡片改为：
  - 第一行仍显示文件名
  - 第二行优先显示 `ocr.preview`
  - 无 OCR 时退回现有文件说明
- 搜索命中 OCR 文本时，列表仍保持“文件卡片”样式，不改成文本卡片。
- 桌面详情抽屉新增 OCR 区块：
  - `ready` 时展示完整识别文本
  - 提供“复制识别文本”动作，复用现有 `CopyText` 绑定
  - `failed / unsupported / 无结果` 时不弹通知，只显示轻量状态或直接不展示内容
- Web 组件不做 OCR 专属 UI 改动；即使接口里带上 OCR 字段，也不在首版消费。

## Public APIs / Types
- 新增后端类型：
  - `ClipboardOCRMeta`
- 扩展后端类型：
  - `ClipboardFileMeta.ocr`
- 扩展前端类型：
  - `ClipboardFileMeta["ocr"]`
- 新增 store 能力：
  - `UpdateClipboardFileOCR(itemID, ocrMeta)`
- 不新增新的 Wails 对外命令；现有 `ListClipboardItems` / `GetClipboardItem` 返回结构自然带出 OCR 数据即可。

## Test Plan
- 复制本地 PNG/JPG 图片文件到系统剪贴板后：
  - 文件项正常入库
  - OCR 异步从无结果变为 `ready`
  - 搜索图片中文字能命中该文件项
  - 列表显示 OCR 摘要
  - 详情可查看并复制完整识别文本
- 接收一条同步来的图片文件后：
  - 接收流程不受 OCR 影响
  - 文件落地后自动启动 OCR
  - 完成后同样支持搜索、摘要和详情复制
- 复制非图片文件到剪贴板后：
  - 不触发 OCR
  - 现有文件项行为不变
- OCR helper 不可用、系统 OCR 不支持、或图片解码失败时：
  - 条目仍正常保存/接收
  - 不弹错误通知
  - 条目标记为 `failed` 或 `unsupported`
  - 后续操作如激活、同步、删除不受影响
- 用旧版 JSON 数据启动：
  - 没有 `ocr` 字段的数据能正常加载
  - 搜索和列表不会报错

## Assumptions
- 首版只面向 Windows 桌面应用，不做跨平台抽象兑现。
- helper 作为应用内置组件一起分发，不要求用户额外安装 OCR 模块。
- 不新增 OCR 设置页、不新增手动“重新识别”入口、不做历史数据回填任务。
- 不把 OCR 结果同步给其他设备；OCR 结果仅作为当前桌面端本地增强能力。
