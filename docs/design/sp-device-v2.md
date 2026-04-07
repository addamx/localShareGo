# v2 Web 设备关联与 Desktop 管理方案

## Summary
- 目标：把当前“基于一次性 Web 链接进入”的能力升级为“可关联设备、可本地恢复、可按设备失效、可申请新关联、可延期”的完整方案。
- 用户侧统一使用“设备关联”表述，不在 UI 中暴露 `token` 概念。
- 明确拆分两类凭证：
  - `待关联入口链接`：Desktop 左侧展示的二维码/复制链接，只用于发起一次新的设备关联。
  - `已关联设备凭证`：Web 浏览器在完成关联后保存到本地、后续用于自动恢复和延期的凭证。
- 失效通知必须按 `deviceId` 定向投递，不允许沿用当前广播式刷新事件模型。
- Desktop 右侧改为“已关联设备列表”，不是在线列表；在线状态仅作为列表附加信息，用于展示和操作反馈。
- 未授权的 `/web` 页面允许调用受限匿名接口，只能发起关联申请或查询申请状态，不能复用已授权接口。

## 产品行为
### 1. Desktop 关联入口
- Desktop 入口文案统一改为“关联设备”。
- 侧边抽屉改为两栏：
  - 左栏：当前 `待关联入口链接`、二维码、复制、浏览器打开、切换 host、刷新入口。
  - 右栏：`已关联设备列表`，每项显示：
    - 名称
    - 最近 IP
    - 在线状态
    - 删除图标
- 左栏“刷新”只刷新 `待关联入口链接`，不影响任何已关联设备。

### 2. Web 首次进入与本地恢复
- 当用户打开 `/web?token=...`：
  - 该链接只用于发起一次新的设备关联。
  - Web 成功完成关联后，服务端返回新的 `已关联设备凭证`。
  - 浏览器把该凭证保存到本地。
  - 当前 URL 用 `replace` 更新为携带“已关联设备凭证”的地址。
- 当用户打开 `/web` 且 URL 不带凭证：
  - 优先读取本地保存的 `已关联设备凭证`。
  - 先向服务端校验该凭证是否有效。
  - 若有效：当前 URL 用 `replace` 补成带凭证地址，再正常进入页面。
  - 若无效：清空本地凭证，并显示“关联已失效”。
  - 若本地无凭证：显示“无可用关联”。
- 同一个 `待关联入口链接` 只能成功关联一次(token 只能和1个deviceId绑定使用，换了浏览器deviceId变了就无法关联)：
  - 一旦某个浏览器完成关联，原始入口链接立即失效。
  - 其他浏览器再用同一个入口链接访问时，必须得到失效结果，不能复用。

### 3. 申请新关联
- Web 在“无可用关联”或“关联已失效”状态下显示“申请关联”按钮。
- 点击后：
  - Web 通过匿名接口提交关联申请，请求体只包含最小必要信息，如本地 `deviceId` 和浏览器生成的设备名称。
  - Desktop 若当前隐藏，则自动显示主窗口。
  - Desktop 弹出审批框，只支持“同意 / 拒绝”。
- 审批结果：
  - 同意：服务端为该 `deviceId` 生成新的 `已关联设备凭证`，Web 轮询或等待结果后写入本地，并进入正常页面。
  - 拒绝：Web 停留在申请失败状态，显示“关联请求已被拒绝”。
- 本版不做匹配码，不做手动输入码流程。

### 4. 已关联设备失效与移除
- Desktop 右侧列表支持删除某个已关联设备。
- 删除后：
  - 该设备对应的 `已关联设备凭证` 立即失效。
  - 若该设备当前在线，服务端向该 `deviceId` 定向发送“关联已失效”事件。
  - Web 收到事件后，立即清空本地凭证并切换到“关联已失效”状态。
- 若该设备当前离线：
  - 不需要即时通知。
  - 下次该浏览器再打开 `/web` 时，本地凭证校验失败并显示“关联已失效”。

### 5. 关联延期
- Web 顶部倒计时文案改为“设备关联剩余时间”。
- 仅当当前 `已关联设备凭证` 的剩余有效期小于等于总有效期一半时，显示“延期”按钮。
- 点击“延期”后：
  - 只延长当前设备对应的 `已关联设备凭证`。
  - 不刷新 Desktop 左栏 `待关联入口链接`。
  - 不影响其他已关联设备。
- 延期成功后，当前设备的剩余时间恢复为新的完整 TTL。

## Key Changes
### 1. 授权模型拆分
- 现有单一 `session/token` 模型升级为两段：
  - `待关联入口`：Desktop 持有，状态为 `pending-entry`，只用于新设备关联。
  - `已关联设备凭证`：绑定到某个 `deviceId`，状态为 `active-device`，用于已关联 Web 的访问、恢复和延期。
- `待关联入口链接` 与 `已关联设备凭证` 必须是不同生命周期对象，不能共享“刷新/延期/删除”语义。
- Desktop 左栏继续只依赖 `待关联入口链接`。
- Web 正常访问 API、SSE、文件下载时只依赖 `已关联设备凭证`。

### 2. Session / Device 数据模型
- `SessionRecord` 扩展为可表达两类会话：
  - `kind`: `entry | device`
  - `deviceId`
  - `deviceName`
  - `lastKnownIP`
  - `revokedAt`
- `DeviceRecord` 扩展为“已关联设备”持久化记录，至少包含：
  - `id`
  - `name`
  - `lastKnownIP`
  - `lastSeenAt`
  - `linkedAt`
  - `updatedAt`
- Desktop 右侧列表的数据来源应为持久化 `DeviceRecord + 当前有效 device session`，不能只取 `presence` 在线列表。

### 3. 在线状态与 IP
- `presence.Registry` 继续负责在线状态，不负责“是否已关联”的持久化真相。
- Web 注册与心跳时，服务端从 HTTP 请求中提取远端地址并更新：
  - `presence.Device.lastKnownIP`
  - 对应 `DeviceRecord.lastKnownIP`
  - 对应 `SessionRecord.lastKnownIP`
- Desktop 右侧列表展示的在线状态来自：
  - `已关联设备` 是否存在于当前 `presence`
- Desktop 右侧列表展示的 IP 优先使用最近一次服务端观察到的 IP。

### 4. 定向失效通知
- 现有 SSE broker 不能直接复用为“广播后前端自行判断”的模式。
- 新增“按 `deviceId` 定向投递”的事件能力：
  - 服务端为每个 Web 连接记录其当前绑定的 `deviceId`
  - 发送 `revoked` 事件时，只推给目标设备对应的连接
- `revoked` 事件负载最少包含：
  - `deviceId`
  - `sessionId`
  - `reason`
  - `ts`
- Web 收到 `revoked` 后不再继续心跳、SSE、数据请求，而是立即转入失效页。

### 5. 匿名接口边界
- 未授权 `/web` 页面允许访问的接口仅限：
  - `POST /api/v1/pair-requests`
  - `GET /api/v1/pair-requests/{id}`
  - 可选：`DELETE /api/v1/pair-requests/{id}` 取消申请
- 匿名接口禁止访问以下能力：
  - 会话详情
  - 设备注册/心跳
  - 剪贴板列表、同步、文件下载
  - SSE 主事件流
- 一旦配对申请被批准，Web 拿到新的 `已关联设备凭证` 后，才允许进入现有已授权链路。

## Public APIs / Types
### 1. Web API
- 新增 `POST /api/v1/session/activate-entry`
  - 入参：`entryToken`、`deviceId`、`deviceName`
  - 行为：消耗一个 `待关联入口链接`，返回新的 `已关联设备凭证`
- 新增 `POST /api/v1/session/renew`
  - 入参：当前 `已关联设备凭证`
  - 行为：仅为当前设备延期；若剩余时间大于半程，返回不可延期
- 新增 `POST /api/v1/pair-requests`
  - 匿名接口
  - 入参：`deviceId`、`deviceName`
  - 返回：`requestId`
- 新增 `GET /api/v1/pair-requests/{id}`
  - 匿名接口
  - 返回申请状态：`pending | approved | rejected | expired`
  - `approved` 时同时返回新的 `已关联设备凭证`
- 新增 `DELETE /api/v1/web-devices/{deviceId}`
  - Desktop/Wails 调用
  - 行为：吊销该设备关联并触发定向失效通知

### 2. Wails 暴露方法
- 新增 `ListLinkedWebDevices()`
  - 返回已关联设备列表，包含在线状态与最近 IP
- 新增 `RemoveLinkedWebDevice(deviceId string)`
  - 删除已关联设备并使其凭证失效
- 新增 `ListPairRequests()`
  - 返回当前待审批的关联申请
- 新增 `ApprovePairRequest(requestId string)`
  - 审批通过并生成设备凭证
- 新增 `RejectPairRequest(requestId string)`
  - 拒绝申请

### 3. 类型调整
- 扩展后端 `SessionResponse`：
  - `sessionId`
  - `sessionStatus`
  - `tokenTtlMinutes`
  - `rotationEnabled`
  - `deviceId`
- 新增前后端共享类型：
  - `LinkedWebDevice`
  - `PairRequestSummary`
  - `RevokedEvent`
- `OnlineDevice` 保留给现有剪贴板同步菜单，不直接作为右侧“已关联设备列表”类型复用。

## 前端改动
### 1. Web
- 新增本地存储键：
  - `localsharego:web:device-credential`
- Web 启动顺序改为：
  - 解析 URL 中是否有凭证
  - 若有且是入口链接，执行 `activate-entry`
  - 若无入口链接，尝试本地 `device-credential`
  - 本地凭证有效则 `replace`
  - 无效则清空并进入失效页
- `useWebClipboardItems` 中当前 `missing / invalid` 状态要拆成面向产品的状态：
  - `no-link`
  - `expired`
  - `requesting`
  - `waiting-approval`
  - `ready`
- Web 失效页和无关联页增加“申请关联”入口。
- Web 正常页增加“延期”按钮与延期中状态。

### 2. Desktop
- `DesktopWebDrawer` 改为左右布局。
- 左栏保留现有 host 选择、二维码、复制、浏览器打开、刷新入口。
- 右栏新增已关联设备列表：
  - 支持在线/离线状态展示
  - 支持最近 IP 展示
  - 支持删除
- Desktop 增加关联审批弹窗：
  - 有新请求时自动拉起
  - 支持同意、拒绝
  - 审批后右侧列表与左栏状态同步刷新

## Test Plan
- 使用 `待关联入口链接` 首次打开 `/web?token=...`：
  - 成功生成 `已关联设备凭证`
  - 浏览器本地保存该凭证
  - URL 被 `replace` 为新的设备凭证地址
  - 原入口链接在第二个浏览器中不可再次成功使用
- 浏览器已完成关联后再次直接打开 `/web`：
  - 若本地凭证有效，能自动恢复并进入正常页
  - 若本地凭证无效，能清空本地状态并显示“关联已失效”
  - 若本地无凭证，显示“无可用关联”
- Desktop 删除一个已关联设备：
  - 该设备在线时立即收到定向 `revoked` 事件并切换到失效页
  - 其他设备不受影响
  - 该设备离线时不会误伤其他在线设备，下次打开时校验失败
- Web 点击“申请关联”：
  - 未授权状态下只能命中匿名接口
  - Desktop 隐藏时会自动显示并弹审批框
  - 审批通过后当前浏览器立即进入可用态
  - 审批拒绝后保持在失败/被拒绝态
- Web 剩余时间超过半程时：
  - 不显示“延期”
- Web 剩余时间低于或等于半程时：
  - 显示“延期”
  - 延期成功后恢复完整 TTL
  - 不影响 Desktop 左侧入口链接
- Desktop 右侧列表：
  - 展示所有已关联设备，而不是只展示在线设备
  - 在线/离线状态与最近 IP 正确更新
  - 删除后列表立即移除对应设备

## Assumptions
- `待关联入口链接` 仍沿用当前 Desktop 侧“刷新生成新入口”的产品心智。
- `已关联设备凭证` 仍可继续通过 URL 查询参数传输，首版不改成 Cookie。
- 本版不做匹配码，不做扫码确认二次校验，不做多端共享一个设备凭证。
- 首版允许一个浏览器实例只持有一个当前有效设备凭证。
- 匿名接口只承载配对申请，不承担任何剪贴板读取或设备在线注册职责。
