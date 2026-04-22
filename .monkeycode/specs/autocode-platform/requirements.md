# Requirements Document

## Introduction

本系统用于在低资源 VPS 上运行自用优先的 AI 自动化中枢。平台需要把多模型接入、任务分析、工具编排、写改测闭环、部署修复、运维治理和交付分发统一到一个轻量控制台中，同时为更高配置环境保留无重构升级空间。

## Glossary

- **平台**: 指本 AI 自动化中枢
- **运行单**: 指一次端到端任务执行记录
- **阶段**: 指闭环中的一个执行步骤或检查节点
- **模型兼容层**: 指屏蔽不同模型平台差异的统一抽象层
- **工具原子能力**: 指可被单独编排、复用和观测的最小能力单元
- **硬件画像**: 指系统对 CPU、内存、磁盘和网络约束的运行时判断结果

## Requirements

### Requirement 1

**User Story:** AS 自用开发者, I want 统一管理多平台模型, so that 我可以在一个入口下接入、筛选和切换不同模型。

#### Acceptance Criteria

1. WHEN 平台启动, 平台 SHALL 加载内置模型目录并展示模型名称、供应商名称、官网链接和可用状态。
2. WHEN 用户查看模型目录, 平台 SHALL 支持按供应商、区域、能力标签、接入方式和排序偏好筛选模型。
3. WHEN 用户配置模型凭据或地址, 平台 SHALL 将模型状态更新为已就绪、未就绪或仅展示。
4. IF 某模型未满足调用条件, 平台 SHALL 阻止该模型参与执行并给出明确原因。

### Requirement 2

**User Story:** AS 自用开发者, I want 平台使用统一的模型兼容层, so that 不同 API 模型和本地模型都可以复用同一套编排逻辑。

#### Acceptance Criteria

1. WHEN 平台注册模型, 平台 SHALL 将其归一为统一的能力描述、上下文限制、成本属性和可用工具模式。
2. WHEN 运行单选择执行模型, 平台 SHALL 通过统一接口发起推理、流式输出和失败重试。
3. IF 某模型不支持指定能力, 平台 SHALL 自动降级到兼容模式或建议替代模型。
4. WHEN 本地模型被启用, 平台 SHALL 将其与远程 API 模型分开治理与展示。

### Requirement 3

**User Story:** AS 自用开发者, I want 平台基于硬件画像自适应运行, so that 它既能稳定跑在 1C1G VPS 上，也能在更高规格机器上自动放大能力。

#### Acceptance Criteria

1. WHEN 平台启动或配置变更, 平台 SHALL 生成当前主机的硬件画像和运行档位。
2. WHEN 调度器准备执行阶段, 平台 SHALL 根据硬件画像限制并发度、模型策略、工具负载和缓存策略。
3. IF 资源不足, 平台 SHALL 优先保留核心闭环能力，并暂停或降级高负载能力。
4. WHEN 检测到更高规格环境, 平台 SHALL 自动开放更高档位能力而不要求重构配置。

### Requirement 4

**User Story:** AS 自用开发者, I want 用选项化方式配置闭环阶段, so that 我可以按任务目标裁剪完整流程而不是手写复杂编排。

#### Acceptance Criteria

1. WHEN 用户创建运行单, 平台 SHALL 允许用户启用、禁用或重排分析、规划、实现、验证、部署和修复阶段。
2. WHEN 用户配置阶段, 平台 SHALL 提供阶段说明、风险提示和推荐默认值。
3. WHILE 运行单执行中, 平台 SHALL 显示当前阶段、后续阶段和每个阶段的摘要结果。
4. IF 某阶段失败, 平台 SHALL 允许用户重试、跳过、转人工确认或进入修复链路。

### Requirement 5

**User Story:** AS 自用开发者, I want 平台以 12 步闭环作为统一骨架, so that 代码、文档、运维和交付任务都能复用同一套执行框架。

#### Acceptance Criteria

1. WHEN 平台分析任务, 平台 SHALL 将运行单映射到统一的闭环骨架，而不是为每类任务维护完全独立流程。
2. WHEN 某任务不需要完整闭环, 平台 SHALL 允许裁剪不必要步骤并保留状态可追踪性。
3. WHEN 阶段结束, 平台 SHALL 记录输入、输出、决策和下一步建议。
4. IF 运行单被暂停或补充需求, 平台 SHALL 能从闭环中的任意安全节点恢复。

### Requirement 6

**User Story:** AS 自用开发者, I want 平台自动去重和自动调度, so that 重复任务不会浪费资源且执行顺序始终可控。

#### Acceptance Criteria

1. WHEN 用户提交新运行单, 平台 SHALL 基于任务摘要、目标路径、模板签名和阶段配置生成去重键。
2. IF 存在相同去重键且状态为排队或运行中, 平台 SHALL 阻止创建重复运行单并返回已有运行单信息。
3. WHEN 调度器扫描队列, 平台 SHALL 同时考虑优先级、资源预算、硬件档位和阶段负载。
4. WHILE 平台处于轻量档位, 平台 SHALL 默认采用保守并发和串行优先策略。

### Requirement 7

**User Story:** AS 自用开发者, I want 平台把工具能力原子化并去重整合, so that 我可以按能力调用而不是在多个重复入口之间来回切换。

#### Acceptance Criteria

1. WHEN 平台加载工具目录, 平台 SHALL 以能力类别、负载等级和适用阶段组织工具原子能力。
2. WHEN 多个工具提供相近能力, 平台 SHALL 支持定义首选入口、候补入口和禁用入口。
3. WHEN 用户单独调用某能力, 平台 SHALL 允许脱离完整运行单直接执行该原子能力。
4. IF 某原子能力未启用或依赖不足, 平台 SHALL 在界面中显示原因和启用方式。

### Requirement 8

**User Story:** AS 自用开发者, I want 在流程中暂停并补充需求, so that 我可以中途调整目标而不必推倒重来。

#### Acceptance Criteria

1. WHEN 用户请求暂停运行单, 平台 SHALL 将运行单切换到暂停状态并停止后续阶段调度。
2. WHILE 运行单处于暂停状态, 平台 SHALL 允许用户追加需求、补充约束和修改模板参数。
3. WHEN 用户恢复运行单, 平台 SHALL 基于最新需求继续后续阶段。
4. IF 补充需求影响已完成阶段, 平台 SHALL 标记受影响阶段为待重新执行。

### Requirement 9

**User Story:** AS 自用开发者, I want 全面的验证闭环, so that 每次执行都能形成可信结果而不是只产出表面内容。

#### Acceptance Criteria

1. WHEN 运行单进入验证链路, 平台 SHALL 支持按需执行静态检查、单元测试、集成测试、端到端测试和人工复核。
2. WHEN 任一验证步骤完成, 平台 SHALL 记录名称、状态、耗时、摘要和证据链接。
3. IF 验证失败, 平台 SHALL 将失败摘要传递给修复阶段并保留原始上下文。
4. WHEN 验证全部通过, 平台 SHALL 标记该运行单为验证通过。

### Requirement 10

**User Story:** AS 自用开发者, I want 平台在部署或测试失败时自动修复, so that 我可以减少手动排查成本。

#### Acceptance Criteria

1. WHEN 测试、部署或运维检查失败, 平台 SHALL 生成结构化失败摘要。
2. WHEN 用户启用自动修复, 平台 SHALL 基于失败摘要创建修复阶段并复用原始上下文。
3. IF 自动修复连续失败达到阈值, 平台 SHALL 暂停运行单并请求人工干预。
4. WHEN 自动修复收敛失败, 平台 SHALL 给出下一步建议而不是静默结束。

### Requirement 11

**User Story:** AS 自用开发者, I want 通过 Web 控制台完成大多数治理动作, so that 部署后不必频繁依赖 SSH。

#### Acceptance Criteria

1. WHEN 用户打开控制台, 平台 SHALL 提供系统总览、运行单管理、模型治理、工具治理和运维建议入口。
2. WHEN 平台检测到资源风险、配置缺失或模型未就绪, 平台 SHALL 在控制台中直接提示处理建议。
3. WHEN 用户执行常见治理动作, 平台 SHALL 优先提供 Web 控制台操作而不是要求手工登录服务器。
4. IF 某动作必须依赖外部环境, 平台 SHALL 明确说明前置条件和影响范围。

### Requirement 12

**User Story:** AS 自用开发者, I want 一个成熟但轻量的中文控制台, so that 我可以像使用产品而不是拼装脚本一样使用平台。

#### Acceptance Criteria

1. WHEN 用户打开控制台, 平台 SHALL 提供中文化、产品化且适配桌面与移动端的布局。
2. WHEN 用户查看运行单详情, 平台 SHALL 展示阶段时间线、摘要、建议、日志和可执行动作。
3. WHEN 用户切换日夜主题, 平台 SHALL 保持核心信息层级和可读性稳定。
4. WHEN 某区域暂无数据, 平台 SHALL 提供自然易懂的空状态与下一步引导。

### Requirement 13

**User Story:** AS 自用开发者, I want 支持轻量产品化交付, so that 平台既能自用，也能在需要时做白标、离线分发和稳定升级。

#### Acceptance Criteria

1. WHEN 平台生成交付包, 平台 SHALL 支持打包品牌资源、配置模板和可选能力清单。
2. WHEN 用户选择离线分发, 平台 SHALL 生成不依赖开发环境的可部署交付形式。
3. WHEN 平台执行升级, 平台 SHALL 区分核心程序、静态资源和插件能力的版本边界。
4. IF 某能力未被启用, 平台 SHALL 避免将其作为常驻运行负担。

### Requirement 14

**User Story:** AS 自用开发者, I want 平台覆盖代码之外的通用自动化任务, so that 文档、配置、研究、发布和运维也能纳入同一个中枢。

#### Acceptance Criteria

1. WHEN 平台分析任务, 平台 SHALL 识别代码、文档、配置、研究、发布、运维和治理等任务类型。
2. WHEN 不同任务类型进入同一闭环骨架, 平台 SHALL 复用统一状态机、摘要结构和观测视图。
3. IF 某任务类型需要专属工具链, 平台 SHALL 通过插件方式接入而不是污染核心流程。
4. WHEN 平台展示任务分类, 平台 SHALL 让用户清楚知道当前运行单属于哪类意图与交付方式。
