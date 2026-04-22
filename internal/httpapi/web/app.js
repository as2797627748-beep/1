const state = {
  templates: [],
  runs: [],
  models: [],
  system: null,
  analysisPreview: null,
  audit: null,
  advice: null,
  localPolicyAdvice: {},
  workflowOptions: [],
  theme: 'night',
  templateSource: '',
  templatePinned: false,
  recentTemplateIds: [],
  runDeploymentSignatures: {},
  selectedRunId: null,
  selectedModelId: null,
  modelFilters: {
    query: '',
    provider: 'all',
    region: 'all',
    readiness: 'all',
  },
  localModelFilters: {
    query: '',
    tier: 'all',
    install: 'all',
    compat: 'all',
    deploy: 'all',
    focus: 'all',
    capability: 'all',
    sort: 'low-friction',
  },
  toolPolicyApplied: false,
  stagePolicyApplied: false,
  tools: [
    { name: 'workspace', enabled: true },
    { name: 'terminal', enabled: true },
    { name: 'tests', enabled: true },
    { name: 'deploy', enabled: true },
    { name: 'logs', enabled: true },
  ],
  dev: {
    currentDir: '.',
    files: [],
    currentFilePath: '',
    currentFileContent: '',
    terminalOutput: '',
    terminalSessions: [],
    previewPort: '8080',
  },
  testMode: 'template',
  selectedTests: [],
};

const views = ['dashboard', 'runs', 'models', 'templates', 'control'];
const stageOptions = ['intent', 'context', 'plan', 'resource', 'model', 'tool', 'implement', 'result', 'test', 'deploy', 'repair', 'finalize'];
let deploymentTimelineFocusTimer = null;
let deploymentTimelineFocusedItem = null;

const runStatusLabels = {
  queued: '待启动',
  running: '进行中',
  paused: '已暂停',
  failed: '需关注',
  completed: '已完成',
};

const stageLabels = {
  intent: '任务识别',
  context: '上下文汇总',
  plan: '方案规划',
  resource: '资源收束',
  model: '模型选择',
  tool: '工具装配',
  implement: '执行生成',
  result: '结果整理',
  test: '验证检查',
  deploy: '部署交付',
  repair: '修复回路',
  finalize: '总结沉淀',
};

const toolLabels = {
  workspace: '工作区协同',
  terminal: '受控终端',
  tests: '验证链路',
  deploy: '稳定部署',
  logs: '运行观测',
  analysis: '意图解析',
  research: '研究整理',
  office: '办公协作',
  assets: '内容资产',
  lint: '规范校验',
  format: '结构整理',
  build: '构建产物',
};

const toolDescriptions = {
  workspace: '文件与补丁',
  terminal: '命令与构建',
  tests: '测试与回归',
  deploy: '部署与同步',
  logs: '日志与动态',
};

const localTierOrder = ['1.5B', '2.7B', '3B', '3.8B', '4B', '7B', '8B', '14B', '8x7B', '32B', '34B', '70B', '72B', '180B', '405B'];

const localInstallStateLabels = {
  inactive: '待下载',
  queued: '已排队',
  downloading: '下载中',
  configuring: '配置中',
  disabled: '已停用',
  ready: '已就绪',
  removed: '已移除',
};

const localDeployLabels = {
  cpu: 'CPU 优先',
  gpu: 'GPU 友好',
  node: '独立节点',
  cluster: '集群级',
};

const localCapabilityLabels = {
  'low-memory': '低内存',
  reasoning: '推理优先',
  'long-context': '长上下文',
  chinese: '中文优先',
  multilingual: '多语优先',
};

const capabilityModeLabels = {
  resident: '常驻',
  'on-demand': '按需',
  'display-only': '仅展示',
  'external-managed': '外部托管',
  'externally-managed': '外部托管',
  builtin: '内建常备',
  optional: '按需挂载',
};

function capabilityModeClass(mode) {
  if (mode === 'resident') return 'completed';
  if (mode === 'on-demand') return 'running';
  if (mode === 'display-only') return 'paused';
  if (mode === 'external-managed') return 'queued';
  if (mode === 'externally-managed') return 'queued';
  if (mode === 'builtin') return 'completed';
  if (mode === 'optional') return 'running';
  return '';
}

const loadTierLabels = {
  light: '轻动作',
  medium: '中负载',
  heavy: '重动作',
};

const localPolicyModeLabels = {
  'display-only': '仅展示',
  'on-demand': '按需启用',
  'externally-managed': '外部托管',
};

const runtimeProfileLabels = {
  'home-lite': '轻载模式',
  'balanced-hybrid': '均衡模式',
  'adaptive-performance': '性能模式',
};

const interfaceModeLabels = {
  atelier: '工作台',
  'mission-control': '任务中枢',
  pocket: '轻量入口',
};

const schedulerModeLabels = {
  'serial-lightweight': '轻量串行调度',
};

const deployModeLabels = {
  'idempotent-app-only': '应用级稳定部署',
};

const checkpointStatusLabels = {
  pending: '待开始',
  running: '进行中',
  in_progress: '进行中',
  completed: '已完成',
  failed: '需关注',
  paused: '已暂停',
};

const validationDepthLabels = {
  essential: '基础闭环',
  standard: '标准闭环',
  extended: '扩展闭环',
};

const analysisIntentLabels = {
  build: '构建',
  fix: '修复',
  organize: '整理',
  'build-deploy': '构建并交付',
  unknown: '待识别',
};

const projectKindLabels = {
  web: 'Web 项目',
  backend: '后端服务',
  game: '游戏项目',
  app: '应用项目',
  software: '软件工程',
  office: '办公任务',
  research: '研究任务',
  daily: '日常事务',
  'docs-content': '文档内容',
  general: '通用任务',
};

const toolCategoryLabels = {
  core: '核心能力',
  knowledge: '知识与研究',
  productivity: '办公与协作',
  creative: '内容与资产',
  quality: '质量保障',
  release: '构建与交付',
  planner: '分析与规划',
  code: '代码与工程',
  office: '办公与协作',
  research: '研究与整理',
  assets: '内容与资产',
  media: '多媒体处理',
  daily: '日常自动化',
  ops: '交付与运维',
  general: '通用能力',
};

const testLayerLabels = {
  quality: '质量',
  unit: '单元',
  integration: '集成',
  e2e: '端到端',
  ops: '运维',
};

const adviceModeLabels = {
  'ops-review': '升级建议',
  'external-hosting': '外部托管建议',
};

const queueAgentLabels = {
  '需求解析Agent': '需求解析',
  '调度策略Agent': '调度策略',
  '模型治理Agent': '模型治理',
  '工具治理Agent': '工具治理',
  '代码生成Agent': '执行产出',
  '结果整理Agent': '结果整理',
  '测试验证Agent': '测试验证',
  '终端执行Agent': '终端执行',
  '错误修复Agent': '修复处理',
  '总结沉淀Agent': '总结沉淀',
};

const regionLabels = {
  global: '海外',
  domestic: '国内',
};

const localPolicySections = [
  {
    key: 'on-demand',
    title: '按需启用',
    description: '这组候选适合保留在当前主机，按需完成下载、启用与停用，无需长期常驻。',
  },
  {
    key: 'externally-managed',
    title: '外部托管',
    description: '这组候选更适合部署到独立节点、GPU 主机或集群，由控制台保留统一管理入口。',
  },
  {
    key: 'display-only',
    title: '仅展示',
    description: '这组候选用于保留目录信息与能力参考，避免在当前机器上直接占用过多资源。',
  },
];

const hostingModeLabels = {
  cpu: 'CPU 独立节点',
  gpu: 'GPU 单机',
  node: 'CPU 独立节点',
  cluster: '集群级托管',
};

function localPolicyModeClass(mode) {
  if (mode === 'on-demand') return 'completed';
  if (mode === 'externally-managed') return 'queued';
  return 'paused';
}

function hostingModeLabel(mode) {
  return hostingModeLabels[mode] || hostingModeLabels.cpu;
}

function optimizationPriorityClass(priority) {
  if (priority === 'high') return 'failed';
  if (priority === 'medium') return 'running';
  return 'paused';
}

function numericTierValue(tier) {
  const normalized = String(tier || '').toLowerCase();
  if (normalized.includes('x')) {
    const parts = normalized.split('x');
    const left = Number.parseFloat(parts[0]) || 0;
    const right = Number.parseFloat(parts[1]) || 0;
    return left * right;
  }
  return Number.parseFloat(normalized.replace(/[^\d.]/g, '')) || 0;
}

function localDeploymentMeta(pack) {
  const tierValue = numericTierValue(pack.sizeTier);
  if (tierValue >= 180) {
    return { key: 'cluster', label: hostingModeLabel('cluster') };
  }
  if (tierValue >= 70) {
    return { key: 'node', label: hostingModeLabel('node') };
  }
  if (pack.variant.includes('AWQ') || pack.variant.includes('EXL2')) {
    return { key: 'gpu', label: hostingModeLabel('gpu') };
  }
  return { key: 'cpu', label: hostingModeLabel('cpu') };
}

function localInstallStateClass(stateName) {
  if (stateName === 'ready') return 'completed';
  if (stateName === 'removed') return 'failed';
  if (stateName === 'queued' || stateName === 'downloading' || stateName === 'configuring') return 'running';
  return 'paused';
}

function localInstallActionLabel(pack) {
  if (pack.installState === 'queued') return '等待部署队列';
  if (pack.installState === 'downloading') return '正在下载';
  if (pack.installState === 'configuring') return '正在配置';
  if (pack.enabled && pack.downloaded) return '停用';
  if (pack.policyMode === 'display-only') return '仅供查看';
  if (pack.policyMode === 'externally-managed') return '查看托管方案';
  if (pack.downloaded) return pack.enabled ? '停用' : '启用';
  return '下载配置并启用';
}

function localInstallBusy(pack) {
  return pack.installState === 'queued' || pack.installState === 'downloading' || pack.installState === 'configuring';
}

function canToggleLocalModel(pack) {
  if (localInstallBusy(pack)) return false;
  if (pack.enabled && pack.downloaded) return true;
  return pack.policyMode === 'on-demand' && pack.allowed;
}

function compareLocalModelLowFriction(left, right) {
  if (left.reviewScore !== right.reviewScore) {
    return left.reviewScore - right.reviewScore;
  }
  if (left.filterScore !== right.filterScore) {
    return left.filterScore - right.filterScore;
  }
  if (left.alignmentScore !== right.alignmentScore) {
    return left.alignmentScore - right.alignmentScore;
  }
  if (left.recommended !== right.recommended) {
    return left.recommended ? -1 : 1;
  }
  return left.name.localeCompare(right.name);
}

function isFlagshipLowFrictionPack(pack) {
  return pack.reviewScore <= 8 && pack.filterScore <= 8 && pack.alignmentScore <= 9;
}

function getLowFrictionRankedLocalModelPacks(packs) {
  return [...packs].sort(compareLocalModelLowFriction);
}

function localCapabilityTags(pack) {
  const tags = [];
  const tierValue = numericTierValue(pack.sizeTier);
  const providerText = `${pack.provider} ${pack.modelName} ${pack.name}`.toLowerCase();
  const text = `${pack.description} ${pack.policyHint} ${pack.runtimeHint} ${pack.systemRequirement}`.toLowerCase();

  if (tierValue > 0 && tierValue <= 4) {
    tags.push('low-memory');
  }
  if (providerText.includes('deepseek') || providerText.includes('phi') || text.includes('推理')) {
    tags.push('reasoning');
  }
  if (providerText.includes('command') || providerText.includes('mixtral') || text.includes('长上下文') || text.includes('上下文')) {
    tags.push('long-context');
  }
  if (providerText.includes('qwen')) {
    tags.push('chinese');
  }
  if (
    providerText.includes('yi') ||
    providerText.includes('mistral') ||
    providerText.includes('mixtral') ||
    providerText.includes('zephyr') ||
    providerText.includes('gemma') ||
    text.includes('多语')
  ) {
    tags.push('multilingual');
  }

  return [...new Set(tags)];
}

function dominantHostingTarget(packs = []) {
  const counts = packs.reduce((acc, pack) => {
    const target = localDeploymentMeta(pack).key;
    acc[target] = (acc[target] || 0) + 1;
    return acc;
  }, {});
  const priority = ['cluster', 'node', 'gpu', 'cpu'];
  return priority.find((key) => counts[key]) || 'cpu';
}

function summarizePolicyReasons(packs = []) {
  return Object.entries(packs.reduce((acc, pack) => {
    const reason = pack.policyDecision?.reason || 'runtime-tier';
    acc[reason] = (acc[reason] || 0) + 1;
    return acc;
  }, {}))
    .sort((left, right) => right[1] - left[1])
    .slice(0, 3)
    .map(([reason, count]) => `${policyDecisionReasonLabel(reason)} ${count}`);
}

function groupRecommendedBadges(packs = []) {
  return packs
    .filter((pack) => pack.recommended)
    .slice(0, 3)
    .map((pack) => `<span class="badge completed">${pack.name}</span>`)
    .join('');
}

function rankPolicySectionCandidates(groupKey, packs = []) {
  const ranked = [...packs];
  if (groupKey === 'on-demand') {
    ranked.sort((left, right) => {
      if (left.recommended !== right.recommended) {
        return left.recommended ? -1 : 1;
      }
      if (left.installState === 'ready' || right.installState === 'ready') {
        if (left.installState !== right.installState) {
          return left.installState === 'ready' ? -1 : 1;
        }
      }
      return compareLocalModelLowFriction(left, right);
    });
    return ranked;
  }
  return ranked;
}

function policyGovernanceSummary(groupKey, packs = []) {
  if (groupKey === 'on-demand') {
    const ready = packs.filter((pack) => pack.installState === 'ready').length;
    return `建议优先选择当前推荐候选，先使用已就绪的 ${ready} 个候选承接任务，再按需补充下载。`;
  }
  if (groupKey === 'externally-managed') {
    return '建议先生成分组托管建议，统一规划节点、监控与容量，再确定采用独立节点、GPU 主机或集群方案。';
  }
  const reasons = summarizePolicyReasons(packs);
  return `建议保留目录展示，并结合主要限制原因${reasons.length ? `：${reasons.join(' / ')}` : ''}评估后续方案，当前不建议直接启用。`;
}

function syncLocalModelFilterControls() {
  const mapping = {
    '#local-tier-filter': 'tier',
    '#local-install-filter': 'install',
    '#local-compat-filter': 'compat',
    '#local-deploy-filter': 'deploy',
    '#local-focus-filter': 'focus',
    '#local-capability-filter': 'capability',
    '#local-sort-filter': 'sort',
  };
  Object.entries(mapping).forEach(([selector, key]) => {
    const element = q(selector);
    if (element) {
      element.value = state.localModelFilters[key];
    }
  });
  const search = q('#local-model-search');
  if (search) {
    search.value = state.localModelFilters.query;
  }
}

function resetLocalModelFilters() {
  state.localModelFilters = {
    query: '',
    tier: 'all',
    install: 'all',
    compat: 'all',
    deploy: 'all',
    focus: 'all',
    capability: 'all',
    sort: 'low-friction',
  };
  syncLocalModelFilterControls();
}

function localFilterSummaryBadges() {
  const { query, tier, install, compat, deploy, focus, capability, sort } = state.localModelFilters;
  const items = [];
  if (query) items.push(`搜索 ${query}`);
  if (tier !== 'all') items.push(`体量 ${tier}`);
  if (install !== 'all') items.push(`状态 ${localInstallStateLabels[install] || install}`);
  if (compat === 'allowed') items.push('机器适配 按需启用');
  if (compat === 'recommended') items.push('机器适配 当前推荐');
  if (deploy !== 'all') items.push(`部署 ${hostingModeLabel(deploy)}`);
  if (focus !== 'all') items.push(`焦点 ${focus}`);
  if (capability !== 'all') items.push(`标签 ${localCapabilityLabels[capability] || capability}`);
  if (sort !== 'low-friction') items.push(`排序 ${sort}`);
  return items;
}

function localFilterSignature() {
  return JSON.stringify(state.localModelFilters);
}

function primaryGovernanceDirection(stats) {
  const candidates = [
    { key: 'on-demand', count: stats.onDemand, label: '按需启用优先' },
    { key: 'externally-managed', count: stats.externallyManaged, label: '外部托管优先' },
    { key: 'display-only', count: stats.displayOnly, label: '仅展示优先' },
  ].sort((left, right) => right.count - left.count);
  return candidates[0]?.label || '按需启用优先';
}

function buildExternalHostingAdviceForGroup(packs = []) {
  const target = dominantHostingTarget(packs);
  const titles = packs.slice(0, 3).map((pack) => pack.name).join(' / ');
  return fetchJSON('/api/system/advice', {
    method: 'POST',
    body: JSON.stringify({
      mode: 'external-hosting',
      goal: `请为当前外部托管分组生成统一建议，重点覆盖 ${titles || '当前筛选候选'} 等模型的 ${hostingModeLabel(target)} 方案`,
      target,
    }),
  });
}

function summarizeAdviceForGroup(advice) {
  if (!advice) {
    return null;
  }
  const lead = advice.suggestions?.find((item) => item && !item.startsWith('当前规划目标:')) || advice.summary;
  return {
    hostingMode: advice.hostingMode || '',
    summary: lead || advice.summary || '托管建议已更新',
    filterSignature: localFilterSignature(),
  };
}

function statusLabel(status) {
  return runStatusLabels[status] || status;
}

function deploymentStatusClass(status) {
  if (status === 'completed') return 'completed';
  if (status === 'running') return 'running';
  if (status === 'failed') return 'failed';
  return 'paused';
}

function deploymentStatusLabel(status) {
  return {
    completed: '已完成',
    running: '进行中',
    failed: '需关注',
    paused: '已暂停',
    pending: '待开始',
  }[status] || status || '待开始';
}

function deploymentTargetLabel(target) {
  return {
    'configured-vps': '已配置 VPS',
  }[target] || target || '待配置';
}

function deploymentModeLabel(mode) {
  return {
    remote: '应用级稳定部署',
    rollback: '版本回退',
    revalidate: '复验',
  }[mode] || mode || '待记录';
}

function deploymentSourceMeta(record) {
  if (record?.mode === 'revalidate') {
    return { label: '复验', tone: 'running' };
  }
  if (record?.mode === 'rollback') {
    return { label: '版本回退', tone: 'running' };
  }
  if (record?.mode === 'remote') {
    return { label: '稳定部署', tone: 'completed' };
  }
  return { label: '部署记录', tone: 'paused' };
}

function deriveDeploymentTargetProfile(target) {
  if (target === 'configured-vps') {
    return {
      title: '已配置 VPS',
      summary: '已接入可用 VPS，可继续发布、回退，并查看线上效果。',
      items: [
        { label: '环境接入', value: '已接入', status: 'completed' },
        { label: '发布方式', value: '应用级稳定部署', status: 'completed' },
        { label: '重点关注', value: '预览与服务状态', status: 'running' },
      ],
    };
  }
  return {
    title: '环境待配置',
    summary: '目标环境信息尚未补齐，请先确认主机、目录与预览入口。',
    items: [
      { label: '环境接入', value: '待配置', status: 'paused' },
      { label: '发布方式', value: '待补充', status: 'paused' },
      { label: '重点关注', value: '目标环境信息', status: 'paused' },
    ],
  };
}

function formatDateTime(value) {
  if (!value) return '时间待记录';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '时间待记录';
  return date.toLocaleString();
}

function stageLabel(stage) {
  return stageLabels[stage] || stage;
}

function runtimeProfileLabel(profile) {
  return runtimeProfileLabels[profile] || profile || '未识别';
}

function interfaceModeLabel(mode) {
  return interfaceModeLabels[mode] || mode || '未识别';
}

function providerLabel(provider) {
  return provider || '平台待补充';
}

function schedulerModeLabel(mode) {
  return schedulerModeLabels[mode] || mode || '未识别';
}

function deployModeLabel(mode) {
  return deployModeLabels[mode] || mode || '未识别';
}

function analysisIntentLabel(intent) {
  return analysisIntentLabels[intent] || intent || '待识别';
}

function projectKindLabel(kind) {
  return projectKindLabels[kind] || kind || '通用任务';
}

function queueAgentLabel(agent) {
  return queueAgentLabels[agent] || agent || '待分配';
}

function queuePhaseLabel(phase) {
  if (phase === 'quality') return '质量检查';
  return stageLabels[phase] || phase || '待安排';
}

function templateKindLabel(kind) {
  return projectKindLabel(kind);
}

function toolCategoryLabel(category) {
  return toolCategoryLabels[category] || category || '通用能力';
}

function adviceModeLabel(mode) {
  return adviceModeLabels[mode] || mode || '建议';
}

function regionLabel(region) {
  return regionLabels[region] || region || '待补充';
}

function testLayerLabel(layer) {
  return testLayerLabels[layer] || layer || '待补充';
}

function fitScoreLabels() {
  return {
    filter: '接入门槛',
    review: '稳定性',
    alignment: '任务贴合度',
  };
}

function checkpointStatusLabel(status) {
  return checkpointStatusLabels[status] || status || '待开始';
}

function checkpointStatusClass(status) {
  if (status === 'completed') return 'completed';
  if (status === 'running' || status === 'in_progress') return 'running';
  if (status === 'failed') return 'failed';
  return 'paused';
}

function invocationStatusLabel(status) {
  if (status === 'completed') return '已完成';
  if (status === 'running' || status === 'in_progress') return '进行中';
  if (status === 'failed') return '需关注';
  return '待处理';
}

function terminalSessionStatusLabel(status) {
  if (status === 'running') return '进行中';
  if (status === 'completed') return '已完成';
  if (status === 'failed') return '需关注';
  if (status === 'pending' || status === 'queued') return '待处理';
  return '已停止';
}

function auditStatusLabel(status) {
  if (status === 'passed') return '已通过';
  if (status === 'failed') return '需关注';
  return '待处理';
}

function checkpointStatusSummary(checkpoints) {
  const summary = {
    completed: 0,
    inProgress: 0,
    pending: 0,
    attention: 0,
  };
  (checkpoints || []).forEach((item) => {
    if (item.status === 'completed') {
      summary.completed += 1;
      return;
    }
    if (item.status === 'running' || item.status === 'in_progress') {
      summary.inProgress += 1;
      return;
    }
    if (item.status === 'failed') {
      summary.attention += 1;
      return;
    }
    summary.pending += 1;
  });
  return summary;
}

function checkpointTitleTone(id) {
  if (id === 'checkpoint-validation' || id === 'checkpoint-security') return '质量检查';
  if (id === 'checkpoint-preview' || id === 'checkpoint-delivery') return '交付检查';
  return '过程检查';
}

function checkpointPriority(status) {
  if (status === 'failed') return 0;
  if (status === 'running' || status === 'in_progress') return 1;
  if (status === 'pending' || status === 'paused') return 2;
  if (status === 'completed') return 3;
  return 4;
}

function unresolvedFailures(run) {
  return (run.failures || []).filter((item) => !item.resolved);
}

function failureCategoryLabel(category) {
  if (category === 'validation') return '验证失败';
  if (category === 'delivery') return '部署失败';
  if (category === 'repair') return '修复处理';
  return '执行异常';
}

function failureTargetLabel(target) {
  if (!target) return '未指定';
  if (target === 'stable-deploy') return '稳定部署';
  if (target === 'deploy-smoke') return '部署后检查';
  if (target === 'preview-check' || target === 'preview-confirmation') return '预览确认';
  return target;
}

function failureTimeValue(failure) {
  const value = failure?.at ? Date.parse(failure.at) : Number.NaN;
  return Number.isNaN(value) ? 0 : value;
}

function sortFailures(items) {
  return [...(items || [])].sort((left, right) => {
    if (Boolean(left.resolved) !== Boolean(right.resolved)) {
      return left.resolved ? 1 : -1;
    }
    return failureTimeValue(right) - failureTimeValue(left);
  });
}

function failureSourceLabel(failure) {
  if (!failure) return '当前记录';
  if (failure.category === 'validation') {
    if (failure.target === 'security-scan') return '安全检查';
    return '验证检查';
  }
  if (failure.category === 'delivery') {
    if (failure.target === 'stable-deploy') return '稳定部署';
    if (failure.target === 'deploy-smoke') return '部署后检查';
    if (failure.target === 'preview-check' || failure.target === 'preview-confirmation') return '预览确认';
    return '交付检查';
  }
  if (failure.category === 'repair') return '修复处理';
  return stageLabel(failure.stage);
}

function failureRecordLabel(failure) {
  if (!failure) return '处理记录';
  if (failure.target) return failureTargetLabel(failure.target);
  return stageLabel(failure.stage);
}

function summarizeFailureGroup(count, activeLabel, emptyLabel) {
  if (!count) return emptyLabel;
  return count === 1 ? `1 项${activeLabel}` : `${count} 项${activeLabel}`;
}

function dedupeFailureAdvice(items) {
  const grouped = new Map();
  sortFailures(items).forEach((failure) => {
    const key = [failure.category || '', failure.target || '', failure.suggestion || '', failure.resolved ? 'resolved' : 'active'].join('|');
    if (!grouped.has(key)) {
      grouped.set(key, { ...failure, reasons: [failure.reason].filter(Boolean), count: 1 });
      return;
    }
    const current = grouped.get(key);
    current.count += 1;
    if (failure.reason && !current.reasons.includes(failure.reason)) {
      current.reasons.push(failure.reason);
    }
    if (failureTimeValue(failure) > failureTimeValue(current)) {
      current.at = failure.at;
      current.stage = failure.stage;
    }
  });
  return [...grouped.values()];
}

function blockerCopyForFailure(failure) {
  const targetLabel = failure.target ? failureTargetLabel(failure.target) : stageLabel(failure.stage);
  const sourceLabel = failureSourceLabel(failure);
  if (failure.category === 'validation') {
    return {
      title: `${targetLabel}仍需处理`,
      summary: failure.reason || `${sourceLabel}仍未通过。`,
      action: failure.suggestion || '请先修正当前问题，再继续后续检查。',
    };
  }
  if (failure.category === 'delivery') {
    return {
      title: `${targetLabel}仍未确认`,
      summary: failure.reason || `${sourceLabel}还不能直接继续推进。`,
      action: failure.suggestion || '请先补齐当前交付环节，再继续后续流程。',
    };
  }
  return {
    title: `${targetLabel}仍待处理`,
    summary: failure.reason || `${sourceLabel}仍有待处理问题。`,
    action: failure.suggestion || '请先处理当前问题，再继续后续流程。',
  };
}

function splitFailureGroups(failures) {
  const groups = {
    validation: [],
    delivery: [],
    advice: [],
    resolved: [],
  };
  (failures || []).forEach((failure) => {
    if (failure.resolved) {
      groups.resolved.push(failure);
      return;
    }
    if (failure.category === 'validation') {
      groups.validation.push(failure);
    } else if (failure.category === 'delivery') {
      groups.delivery.push(failure);
    }
    groups.advice.push(failure);
  });
  groups.validation = sortFailures(groups.validation);
  groups.delivery = sortFailures(groups.delivery);
  groups.advice = sortFailures(groups.advice);
  groups.resolved = sortFailures(groups.resolved);
  return groups;
}

function renderFailureCards(items, mode = 'default') {
  return (items || []).map((failure) => {
    const badgeClass = failure.resolved ? 'completed' : 'failed';
    const badgeLabel = failure.resolved ? '已处理' : (failureCategoryLabel(failure.category || '') || '需关注');
    const timeMeta = failure.at ? `<p class="run-meta">记录时间：${formatDateTime(failure.at)}</p>` : '';
    const sourceMeta = `<p class="run-meta">来源：${failureSourceLabel(failure)}</p>`;
    const relatedMeta = failure.reasons && failure.reasons.length > 1 ? `<p class="run-meta">关联问题：${failure.reasons.length} 项</p>` : '';
    if (mode === 'advice') {
      return `<div class="list-card"><div class="list-row"><strong>${failureTargetLabel(failure.target) || stageLabel(failure.stage)}</strong><span class="badge ${badgeClass}">${badgeLabel}</span></div><p>${failure.suggestion || '请结合当前上下文继续处理。'}</p>${sourceMeta}<p class="run-meta">对应问题：${failure.reasons?.[0] || failure.reason}</p>${relatedMeta}${timeMeta}</div>`;
    }
    if (mode === 'resolved') {
      return `<div class="list-card"><div class="list-row"><strong>${failureRecordLabel(failure)}</strong><span class="badge ${badgeClass}">${badgeLabel}</span></div><p>${resolvedFailureHint(failure)}</p>${sourceMeta}<p class="run-meta">处理阶段：${stageLabel(failure.stage)}</p><p class="run-meta">原始问题：${failure.reason}</p>${timeMeta}</div>`;
    }
    return `<div class="list-card"><div class="list-row"><strong>${stageLabel(failure.stage)}</strong><span class="badge ${badgeClass}">${badgeLabel}</span></div><p>${failure.reason}</p>${sourceMeta}<p class="run-meta">建议：${failure.suggestion || '请结合当前上下文继续处理。'}</p><p class="run-meta">关联目标：${failureTargetLabel(failure.target)}</p>${timeMeta}</div>`;
  }).join('');
}

function deriveQualityBlocker(run, checkpoints) {
  const failures = unresolvedFailures(run);
  if (failures.length > 0) {
    const item = sortFailures(failures)[0];
    return { ...blockerCopyForFailure(item), tone: 'failed' };
  }
  const failedReport = (run.testReports || []).find((item) => item.status && item.status !== 'passed');
  if (failedReport) {
    return {
      title: `${failedReport.name} 尚未通过`,
      summary: failedReport.summary || '当前验证结果仍需处理。',
      action: '请先修复该项，再继续后续检查。',
      tone: 'failed',
    };
  }
  const checkpoint = [...(checkpoints || [])].sort((left, right) => checkpointPriority(left.status) - checkpointPriority(right.status))[0];
  if (checkpoint && checkpoint.status !== 'completed') {
    return {
      title: checkpoint.title,
      summary: checkpoint.summary || '当前检查项仍未完成。',
      action: checkpoint.gate || '请按当前要求继续推进。',
      tone: checkpointStatusClass(checkpoint.status),
    };
  }
  return {
    title: '当前进展顺畅',
    summary: '当前关键检查都已完成，可以继续推进发布。',
    action: '可继续部署准备，或补充发布记录。',
    tone: 'completed',
  };
}

function deriveQualityBlockerQuickAction(run, checkpoints, deploymentOverview) {
  const checkpoint = [...(checkpoints || [])].sort((left, right) => checkpointPriority(left.status) - checkpointPriority(right.status))[0];
  if (unresolvedFailures(run).length || (run.testReports || []).some((item) => item.status && item.status !== 'passed')) {
    return null;
  }
  if (!checkpoint || checkpoint.status === 'completed') {
    const [suggested] = deriveDeploymentQuickActions(run, deploymentOverview);
    return suggested && !suggested.disabled ? suggested : null;
  }
  if (checkpoint.id === 'checkpoint-preview') {
    const [suggested] = deriveDeploymentQuickActions(run, deploymentOverview);
    if (suggested && suggested.action === 'revalidate-run' && !suggested.disabled) {
      return suggested;
    }
    return {
      action: 'open-preview',
        label: '打开预览',
        disabled: false,
        tone: 'running',
      };
  }
  if (checkpoint.id === 'checkpoint-delivery') {
    const [suggested] = deriveDeploymentQuickActions(run, deploymentOverview);
    return suggested && !suggested.disabled ? suggested : {
      action: 'deploy-run',
      label: '执行稳定部署',
      disabled: !run.remoteDeployEnabled,
      tone: run.remoteDeployEnabled ? 'completed' : 'paused',
    };
  }
  return null;
}

function deploymentReadinessItems(run, checkpoints) {
  const items = [];
  const validation = (checkpoints || []).find((item) => item.id === 'checkpoint-validation');
  const security = (checkpoints || []).find((item) => item.id === 'checkpoint-security');
  const preview = (checkpoints || []).find((item) => item.id === 'checkpoint-preview');
  const delivery = (checkpoints || []).find((item) => item.id === 'checkpoint-delivery');
  items.push({
    label: '稳定部署',
    value: run.remoteDeployEnabled ? '已开启' : '未开启',
    status: run.remoteDeployEnabled ? 'completed' : 'paused',
  });
  items.push({
    label: '验证进度',
    value: validation ? checkpointStatusLabel(validation.status) : '待开始',
    status: validation ? checkpointStatusClass(validation.status) : 'paused',
  });
  if (security) {
    items.push({
      label: '安全进度',
      value: checkpointStatusLabel(security.status),
      status: checkpointStatusClass(security.status),
    });
  }
  if (preview) {
    items.push({
      label: '预览进度',
      value: checkpointStatusLabel(preview.status),
      status: checkpointStatusClass(preview.status),
    });
  }
  items.push({
    label: '发布准备',
    value: delivery ? checkpointStatusLabel(delivery.status) : '待开始',
    status: delivery ? checkpointStatusClass(delivery.status) : 'paused',
  });
  return items;
}

function deriveDeploymentOverview(run, checkpoints) {
  const records = run.deployments || [];
  const latestRecord = records.length ? records[records.length - 1] : null;
  const effectiveRecord = latestRecord;
  const previewCheckpoint = (checkpoints || []).find((item) => item.id === 'checkpoint-preview');
  const readinessItems = deploymentReadinessItems(run, checkpoints);
  const readyCount = readinessItems.filter((item) => item.status === 'completed').length;
  const activeCount = readinessItems.filter((item) => item.status === 'running').length;
  const attentionCount = readinessItems.filter((item) => item.status === 'failed').length;
  const latestLog = (run.logs || []).length ? run.logs[run.logs.length - 1] : null;
  const previewConfirmed = previewCheckpoint?.status === 'completed';
  const previewPending = !previewCheckpoint || ['pending', 'paused'].includes(previewCheckpoint.status);
  let statusKey = 'not-started';
  let statusTitle = '尚未进入部署阶段';
  let statusSummary = '当前仍在准备阶段，请先完成质量总览中的剩余事项。';
  let statusTone = 'paused';
  if (!run.remoteDeployEnabled) {
    statusKey = 'deploy-disabled';
    statusTitle = '尚未开启稳定部署';
    statusSummary = '当前仍以本地验证为主，可按需开启稳定部署。';
  } else if (latestRecord?.mode === 'revalidate' && previewConfirmed) {
    statusKey = 'revalidate-preview-confirmed';
    statusTitle = '复验完成，预览已查看';
    statusSummary = '可继续观察服务状态，或进入下一轮调整。';
    statusTone = 'completed';
  } else if (latestRecord?.mode === 'rollback' && previewConfirmed) {
    statusKey = 'rollback-preview-confirmed';
    statusTitle = '回退完成，预览已查看';
    statusSummary = '请继续留意服务状态。';
    statusTone = 'completed';
  } else if (latestRecord && previewConfirmed) {
    statusKey = 'deploy-preview-confirmed';
    statusTitle = '部署完成，预览已查看';
    statusSummary = '可继续观察服务表现，或进入下一步。';
    statusTone = 'completed';
  } else if (latestRecord?.mode === 'revalidate' && previewPending) {
    statusKey = 'revalidate-preview-pending';
    statusTitle = '复验完成，待查看预览';
    statusSummary = '下一步请查看预览，并留意服务状态。';
    statusTone = 'running';
  } else if (latestRecord?.mode === 'rollback' && previewPending) {
    statusKey = 'rollback-revalidate-pending';
    statusTitle = '回退完成，待复验';
    statusSummary = '请先完成复验，再查看预览与服务状态。';
    statusTone = 'running';
  } else if (latestLog?.message?.includes('已执行稳定部署') && previewPending) {
    statusKey = 'deploy-preview-pending';
    statusTitle = '部署完成，待查看预览';
    statusSummary = '请查看预览，并留意服务状态。';
    statusTone = 'running';
  } else if (latestRecord) {
    statusKey = latestRecord.status === 'completed' ? 'deploy-latest-completed' : 'deploy-latest-status';
    statusTitle = latestRecord.status === 'completed' ? '最近一次部署已完成' : `最近一次部署状态：${latestRecord.status}`;
    statusSummary = latestRecord.summary || '最近一次部署结果已记录。';
    statusTone = latestRecord.status === 'completed' ? 'completed' : latestRecord.status === 'running' ? 'running' : 'failed';
  } else if (attentionCount > 0) {
    statusKey = 'deploy-blocked';
    statusTitle = '部署前仍有待处理项';
    statusSummary = '当前仍有待处理内容，请先完成修复与复验。';
    statusTone = 'failed';
  } else if (activeCount > 0) {
    statusKey = 'deploy-preparing';
    statusTitle = '部署准备进行中';
    statusSummary = '构建、预览与部署准备仍在推进。';
    statusTone = 'running';
  } else if (readyCount === readinessItems.length && readinessItems.length > 0) {
    statusKey = 'deploy-ready';
    statusTitle = '可以开始稳定部署';
    statusSummary = '关键准备已齐备，可以开始发布版本。';
    statusTone = 'completed';
  }
  let rollbackTitle = '回退方案待补充';
  let rollbackSummary = '建议在生成首个版本后补齐回退说明与确认步骤。';
  if (latestRecord) {
    rollbackTitle = `可回退到版本 ${latestRecord.version}`;
    rollbackSummary = '如需回退，请先切回该版本对应内容，再完成复验并查看服务状态。';
  }
  let actionTitle = '暂无动作';
  let actionSummary = '完成首次稳定部署后，将显示最近一次发布、回退或复验。';
  let actionTone = 'paused';
  if (latestRecord?.mode === 'revalidate') {
    actionTitle = '复验';
    actionSummary = previewConfirmed
      ? '预览已查看，可继续观察服务状态。'
      : '待查看预览。';
    actionTone = previewConfirmed ? 'completed' : 'running';
  } else if (latestRecord) {
    if (latestRecord.mode === 'rollback') {
      actionTitle = '回退';
      actionSummary = previewConfirmed
        ? `已回退到 ${latestRecord.version || '目标版本'}，预览已查看。`
        : `已回退到 ${latestRecord.version || '目标版本'}，下一步建议完成复验。`;
      actionTone = previewConfirmed ? 'completed' : 'running';
    } else {
      actionTitle = '稳定部署';
      actionSummary = previewConfirmed
        ? `已部署到 ${deploymentTargetLabel(latestRecord.target)}，版本 ${latestRecord.version || '待补充'} 已记录，预览已查看。`
        : previewPending
        ? `已部署到 ${deploymentTargetLabel(latestRecord.target)}，版本 ${latestRecord.version || '待补充'} 已记录，待查看预览。`
        : `已部署到 ${deploymentTargetLabel(latestRecord.target)}，版本 ${latestRecord.version || '待补充'} 已记录。`;
      actionTone = previewConfirmed
        ? 'completed'
        : previewPending && latestRecord.status === 'completed'
          ? 'running'
          : latestRecord.status === 'completed'
            ? 'completed'
            : deploymentStatusClass(latestRecord.status);
    }
  }
  return {
    readinessItems,
    latestRecord,
    effectiveRecord,
    effectiveVersion: effectiveRecord?.version || '暂无版本',
    effectiveTarget: deploymentTargetLabel(effectiveRecord?.target),
    statusKey,
    statusTitle,
    statusSummary,
    statusTone,
    rollbackTitle,
    rollbackSummary,
    actionTitle,
    actionSummary,
    actionTone,
  };
}

function isDeploymentStatus(deploymentOverview, ...statusKeys) {
  return statusKeys.includes(deploymentOverview?.statusKey);
}

function deriveDeploymentQuickActions(run, deploymentOverview) {
  if (!run.remoteDeployEnabled) {
    return [{
      action: 'deploy-run',
      label: '执行稳定部署',
      disabled: true,
      tone: 'paused',
    }];
  }
  if (isDeploymentStatus(deploymentOverview, 'rollback-revalidate-pending')) {
    return [{
      action: 'revalidate-run',
      label: '执行复验',
      disabled: false,
      tone: 'running',
    }];
  }
  if (isDeploymentStatus(deploymentOverview, 'deploy-preview-pending', 'revalidate-preview-pending')) {
    return [{
      action: 'open-preview',
      label: '打开预览',
      disabled: false,
      tone: 'running',
    }];
  }
  if (isDeploymentStatus(deploymentOverview, 'deploy-preview-confirmed', 'revalidate-preview-confirmed', 'rollback-preview-confirmed')) {
    return [{
      action: 'open-preview',
      label: '再次打开预览',
      disabled: false,
      tone: 'completed',
    }];
  }
  return [{
    action: 'deploy-run',
    label: '执行稳定部署',
    disabled: false,
    tone: deploymentOverview.statusTone === 'failed' ? 'failed' : 'completed',
  }];
}

function shouldShowDeploymentTimelinePreviewAction(deploymentOverview) {
  return isDeploymentStatus(
    deploymentOverview,
    'deploy-preview-pending',
    'revalidate-preview-pending',
    'deploy-preview-confirmed',
    'revalidate-preview-confirmed',
    'rollback-preview-confirmed',
  );
}

function deploymentTimelinePreviewActionLabel(deploymentOverview) {
  return isDeploymentStatus(
    deploymentOverview,
    'deploy-preview-confirmed',
    'revalidate-preview-confirmed',
    'rollback-preview-confirmed',
  ) ? '再次打开预览' : '打开预览';
}

function deriveRollbackSuggestionAction(deploymentOverview) {
  if (!isDeploymentStatus(deploymentOverview, 'rollback-revalidate-pending')) {
    return null;
  }
  return {
    action: 'revalidate-run',
    label: '回退后复验',
    disabled: false,
  };
}

function deriveRunDeploymentSummary(run) {
  const checkpoints = run.analysis?.checkpoints || [];
  const overview = deriveDeploymentOverview(run, checkpoints);
  const readinessItems = overview.readinessItems || [];
  const completedCount = readinessItems.filter((item) => item.status === 'completed').length;
  if (!run.remoteDeployEnabled) {
    return {
      title: '稳定部署未开启',
      summary: '当前仍以本地验证为主，后续可按需开启稳定部署。',
      tone: 'paused',
      progressLabel: '未开启',
      progressValue: 0,
    };
  }
  return {
    title: overview.statusTitle,
    summary: overview.actionSummary || overview.statusSummary,
    tone: overview.statusTone,
    progressLabel: `${completedCount}/${readinessItems.length || 0}`,
    progressValue: readinessItems.length ? Math.round((completedCount / readinessItems.length) * 100) : 0,
  };
}

function focusRunDeploymentSection() {
  q('#run-deployment-section')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function buildRunDeploymentSignature(summary) {
  return [summary.title, summary.progressLabel, summary.summary, summary.tone].join('|');
}

async function runDeploymentAction(button, pendingLabel, failureMessage, action) {
  if (!button) {
    try {
      await action();
    } catch (error) {
      pushEvent(`${failureMessage}：${error.message}`);
      throw error;
    }
    return;
  }
  const previousLabel = button.textContent;
  button.disabled = true;
  button.classList.add('action-button-busy');
  if (pendingLabel) {
    button.textContent = pendingLabel;
  }
  try {
    await action();
  } catch (error) {
    pushEvent(`${failureMessage}：${error.message}`);
    throw error;
  } finally {
    button.disabled = false;
    button.classList.remove('action-button-busy');
    button.textContent = previousLabel;
  }
}

function focusDeploymentTimelineRecord(version) {
  const item = !version
    ? q('[data-deployment-latest="true"]')
    : q(`[data-deployment-version="${CSS.escape(version)}"]`);
  if (!item) {
    return;
  }
  if (deploymentTimelineFocusedItem && deploymentTimelineFocusedItem !== item) {
    deploymentTimelineFocusedItem.classList.remove('deployment-timeline-item-focused');
  }
  item.scrollIntoView({ behavior: 'smooth', block: 'center' });
  item.classList.add('deployment-timeline-item-focused');
  deploymentTimelineFocusedItem = item;
  if (deploymentTimelineFocusTimer) {
    window.clearTimeout(deploymentTimelineFocusTimer);
  }
  deploymentTimelineFocusTimer = window.setTimeout(() => {
    item.classList.remove('deployment-timeline-item-focused');
    deploymentTimelineFocusTimer = null;
    if (deploymentTimelineFocusedItem === item) {
      deploymentTimelineFocusedItem = null;
    }
  }, 1800);
}

function deriveDeploymentDependencies(run, checkpoints) {
  const validation = (checkpoints || []).find((item) => item.id === 'checkpoint-validation');
  const security = (checkpoints || []).find((item) => item.id === 'checkpoint-security');
  const preview = (checkpoints || []).find((item) => item.id === 'checkpoint-preview');
  const delivery = (checkpoints || []).find((item) => item.id === 'checkpoint-delivery');
  const plannedTests = (run.selectedTests && run.selectedTests.length ? run.selectedTests : run.analysis?.recommendedTests) || [];
  const items = [
    {
      label: '验证要求',
      summary: validation?.gate || '完成基础验证后才能进入稳定部署。',
      status: validation ? checkpointStatusClass(validation.status) : 'paused',
    },
    {
      label: '部署材料',
      summary: delivery?.gate || '构建产物与交付材料准备完成后才能进入部署。',
      status: delivery ? checkpointStatusClass(delivery.status) : 'paused',
    },
  ];
  if (security || plannedTests.includes('security-scan')) {
    items.push({
      label: '安全要求',
      summary: security?.gate || '安全检查通过后再执行稳定部署。',
      status: security ? checkpointStatusClass(security.status) : 'paused',
    });
  }
  if (preview) {
    items.push({
      label: '预览要求',
      summary: preview.gate || '预览查看正常后再执行稳定部署。',
      status: checkpointStatusClass(preview.status),
    });
  }
  return items;
}

function q(selector) {
  return document.querySelector(selector);
}

function setView(view) {
  views.forEach((name) => {
    q(`#${name}-view`)?.classList.toggle('hidden', name !== view);
    document.querySelector(`[data-view="${name}"]`)?.classList.toggle('active', name === view);
  });
}

async function fetchJSON(url, options = {}) {
  const response = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: '请求失败' }));
    throw new Error(payload.error || '请求失败');
  }
  return response.json();
}

async function loadDevFiles(path = state.dev.currentDir || '.') {
  const payload = await fetchJSON(`/api/dev/files?path=${encodeURIComponent(path)}`);
  state.dev.currentDir = path;
  state.dev.files = payload.items || [];
  renderDevFiles();
}

async function openDevFile(path) {
  const payload = await fetchJSON(`/api/dev/file?path=${encodeURIComponent(path)}`);
  state.dev.currentFilePath = payload.path || path;
  state.dev.currentFileContent = payload.content || '';
  renderDevEditor();
}

async function saveDevFile() {
  if (!state.dev.currentFilePath) {
    pushEvent('请先选择一个文件再保存。');
    return;
  }
  const payload = await fetchJSON('/api/dev/file', {
    method: 'POST',
    body: JSON.stringify({ path: state.dev.currentFilePath, content: q('#dev-editor')?.value || '' }),
  });
  state.dev.currentFileContent = q('#dev-editor')?.value || '';
  renderDevEditor();
  await loadDevFiles(state.dev.currentDir);
  await trackSelectedRunDevActivity({
    kind: 'file-save',
    target: payload.path,
    status: 'completed',
    detail: `保存文件 ${payload.path}`,
  });
  pushEvent(`文件已保存：${payload.path}`);
}

async function rollbackDevFile() {
  if (!state.dev.currentFilePath) {
    pushEvent('当前没有可回滚的文件。');
    return;
  }
  await fetchJSON('/api/dev/file/rollback', {
    method: 'POST',
    body: JSON.stringify({ path: state.dev.currentFilePath }),
  });
  await openDevFile(state.dev.currentFilePath);
  await loadDevFiles(state.dev.currentDir);
  await trackSelectedRunDevActivity({
    kind: 'file-rollback',
    target: state.dev.currentFilePath,
    status: 'completed',
    detail: `回滚文件 ${state.dev.currentFilePath}`,
  });
  pushEvent(`已回滚文件：${state.dev.currentFilePath}`);
}

function renderDevRunBinding() {
  const node = q('#dev-run-binding');
  if (!node) return;
  const run = state.runs.find((item) => item.id === state.selectedRunId);
  if (!run) {
    node.textContent = '当前未绑定任务，选择右侧任务后可把工作台动作写回任务轨迹。';
    return;
  }
  node.textContent = `当前绑定任务：${run.title}`;
}

function renderDevFiles() {
  const root = q('#dev-files');
  if (!root) return;
  root.innerHTML = '';
  if (state.dev.currentDir !== '.') {
    const upButton = document.createElement('button');
    upButton.type = 'button';
    upButton.className = 'dev-file-item';
    upButton.innerHTML = '<strong>..</strong><span>返回上级目录</span>';
    upButton.onclick = async () => {
      const parts = state.dev.currentDir.split('/').filter(Boolean);
      parts.pop();
      await loadDevFiles(parts.join('/') || '.');
    };
    root.appendChild(upButton);
  }
  (state.dev.files || []).forEach((item) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'dev-file-item';
    button.innerHTML = `<strong>${item.name}</strong><span>${item.isDir ? '目录' : `${Math.max(0, Math.round((item.size || 0) / 1024))} KB`}</span>`;
    button.onclick = async () => {
      try {
        if (item.isDir) {
          await loadDevFiles(item.path);
          return;
        }
        await openDevFile(item.path);
      } catch (error) {
        pushEvent(`打开文件失败：${error.message}`);
      }
    };
    root.appendChild(button);
  });
  if ((state.dev.files || []).length === 0) {
    root.innerHTML = '<div class="detail-empty">当前目录下没有可展示内容。</div>';
  }
}

function renderDevEditor() {
  const pathNode = q('#dev-editor-path');
  const editor = q('#dev-editor');
  if (pathNode) {
    pathNode.textContent = state.dev.currentFilePath || '未打开文件';
  }
  if (editor && editor.value !== state.dev.currentFileContent) {
    editor.value = state.dev.currentFileContent || '';
  }
}

function renderDevTerminal() {
  const output = q('#dev-terminal-output');
  const sessions = q('#dev-terminal-sessions');
  if (output) {
    output.textContent = state.dev.terminalOutput || '等待执行命令。';
  }
  if (!sessions) return;
  sessions.innerHTML = (state.dev.terminalSessions || []).map((item) => `
    <div class="list-card compact-card">
      <div class="list-row split-row">
        <strong>${item.command}</strong>
        <span class="badge ${item.status === 'running' ? 'running' : item.status === 'completed' ? 'completed' : 'failed'}">${terminalSessionStatusLabel(item.status)}</span>
      </div>
      <p class="run-meta">目录 ${item.cwd || '.'} · 退出码 ${item.exitCode ?? 0}${item.runId ? ' · 已关联任务' : ''}</p>
      ${item.status === 'running' ? `<div class="actions compact-actions"><button type="button" class="ghost-button" data-kill-session="${item.id}">停止</button></div>` : ''}
      <pre class="terminal-output">${item.output || '等待输出...'}</pre>
    </div>
  `).join('') || '<div class="detail-empty">当前没有后台终端会话。</div>';
  sessions.querySelectorAll('[data-kill-session]').forEach((button) => {
    button.addEventListener('click', async () => {
      try {
        await stopDevSession(button.dataset.killSession);
      } catch (error) {
        pushEvent(`停止后台命令失败：${error.message}`);
      }
    });
  });
}

function renderDevPreview() {
  const frame = q('#dev-preview-frame');
  const portInput = q('#dev-preview-port');
  if (portInput && portInput.value !== state.dev.previewPort) {
    portInput.value = state.dev.previewPort;
  }
  if (frame && state.dev.previewPort) {
    frame.src = `/api/dev/preview/${encodeURIComponent(state.dev.previewPort)}`;
  }
}

async function openDeploymentPreview(port) {
  const previewPort = String(port || '').trim() || '8080';
  state.dev.previewPort = previewPort;
  setView('dev');
  renderDevPreview();
  await trackSelectedRunDevActivity({
    kind: 'preview-open',
    target: previewPort,
    status: 'completed',
    detail: `打开预览 ${previewPort}`,
  });
  pushEvent(`已打开预览：${previewPort}`);
}

async function loadDevSessions() {
  const payload = await fetchJSON('/api/dev/terminal/sessions');
  state.dev.terminalSessions = payload.items || [];
  if (Array.isArray(payload.updatedRuns) && payload.updatedRuns.length > 0) {
    mergeRuns(payload.updatedRuns);
  }
  renderDevTerminal();
}

async function stopDevSession(id) {
  const payload = await fetchJSON(`/api/dev/terminal/sessions/${encodeURIComponent(id)}/kill`, {
    method: 'POST',
  });
  if (payload.session?.id) {
    state.dev.terminalSessions = (state.dev.terminalSessions || []).map((item) => item.id === payload.session.id ? payload.session : item);
  }
  if (Array.isArray(payload.updatedRuns) && payload.updatedRuns.length > 0) {
    mergeRuns(payload.updatedRuns);
  }
  renderDevTerminal();
  pushEvent(`后台命令已停止：${payload.session?.command || id}`);
}

async function runDevCommand(background) {
  const command = q('#dev-terminal-command')?.value?.trim();
  const cwd = q('#dev-terminal-cwd')?.value?.trim() || '.';
  if (!command) {
    pushEvent('请输入要执行的命令。');
    return;
  }
  const payload = await fetchJSON('/api/dev/terminal/exec', {
    method: 'POST',
    body: JSON.stringify({ command, cwd, background, runId: background ? state.selectedRunId : '' }),
  });
  if (background) {
    await trackSelectedRunDevActivity({
      kind: 'command',
      target: payload.cwd || cwd,
      command: payload.command || command,
      status: payload.status || 'running',
      detail: `后台执行 ${payload.command || command}`,
    });
    pushEvent(`后台命令已启动：${payload.command}`);
    await loadDevSessions();
    return;
  }
  state.dev.terminalOutput = payload.output || '';
  renderDevTerminal();
  await trackSelectedRunDevActivity({
    kind: 'command',
    target: payload.cwd || cwd,
    command: payload.command || command,
    status: payload.status || 'completed',
    detail: `前台执行 ${payload.command || command}`,
  });
  pushEvent(`前台命令执行完成：${payload.command}`);
}

async function trackSelectedRunDevActivity(payload) {
  if (!state.selectedRunId) {
    return;
  }
  try {
    const run = await fetchJSON(`/api/runs/${state.selectedRunId}/dev-activity`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
    state.runs = state.runs.map((item) => item.id === run.id ? run : item);
    renderRuns();
    renderDevRunBinding();
  } catch (error) {
    pushEvent(`任务轨迹同步失败：${error.message}`);
  }
}

function mergeRuns(updatedRuns) {
  if (!Array.isArray(updatedRuns) || updatedRuns.length === 0) {
    return;
  }
  const next = new Map(state.runs.map((item) => [item.id, item]));
  updatedRuns.forEach((run) => {
    if (run?.id) {
      next.set(run.id, run);
    }
  });
  state.runs = Array.from(next.values()).sort((left, right) => new Date(right.createdAt || 0) - new Date(left.createdAt || 0));
  renderRuns();
}

function renderTools() {
  const root = q('#tool-chips');
  root.innerHTML = '';
  const policy = state.system?.runtimePolicy;
  state.tools.forEach((tool) => {
    const button = document.createElement('button');
    const policyDisabled = Array.isArray(policy?.defaultDisabledTools) && policy.defaultDisabledTools.includes(tool.name);
    const toolMeta = getToolCatalogEntry(tool.name);
    const heavy = Boolean(toolMeta?.heavy);
    button.type = 'button';
    button.className = `chip ${tool.enabled ? 'active' : ''}`;
    button.innerHTML = `<strong>${toolLabels[tool.name] || tool.name}</strong><span>${toolDescriptions[tool.name] || '可加入当前任务'}${heavy ? ' · 重动作' : ''}${policyDisabled ? ' · 当前机器默认关闭' : ''}</span>`;
    button.onclick = () => {
      if (!tool.enabled && !canEnableTool(tool.name)) {
        renderRuntimePolicyHint();
        return;
      }
      tool.enabled = !tool.enabled;
      syncRuntimePolicyFormDefaults();
      renderTools();
    };
    root.appendChild(button);
  });
  renderAtomicToolPreview();
  renderRuntimePolicyHint();
}

function syncTestDefaults(forceSelection = false) {
  const select = q('#test-mode-select');
  if (select && forceSelection) {
    select.value = state.testMode || 'template';
  }
  if (forceSelection && state.selectedTests.length === 0 && state.analysisPreview?.recommendedTests?.length) {
    state.selectedTests = [...state.analysisPreview.recommendedTests];
  }
  renderTestOptions();
}

function renderTestOptions() {
  const root = q('#test-chips');
  if (!root) return;
  root.innerHTML = '';
  const recommended = new Set(state.analysisPreview?.recommendedTests || []);
  const catalog = state.system?.testCatalog || [];
  catalog.forEach((item) => {
    const button = document.createElement('button');
    button.type = 'button';
    const selected = state.selectedTests.includes(item.name);
    button.className = `chip ${selected ? 'active' : ''}`;
    button.innerHTML = `<strong>${item.name}</strong><span>${testLayerLabel(item.layer)}${recommended.has(item.name) ? ' · 推荐' : ''}</span>`;
    button.onclick = () => {
      if (selected) {
        state.selectedTests = state.selectedTests.filter((name) => name !== item.name);
      } else {
        state.selectedTests = [...state.selectedTests, item.name];
      }
      renderTestOptions();
    };
    root.appendChild(button);
  });
  if (catalog.length === 0) {
    root.innerHTML = '<div class="detail-empty">测试目录加载后可在这里选择测试项。</div>';
  }
}

function getToolCatalogEntry(name) {
  return (state.system?.toolCatalog || []).find((tool) => tool.name === name);
}

function isHeavyTool(name) {
  return Boolean(getToolCatalogEntry(name)?.heavy);
}

function enabledHeavyToolCount() {
  return state.tools.filter((tool) => tool.enabled && isHeavyTool(tool.name)).length;
}

function isToolEnabled(name) {
  return state.tools.some((tool) => tool.name === name && tool.enabled);
}

function canEnableTool(name) {
  const policy = state.system?.runtimePolicy || {};
  const maxHeavyActions = policy.maxHeavyActions || 0;
  const heavy = isHeavyTool(name);
  if (!heavy) {
    return true;
  }
  if (name === 'deploy' && !policy.allowBackgroundJobs) {
    pushEvent('当前机器默认关闭稳定部署，请先保持轻载运行。');
    return false;
  }
  if (enabledHeavyToolCount() >= maxHeavyActions) {
    pushEvent(`当前机器最多同时启用 ${maxHeavyActions} 个重动作工具，请先关闭其他重工具。`);
    return false;
  }
  return true;
}

function stageDefaultSelection() {
  const policy = state.system?.runtimePolicy || {};
  const defaults = ['intent', 'context', 'plan', 'resource', 'model', 'tool', 'implement', 'result', 'test', 'finalize'];
  if ((policy.maxHeavyActions || 0) > 1) {
    defaults.push('repair');
  }
  if ((policy.maxHeavyActions || 0) > 1 && policy.allowBackgroundJobs && isToolEnabled('deploy')) {
    defaults.push('deploy');
  }
  return stageOptions.filter((stage) => defaults.includes(stage) && !isStageBlocked(stage));
}

function isStageBlocked(stage) {
  const policy = state.system?.runtimePolicy || {};
  if (stage === 'deploy') {
    return !policy.allowBackgroundJobs || !isToolEnabled('deploy');
  }
  if (stage === 'test') {
    return !isToolEnabled('tests');
  }
  return false;
}

function syncRuntimePolicyStageOptions(forceSelection = false) {
  const select = q('#stages-select');
  if (!select) return;
  Array.from(select.options).forEach((option) => {
    option.disabled = isStageBlocked(option.value);
    option.textContent = stageLabels[option.value] || option.value;
    if (option.disabled) {
      option.selected = false;
      option.textContent = `${option.textContent}（当前受限）`;
    }
  });
  const selectedValues = Array.from(select.selectedOptions).map((option) => option.value);
  if (forceSelection || !state.stagePolicyApplied || selectedValues.length === 0) {
    const defaults = new Set(stageDefaultSelection());
    Array.from(select.options).forEach((option) => {
      option.selected = defaults.has(option.value);
    });
    state.stagePolicyApplied = true;
  }
  renderAtomicToolPreview();
}

function syncRuntimePolicyFormDefaults(forceSelection = false) {
  const policy = state.system?.runtimePolicy;
  if (!policy) return;
  const autoRepairInput = q('input[name="autoRepairEnabled"]');
  const autoRepairModeInput = q('#auto-repair-mode');
  const remoteDeployInput = q('input[name="remoteDeployEnabled"]');
  if (autoRepairInput && forceSelection) {
    autoRepairInput.checked = (policy.maxHeavyActions || 0) > 1;
  }
  if (autoRepairModeInput && forceSelection) {
    autoRepairModeInput.value = policy.profile === 'adaptive-performance' ? 'aggressive' : policy.profile === 'home-lite' ? 'lite' : 'standard';
  }
  if (autoRepairModeInput && autoRepairInput) {
    autoRepairModeInput.disabled = !autoRepairInput.checked;
    if (!autoRepairInput.checked) {
      autoRepairModeInput.value = 'lite';
    }
  }
  if (remoteDeployInput) {
    const deployAllowed = policy.allowBackgroundJobs && isToolEnabled('deploy');
    remoteDeployInput.disabled = !deployAllowed;
    if (!deployAllowed) {
      remoteDeployInput.checked = false;
    } else if (forceSelection) {
      remoteDeployInput.checked = policy.profile === 'adaptive-performance';
    }
  }
  syncRuntimePolicyStageOptions(forceSelection);
  if (forceSelection && !state.testMode) {
    state.testMode = policy.profile === 'home-lite' ? 'light' : 'template';
  }
  renderRuntimePolicyHint();
}

function getLocalModelPolicyStats(packs = []) {
  return {
    total: packs.length,
    ready: packs.filter((pack) => pack.installState === 'ready').length,
    onDemand: packs.filter((pack) => pack.policyMode === 'on-demand').length,
    externallyManaged: packs.filter((pack) => pack.policyMode === 'externally-managed').length,
    displayOnly: packs.filter((pack) => pack.policyMode === 'display-only').length,
    recommended: packs.filter((pack) => pack.recommended).length,
  };
}

function buildLocalModelPolicyNarrative(policy, stats) {
  if (!policy) {
    return {
      summary: '正在根据当前机器规格分析本地模型建议。',
      badges: [],
    };
  }
  const summary = policy.allowLocalModels
    ? `当前机器以“按需启用”为主，可直接处理 ${stats.onDemand} 个本地候选；高体量候选将优先转为外部托管或仅展示。`
    : `当前机器以“仅展示”为主，${stats.displayOnly} 个候选保留目录信息，避免本地推理占用过多资源。`;
  return {
    summary,
    badges: [
      `按需启用 ${stats.onDemand}`,
      `外部托管 ${stats.externallyManaged}`,
      `仅展示 ${stats.displayOnly}`,
      `已就绪 ${stats.ready}`,
    ],
  };
}

function selectedStageValues() {
  return Array.from(q('#stages-select')?.selectedOptions || []).map((option) => option.value);
}

function shouldPreferPreviewAtomicTool(current, candidate, enabledTools) {
  const currentPreferred = enabledTools.has(current.preferredProvider);
  const candidatePreferred = enabledTools.has(candidate.preferredProvider);
  if (currentPreferred !== candidatePreferred) {
    return candidatePreferred;
  }
  if (Boolean(current.localFirst) !== Boolean(candidate.localFirst)) {
    return Boolean(candidate.localFirst);
  }
  if (Boolean(current.recommended) !== Boolean(candidate.recommended)) {
    return Boolean(candidate.recommended);
  }
  if ((current.priority || 0) !== (candidate.priority || 0)) {
    return (candidate.priority || 0) < (current.priority || 0);
  }
  return (candidate.id || '') < (current.id || '');
}

function currentAtomicAssemblyPreview() {
  const registry = state.system?.atomicTools || [];
  const enabledTools = new Set(state.tools.filter((tool) => tool.enabled).map((tool) => tool.name));
  const selectedStages = new Set(selectedStageValues());
  const selectedByGroup = new Map();
  registry.forEach((item) => {
    const stageMatch = (item.stageKinds || []).some((stage) => selectedStages.has(stage));
    const providerMatch = enabledTools.has(item.preferredProvider) || (item.fallbackProviders || []).some((provider) => enabledTools.has(provider));
    if (!item.allowed || !stageMatch || !providerMatch) {
      return;
    }
    const group = item.dedupGroup || item.id;
    const current = selectedByGroup.get(group);
    if (!current || shouldPreferPreviewAtomicTool(current, item, enabledTools)) {
      selectedByGroup.set(group, item);
    }
  });
  return Array.from(selectedByGroup.values()).sort((left, right) => {
    if ((left.priority || 0) !== (right.priority || 0)) {
      return (left.priority || 0) - (right.priority || 0);
    }
    return (left.name || '').localeCompare(right.name || '');
  });
}

function bundleDefinitionList() {
  return [
    {
      id: 'bundle-local-core',
      name: '本地核心闭环',
      description: '适合大多数本地优先任务，优先保留分析、工作区改写、验证和观测。',
      match: (tool) => ['planner', 'code', 'quality', 'ops'].includes(tool.category),
    },
    {
      id: 'bundle-local-delivery',
      name: '本地交付闭环',
      description: '在当前机器允许时补上构建与交付入口，形成从改写到发布的完整链路。',
      match: (tool) => ['planner', 'code', 'quality', 'release', 'ops'].includes(tool.category),
    },
    {
      id: 'bundle-knowledge-office',
      name: '知识与办公',
      description: '适合研究、文档、办公和日常事务，减少无关的工程重动作。',
      match: (tool) => ['planner', 'research', 'knowledge', 'office', 'daily'].includes(tool.category),
    },
    {
      id: 'bundle-creative-assets',
      name: '内容与资产',
      description: '适合内容规划、资源整理和多媒体处理，默认仍以本地优先组织为主。',
      match: (tool) => ['planner', 'assets', 'media', 'research'].includes(tool.category),
    },
  ];
}

function bundleFitSummary(bundle) {
  if (bundle.recommendationReason) {
    return bundle.recommendationReason;
  }
  if (bundle.id === 'bundle-local-delivery') {
    return '适合需要构建、验证并最终交付的任务。';
  }
  if (bundle.id === 'bundle-knowledge-office') {
    return '适合研究、文档、办公和日常整理类任务。';
  }
  if (bundle.id === 'bundle-creative-assets') {
    return '适合素材规划、内容整理和多媒体处理任务。';
  }
  return '适合分析、改写和验证为主的本地优先任务。';
}

function preferredBundleId(bundles) {
  const analysisPreferred = state.analysisPreview?.recommendedBundle?.id;
  if (analysisPreferred && bundles.some((bundle) => bundle.id === analysisPreferred)) {
    return analysisPreferred;
  }
  const ranked = [...bundles].sort((left, right) => {
    if (left.items.length !== right.items.length) {
      return right.items.length - left.items.length;
    }
    const leftLocal = left.items.filter((item) => item.localFirst).length;
    const rightLocal = right.items.filter((item) => item.localFirst).length;
    if (leftLocal !== rightLocal) {
      return rightLocal - leftLocal;
    }
    return left.name.localeCompare(right.name);
  });
  return ranked[0]?.id || '';
}

function previewBundles() {
  const preview = currentAtomicAssemblyPreview();
  const analysisPreferred = state.analysisPreview?.recommendedBundle;
  const bundles = bundleDefinitionList().map((bundle) => {
    const items = preview.filter((tool) => bundle.match(tool));
    return {
      ...bundle,
      items,
      recommendationReason: analysisPreferred?.id === bundle.id ? analysisPreferred.reason : '',
    };
  }).filter((bundle) => bundle.items.length > 0);
  const preferred = preferredBundleId(bundles);
  return bundles.map((bundle) => ({
    ...bundle,
    recommended: bundle.id === preferred,
  })).sort((left, right) => {
    if (Boolean(left.recommended) !== Boolean(right.recommended)) {
      return left.recommended ? -1 : 1;
    }
    if (left.items.length !== right.items.length) {
      return right.items.length - left.items.length;
    }
    return left.name.localeCompare(right.name);
  });
}

function applyAtomicBundle(bundleId) {
  const bundle = previewBundles().find((item) => item.id === bundleId);
  if (!bundle) return;
  const requiredProviders = new Set();
  const requiredStages = new Set(selectedStageValues());
  bundle.items.forEach((item) => {
    if (item.preferredProvider) {
      requiredProviders.add(item.preferredProvider);
    }
    (item.fallbackProviders || []).forEach((provider) => requiredProviders.add(provider));
    (item.stageKinds || []).forEach((stage) => requiredStages.add(stage));
  });
  state.tools = state.tools.map((tool) => ({
    ...tool,
    enabled: requiredProviders.has(tool.name) || (tool.name === 'logs' && bundle.id !== 'bundle-knowledge-office') || (tool.name === 'analysis'),
  }));
  const select = q('#stages-select');
  Array.from(select?.options || []).forEach((option) => {
    option.selected = requiredStages.has(option.value) && !option.disabled;
  });
  renderTools();
  syncRuntimePolicyFormDefaults();
  renderAtomicBundles();
  pushEvent(`已应用工具方案：${bundle.name}`);
}

function renderAtomicBundles() {
  const root = q('#atomic-bundles');
  if (!root) return;
  const bundles = previewBundles();
  if (bundles.length === 0) {
    root.innerHTML = '<div class="detail-empty">当前阶段与工具组合还不足以生成可用方案。</div>';
    return;
  }
  root.innerHTML = bundles.map((bundle) => `
    <div class="list-card compact-card atomic-bundle-card ${bundle.recommended ? 'atomic-bundle-card-recommended' : ''}">
      <div class="list-row split-row">
        <strong>${bundle.name}</strong>
        <div class="stage-list">
          <span class="badge">${bundle.items.length} 项能力</span>
          ${bundle.recommended ? '<span class="badge completed">当前最适合</span>' : ''}
        </div>
      </div>
      <p>${bundleFitSummary(bundle)}</p>
      <div class="stage-list">${bundle.items.slice(0, 6).map((item) => `<span class="badge ${item.localFirst ? 'completed' : 'paused'}">${item.name}</span>`).join('')}</div>
      <div class="actions"><button type="button" data-bundle-id="${bundle.id}">应用工具方案</button></div>
    </div>
  `).join('');
  root.querySelectorAll('button[data-bundle-id]').forEach((button) => {
    button.addEventListener('click', () => applyAtomicBundle(button.dataset.bundleId));
  });
}

function renderAtomicToolPreview() {
  const root = q('#atomic-tool-preview');
  if (!root) return;
  const bundles = previewBundles();
  const recommendedBundle = bundles.find((bundle) => bundle.recommended);
  const preferredTemplate = currentPreferredTemplate();
  if (!state.system?.atomicTools) {
    root.className = 'panel-subtle runtime-policy-hint span-2';
    root.textContent = '正在根据阶段、工具和机器情况整理推荐方案。';
    return;
  }
  const preview = currentAtomicAssemblyPreview();
  const selectedStages = selectedStageValues();
  if (preview.length === 0) {
    root.className = 'panel-subtle runtime-policy-hint span-2';
    root.innerHTML = `<div class="runtime-policy-head"><strong>推荐方案</strong><span class="badge">待启用</span></div><p>${state.analysisPreview?.recommendedBundle?.reason || '当前阶段或工具组合还不足以整理出更合适的本地方案，建议开启工作区、分析、验证等核心入口。'}</p><p class="run-meta">当前阶段：${selectedStages.map((stage) => stageLabels[stage] || stage).join(' / ') || '未选择'}</p>`;
    return;
  }
  const categories = preview.map((item) => `<span class="badge">${toolCategoryLabel(item.category)}</span>`).join('');
  const localFirstCount = preview.filter((item) => item.localFirst).length;
  const recommendedItems = recommendedBundle?.items || preview;
  root.className = 'panel-subtle runtime-policy-hint span-2';
  root.innerHTML = `
    <div class="runtime-policy-head">
      <strong>推荐方案</strong>
      <span class="badge">${preview.length} 项已装配</span>
    </div>
    <p>系统会优先结合本地方式、同类能力和当前机器情况，给出更合适的可用入口。</p>
    <div class="policy-snapshot">
      <div>
        <span>本地优先能力</span>
        <strong>${localFirstCount} / ${preview.length}</strong>
      </div>
      <div>
        <span>当前阶段</span>
        <strong>${selectedStages.map((stage) => stageLabels[stage] || stage).join(' / ') || '未选择'}</strong>
      </div>
      <div>
        <span>当前建议</span>
        <strong>${recommendedBundle?.name || '本地核心闭环'}</strong>
      </div>
      <div>
        <span>当前模板</span>
        <strong>${preferredTemplate?.name || '待选择模板'}</strong>
      </div>
    </div>
    <p class="run-meta">${bundleFitSummary(recommendedBundle || { id: 'bundle-local-core' })}</p>
      ${preferredTemplate ? `<p class="run-meta">模板摘要：${preferredTemplate.name} · ${preferredTemplate.recommendedBundle?.name || '系统自动匹配'} · ${(preferredTemplate.defaultStages || []).map((stage) => stageLabel(stage)).join(' / ')}</p>` : ''}
    <div class="stage-list">${categories}</div>
    <div class="stack-list atomic-preview-list">${preview.map((item) => `
      <div class="list-card compact-card">
        <div class="list-row split-row">
          <strong>${item.name}</strong>
          <span class="badge ${item.localFirst ? 'completed' : 'paused'}">${item.localFirst ? '本地优先' : '备用入口'}</span>
        </div>
        <p>${item.summary}</p>
        <p class="run-meta">首选入口 ${toolLabels[item.preferredProvider] || item.preferredProvider} · 已按同类能力整理</p>
      </div>`).join('')}</div>
    <div class="stage-list">${recommendedItems.slice(0, 4).map((item) => `<span class="badge completed">${item.name}</span>`).join('')}</div>
  `;
  renderAtomicBundles();
}

function renderRuntimePolicyHint() {
  const root = q('#runtime-policy-hint');
  if (!root) return;
  const policy = state.system?.runtimePolicy;
  if (!policy) {
    root.className = 'panel-subtle runtime-policy-hint span-2';
    root.textContent = '正在整理当前机器的使用建议。';
    return;
  }
  const heavyUsed = enabledHeavyToolCount();
  const selectedStages = Array.from(q('#stages-select')?.selectedOptions || []).map((option) => stageLabels[option.value] || option.value);
  const selectedStageText = selectedStages.length ? selectedStages.join(' / ') : '尚未选择阶段';
  const repairMode = q('#auto-repair-mode')?.value || 'standard';
  const autoRepairEnabled = q('input[name="autoRepairEnabled"]')?.checked;
  const localStats = getLocalModelPolicyStats(state.system?.builtinModelPacks || []);
  const localNarrative = buildLocalModelPolicyNarrative(policy, localStats);
  const primaryDirection = primaryGovernanceDirection(localStats);
  const summary = [
    `当前模式 ${runtimeProfileLabel(policy.profile)}`,
    `重动作 ${heavyUsed}/${policy.maxHeavyActions || 0}`,
    `后台任务${policy.allowBackgroundJobs ? '可按需启用' : '暂不开放'}`,
    `本地模型${policy.allowLocalModels ? '按需启用优先' : '仅展示优先'}`,
    `自动修复${autoRepairEnabled ? `为${repairMode}` : '关闭'}`,
  ];
  root.className = `panel-subtle runtime-policy-hint span-2 ${heavyUsed >= (policy.maxHeavyActions || 0) && (policy.maxHeavyActions || 0) > 0 ? 'runtime-policy-hint-warning' : ''}`;
  root.innerHTML = `
    <div class="runtime-policy-head">
      <strong>使用建议</strong>
      <span class="badge">${runtimeProfileLabel(policy.profile)}</span>
    </div>
    <p>${policy.summary || '系统正在根据当前机器规格整理使用建议。'}</p>
    <div class="policy-snapshot">
      <div>
        <span>本地模型方式</span>
        <strong>${policy.allowLocalModels ? '按需启用优先' : '仅展示优先'}</strong>
      </div>
      <div>
        <span>本地模型说明</span>
        <strong>${localNarrative.summary}</strong>
      </div>
      <div>
        <span>当前重点</span>
        <strong>${primaryDirection}</strong>
      </div>
    </div>
    <div class="stage-list">
      ${summary.map((item) => `<span class="badge">${item}</span>`).join('')}
    </div>
    <div class="stage-list">
      ${localNarrative.badges.map((item) => `<span class="badge">${item}</span>`).join('')}
    </div>
      <p class="run-meta">当前默认阶段：${selectedStageText}</p>
  `;
}

function renderTemplates() {
  const select = q('#template-select');
  const list = q('#templates-list');
  const current = select.value;
  const preferredId = preferredTemplateId();
  select.innerHTML = '';
  list.innerHTML = '';
  const groups = groupedTemplates();
  groups.flatMap((group) => group.items).forEach((tpl) => {
    const option = document.createElement('option');
    option.value = tpl.id;
    option.textContent = `${tpl.name} (${templateKindLabel(tpl.kind)})`;
    option.selected = tpl.id === current;
    select.appendChild(option);
  });
  list.innerHTML = groups.map((group) => `
    <section class="template-group ${group.collapsed ? 'template-group-collapsed' : ''}">
      <button type="button" class="template-group-head" data-template-group="${group.id}">
        <div>
          <strong>${group.name}</strong>
          <p>${group.summary}</p>
          <p class="run-meta template-group-reason">${templateGroupRecommendation(group)}</p>
        </div>
        <div class="stage-list">
          <span class="badge">${group.items.length} 个模板</span>
          ${group.recommended ? '<span class="badge completed">当前优先组</span>' : '<span class="badge">按需展开</span>'}
        </div>
      </button>
      <div class="stack-list template-group-body">
        ${group.items.map((tpl) => {
          const stageBadges = (tpl.defaultStages || []).map((stage) => `<span class="badge">${stageLabel(stage)}</span>`).join('');
          const toolBadges = (tpl.defaultTools || []).slice(0, 6).map((tool) => `<span class="badge">${toolLabels[tool] || tool}</span>`).join('');
          return `
            <div class="list-card ${tpl.id === preferredId ? 'template-card-recommended' : ''}">
              <div class="list-row split-row">
                <h3>${tpl.name}</h3>
                <div class="stage-list">
                  <span class="badge">${templateKindLabel(tpl.kind)}</span>
                  ${tpl.id === preferredId ? '<span class="badge completed">当前最适合</span>' : ''}
                  ${tpl.recommendedBundle?.name ? `<span class="badge completed">${tpl.recommendedBundle.name}</span>` : ''}
                </div>
              </div>
              <p>${tpl.description}</p>
              <div class="stage-list">${stageBadges || '<span class="badge">默认阶段待补充</span>'}</div>
              <div class="stage-list">${toolBadges || '<span class="badge">默认工具待补充</span>'}</div>
              ${tpl.recommendedBundle?.reason ? `<p class="run-meta">${tpl.recommendedBundle.reason}</p>` : ''}
              <div class="actions template-card-actions">
                <button type="button" data-template-apply="${tpl.id}">${tpl.id === preferredId ? '应用当前推荐模板' : '快速应用模板'}</button>
                <button type="button" class="ghost-button" data-template-inline="${tpl.id}">只应用不跳转</button>
              </div>
            </div>
          `;
        }).join('')}
      </div>
    </section>
  `).join('');
  list.querySelectorAll('[data-template-group]').forEach((button) => {
    button.addEventListener('click', () => {
      const groupId = button.dataset.templateGroup;
      state.templateGroupState[groupId] = !state.templateGroupState[groupId];
      renderTemplates();
    });
  });
  list.querySelectorAll('[data-template-apply]').forEach((button) => {
    button.addEventListener('click', () => {
      applyTemplateFromLibrary(button.dataset.templateApply);
    });
  });
  list.querySelectorAll('[data-template-inline]').forEach((button) => {
    button.addEventListener('click', () => {
      applyTemplateInline(button.dataset.templateInline);
    });
  });
}

function templateGroupMeta(template) {
  const bundleId = template.recommendedBundle?.id;
  if (bundleId === 'bundle-local-delivery') {
    return { id: 'delivery', name: '闭环与发布', summary: '适合完整实现、验证、交付与修复链路。' };
  }
  if (bundleId === 'bundle-knowledge-office') {
    return { id: 'knowledge', name: '知识与办公', summary: '适合研究整理、文档沉淀和办公交付。' };
  }
  if (bundleId === 'bundle-creative-assets') {
    return { id: 'creative', name: '内容与资产', summary: '适合素材规划、内容组织和多媒体处理。' };
  }
  if (template.kind === 'toolset') {
    return { id: 'foundation', name: '基础能力', summary: '适合先补齐底层工具入口和基础执行栈。' };
  }
  return { id: 'general', name: '通用模板', summary: '适合暂未明确方向的常规任务。' };
}

function preferredTemplateGroupId() {
  const preferred = currentPreferredTemplate();
  return preferred ? templateGroupMeta(preferred).id : '';
}

function groupedTemplates() {
  const preferredGroupId = preferredTemplateGroupId();
  const groups = new Map();
  rankedTemplates().forEach((template) => {
    const meta = templateGroupMeta(template);
    if (!groups.has(meta.id)) {
      const collapsed = Object.prototype.hasOwnProperty.call(state.templateGroupState, meta.id)
        ? state.templateGroupState[meta.id]
        : meta.id !== preferredGroupId && preferredGroupId !== '';
      groups.set(meta.id, {
        ...meta,
        recommended: meta.id === preferredGroupId,
        collapsed,
        items: [],
      });
    }
    groups.get(meta.id).items.push(template);
  });
  return [...groups.values()];
}

function templateGroupRecommendation(group) {
  const analysis = state.analysisPreview;
  if (!group.recommended) {
    return '当前任务与该组的匹配度较低，按需展开查看即可。';
  }
  if (analysis?.recommendedBundle?.name) {
    return `当前任务更适合“${analysis.recommendedBundle.name}”路线，因此优先展开这一组。`;
  }
  return '当前任务的阶段、工具方案和模板结构更贴近这一组，适合作为默认入口。';
}

function templateIntentScore(template, analysis) {
  let score = 0;
  const stages = new Set(template.defaultStages || []);
  (analysis?.recommendedStages || []).forEach((stage) => {
    if (stages.has(stage)) {
      score += 6;
    }
  });
  if (analysis?.recommendedBundle?.id && template.recommendedBundle?.id === analysis.recommendedBundle.id) {
    score += 60;
  }
  if (analysis?.intent === 'build-deploy' && stages.has('deploy')) {
    score += 24;
  }
  if (analysis?.intent === 'fix' && stages.has('repair')) {
    score += 18;
  }
  if (analysis?.projectKind === 'research' || analysis?.projectKind === 'office' || analysis?.projectKind === 'docs-content' || analysis?.projectKind === 'daily') {
    if (template.recommendedBundle?.id === 'bundle-knowledge-office') {
      score += 20;
    }
  }
  if (analysis?.projectKind === 'game' && template.recommendedBundle?.id === 'bundle-creative-assets') {
    score += 20;
  }
  if (template.kind === 'workflow') {
    score += 4;
  }
  return score;
}

function rankedTemplates() {
  const analysis = state.analysisPreview;
  return [...state.templates].sort((left, right) => {
    const leftScore = templateIntentScore(left, analysis);
    const rightScore = templateIntentScore(right, analysis);
    if (leftScore !== rightScore) {
      return rightScore - leftScore;
    }
    if (left.kind !== right.kind) {
      return left.kind.localeCompare(right.kind);
    }
    return left.name.localeCompare(right.name);
  });
}

function preferredTemplateId() {
  return rankedTemplates()[0]?.id || '';
}

function currentPreferredTemplate() {
  return templateById(preferredTemplateId());
}

function templateById(templateId) {
  return state.templates.find((tpl) => tpl.id === templateId) || null;
}

function runTemplate(run) {
  return templateById(run?.templateId);
}

function runTemplateLabel(run) {
  return runTemplate(run)?.name || run.templateId || '未选择模板';
}

function runBundleLabel(run) {
  return runTemplate(run)?.recommendedBundle?.name || run.analysis?.recommendedBundle?.name || '待确认';
}

function runAssemblySummary(run) {
  const assembled = run.assembledTools || [];
  const localFirst = assembled.filter((tool) => tool.localFirst).length;
  if (assembled.length === 0) {
    return '工具方案尚未就绪';
  }
  return `装配 ${assembled.length} 项 · 本地优先 ${localFirst}/${assembled.length}`;
}

function runOffsetHint(run) {
  const template = runTemplate(run);
  const offsets = templateOffsetItems(run, template);
  if (offsets.length === 0) {
    return '模板结构与实际装配基本一致';
  }
  if (offsets.length === 1) {
    return '模板结构存在轻微偏移';
  }
  return `模板结构有 ${offsets.length} 处偏移`;
}

function runFailureHeadline(run) {
  const failures = sortFailures(run?.failures || []);
  const active = failures.find((item) => !item.resolved);
  if (active) {
    return `${failureSourceLabel(active)}待处理`;
  }
  if (failures.length > 0) {
    return `${failureSourceLabel(failures[0])}已处理`;
  }
  return '当前进展稳定';
}

function resolvedFailureHint(failure) {
  if (!failure) return '当前暂无处理记录';
  if (failure.category === 'validation') {
    return `${failureTargetLabel(failure.target)}已复验通过`;
  }
  if (failure.category === 'delivery') {
    if (failure.target === 'preview-check' || failure.target === 'preview-confirmation') {
      return '预览确认已完成';
    }
    if (failure.target === 'deploy-smoke') {
      return '部署后检查已复验通过';
    }
    if (failure.target === 'stable-deploy') {
      return '稳定部署问题已处理';
    }
  }
  return `已处理：${failureSourceLabel(failure)}`;
}

function runFailureActionHint(failure) {
  if (!failure) return '请按当前记录继续处理。';
  if (failure.category === 'validation') {
    return failure.suggestion || `请先完成${failureTargetLabel(failure.target)}，再继续后续检查。`;
  }
  if (failure.category === 'delivery') {
    if (failure.target === 'preview-check' || failure.target === 'preview-confirmation') {
      return failure.suggestion || '请先完成预览确认，再继续后续推进。';
    }
    if (failure.target === 'deploy-smoke') {
      return failure.suggestion || '请先完成部署后检查，再继续后续推进。';
    }
    if (failure.target === 'stable-deploy') {
      return failure.suggestion || '请先处理稳定部署问题，再继续后续推进。';
    }
    return failure.suggestion || `请先完成${failureTargetLabel(failure.target)}，再继续后续推进。`;
  }
  return failure.suggestion || '请先处理当前问题，再继续后续流程。';
}

function runFailureSummary(run) {
  const groups = splitFailureGroups(run?.failures || []);
  const active = [...groups.validation, ...groups.delivery];
  if (active.length > 0) {
    const top = sortFailures(active)[0];
    return runFailureActionHint(top);
  }
  if (groups.resolved.length > 0) {
    return resolvedFailureHint(groups.resolved[0]);
  }
  return '当前没有待处理问题，可继续推进。';
}

function runFailureContextMeta(run) {
  const failures = sortFailures(run?.failures || []);
  const focus = failures.find((item) => !item.resolved) || failures[0];
  if (!focus) return '当前没有额外上下文需要关注。';
  return `来源 ${failureSourceLabel(focus)} · 关联目标 ${failureTargetLabel(focus.target)}`;
}

function runFailureStatusMeta(run) {
  const groups = splitFailureGroups(run?.failures || []);
  const adviceCount = dedupeFailureAdvice(groups.advice).length;
  const fragments = [];
  if (groups.validation.length > 0) {
    fragments.push(`验证待处理 ${groups.validation.length}`);
  }
  if (groups.delivery.length > 0) {
    fragments.push(`部署待处理 ${groups.delivery.length}`);
  }
  if (adviceCount > 0) {
    fragments.push(`处理建议 ${adviceCount}`);
  }
  if (groups.resolved.length > 0) {
    fragments.push(`处理记录 ${groups.resolved.length}`);
  }
  return fragments.join(' · ') || '当前无失败摘要';
}

function templateSourceLabel(source) {
  switch (source) {
    case 'auto':
      return '智能预演自动推荐';
    case 'manual-select':
      return '手动切换';
    case 'template-library':
      return '模板库快速应用';
    case 'template-inline':
      return '模板区直接应用';
    case 'recent-template':
      return '最近使用模板';
    default:
      return source || '未记录';
  }
}

function devActivityLabel(kind) {
  switch (kind) {
    case 'file-save':
      return '保存文件';
    case 'file-rollback':
      return '回滚文件';
    case 'preview-open':
      return '打开预览';
    case 'command':
      return '执行命令';
    default:
      return kind || '工作台活动';
  }
}

function devActivitySummary(item) {
  if (!item) return '工作台活动';
  return item.detail || item.command || item.target || devActivityLabel(item.kind);
}

function devActivityMeta(item) {
  if (!item) return '未指定目标';
  if (item.target && item.command) {
    return `${item.target} · ${item.command}`;
  }
  return item.target || item.command || '未指定目标';
}

function runLogLevelLabel(level) {
  switch (level) {
    case 'info':
      return '进展';
    case 'warn':
      return '提醒';
    case 'error':
      return '异常';
    default:
      return level || '记录';
  }
}

function policyDecisionAreaLabel(area) {
  switch (area) {
    case 'tool':
      return '工具策略';
    case 'stage':
      return '阶段调整';
    case 'deploy':
      return '部署策略';
    case 'repair':
      return '修复策略';
    case 'atomic-tool':
      return '能力装配';
    case 'local-model':
      return '本地模型';
    case 'template':
      return '模板处理';
    default:
      return area || '系统判断';
  }
}

function policyDecisionActionLabel(action) {
  switch (action) {
    case 'disabled':
      return '已关闭';
    case 'removed':
      return '已移除';
    case 'downgraded':
      return '已收束';
    case 'set':
      return '已设置';
    case 'selected':
      return '已选用';
    case 'missing':
      return '待补齐';
    case 'applied':
      return '已应用';
    case 'source':
      return '来源';
    default:
      return action || '已记录';
  }
}

function policyDecisionReasonLabel(reason) {
  switch (reason) {
    case 'default-disabled':
      return '默认关闭';
    case 'background-jobs-disabled':
      return '后台重任务受限';
    case 'heavy-limit':
      return '重动作数量受限';
    case 'deploy-unavailable':
      return '部署不可用';
    case 'tests-unavailable':
      return '验证不可用';
    case 'repair-disabled':
      return '自动修复关闭';
    case 'runtime-policy':
      return '按运行策略调整';
    case 'no-assembly':
      return '装配结果缺失';
    case 'runtime-tier':
      return '按机器分层调整';
    default:
      return reason || '系统记录';
  }
}

function rememberRecentTemplate(templateId) {
  const target = templateById(templateId);
  if (!target) return;
  state.recentTemplateIds = [templateId, ...state.recentTemplateIds.filter((id) => id !== templateId)].slice(0, 4);
  renderRecentTemplates();
}

function recentTemplates() {
  const order = new Map(state.recentTemplateIds.map((id, index) => [id, index]));
  return state.recentTemplateIds
    .map((id) => templateById(id))
    .filter(Boolean)
    .sort((left, right) => {
      const leftScore = templateIntentScore(left, state.analysisPreview);
      const rightScore = templateIntentScore(right, state.analysisPreview);
      if (leftScore !== rightScore) {
        return rightScore - leftScore;
      }
      return (order.get(left.id) || 0) - (order.get(right.id) || 0);
    });
}

function renderRecentTemplates() {
  const root = q('#recent-templates');
  if (!root) return;
  const items = recentTemplates();
  if (items.length === 0) {
    root.innerHTML = '<div class="detail-empty">最近使用的模板会保留在这里，方便直接回填到创建区。</div>';
    return;
  }
  root.innerHTML = items.map((template) => `
    <button type="button" class="chip recent-template-chip" data-recent-template="${template.id}">
      <strong>${template.name}</strong>
      <span>${template.recommendedBundle?.name || '系统自动匹配'}</span>
    </button>
  `).join('');
  root.querySelectorAll('[data-recent-template]').forEach((button) => {
    button.addEventListener('click', () => {
      applyTemplateFromLibrary(button.dataset.recentTemplate, 'recent-template');
    });
  });
}

function toolNameBadge(tool) {
  return toolLabels[tool] || tool;
}

function templateOffsetItems(run, template) {
  const offsets = [];
  const actualStages = new Set((run.stages || []).map((stage) => stage.kind));
  (template?.defaultStages || []).forEach((stage) => {
    if (!actualStages.has(stage)) {
      offsets.push(`模板阶段 ${stageLabel(stage)} 未进入实际执行，多半是当前使用建议或工具开关暂不需要该阶段。`);
    }
  });
  const actualAssembly = new Set((run.assembledTools || []).map((tool) => tool.category));
  const bundleName = template?.recommendedBundle?.name || run.analysis?.recommendedBundle?.name;
  if (bundleName && actualAssembly.size > 0) {
    offsets.push(`当前实际能力更贴近“${bundleName}”，但仍会优先考虑本地方式、同类能力和当前机器限制。`);
  }
  (run.policyDecisions || [])
    .filter((decision) => decision.area !== 'template')
    .slice(0, 3)
    .forEach((decision) => offsets.push(decision.message));
  return Array.from(new Set(offsets.filter(Boolean)));
}

function applyTemplateSelection(templateId) {
  const template = templateById(templateId);
  if (!template) return;
  const select = q('#stages-select');
  const defaultStages = new Set(template.defaultStages || []);
  if (select && defaultStages.size > 0) {
    Array.from(select.options).forEach((option) => {
      option.selected = defaultStages.has(option.value) && !option.disabled;
    });
  }
  if (Array.isArray(template.defaultTools) && template.defaultTools.length > 0) {
    const enabledSet = new Set(template.defaultTools);
    state.tools = state.tools.map((tool) => {
      if (enabledSet.has(tool.name)) {
        return { ...tool, enabled: true };
      }
      return tool;
    });
    renderTools();
  }
  renderRuntimePolicyHint();
  renderAnalysisPreview();
  renderAtomicToolPreview();
}

function focusRunForm() {
  setView('dashboard');
  q('#run-form')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function applyTemplateFromLibrary(templateId, source = 'template-library') {
  const select = q('#template-select');
  if (!select) return;
  state.templatePinned = true;
  state.templateSource = source;
  select.value = templateId;
  applyTemplateSelection(templateId);
  rememberRecentTemplate(templateId);
  focusRunForm();
  pushEvent(`已快速应用模板：${templateById(templateId)?.name || templateId}`);
}

function applyTemplateInline(templateId) {
  const select = q('#template-select');
  if (!select) return;
  state.templatePinned = true;
  state.templateSource = 'template-inline';
  select.value = templateId;
  applyTemplateSelection(templateId);
  rememberRecentTemplate(templateId);
  pushEvent(`已应用模板但保留当前视图：${templateById(templateId)?.name || templateId}`);
}

function autoSelectPreferredTemplate(notify = false) {
  const select = q('#template-select');
  if (!select || select.options.length === 0) return;
  const preferredId = preferredTemplateId();
  if (!preferredId) return;
  if (state.templatePinned && templateById(select.value)) return;
  if (select.value !== preferredId) {
    select.value = preferredId;
  }
  state.templateSource = 'auto';
  applyTemplateSelection(preferredId);
  if (notify && state.analysisPreview) {
    pushEvent(`已按当前任务自动切换模板：${templateById(preferredId)?.name || preferredId}`);
  }
}

function getFilteredModels() {
  const { query, provider, region, readiness } = state.modelFilters;
  return state.models.filter((model) => {
    const haystack = `${model.name} ${model.provider} ${model.tags.join(' ')}`.toLowerCase();
    if (query && !haystack.includes(query)) {
      return false;
    }
    if (provider !== 'all' && model.provider !== provider) {
      return false;
    }
    if (region !== 'all' && model.region !== region) {
      return false;
    }
    if (readiness === 'ready' && !model.configured) {
      return false;
    }
    if (readiness === 'missing' && model.configured) {
      return false;
    }
    return true;
  });
}

function renderModelSelectorSummary(root, models) {
  if (!root) return;
  const providers = new Set(models.map((model) => model.provider)).size;
  const ready = models.filter((model) => model.configured).length;
  const domestic = models.filter((model) => model.region === 'domestic').length;
  root.innerHTML = `
    <div class="selector-summary-card">
      <div>
        <strong>${models.length}</strong>
        <span>当前可用模型</span>
      </div>
      <div>
        <strong>${providers}</strong>
        <span>覆盖平台</span>
      </div>
      <div>
        <strong>${ready}</strong>
        <span>已完成接入识别</span>
      </div>
      <div>
        <strong>${domestic}</strong>
        <span>国内模型候选</span>
      </div>
      <div>
        <strong>低限制优先</strong>
        <span>排序按接入门槛 / 稳定性 / 任务贴合度递增</span>
      </div>
    </div>
  `;
}

function renderLocalModelFilterOptions() {
  const select = q('#local-tier-filter');
  if (!select || !state.system) return;
  const current = state.localModelFilters.tier;
  const tiers = Array.from(new Set((state.system.builtinModelPacks || []).map((pack) => pack.sizeTier).filter(Boolean)));
  tiers.sort((left, right) => localTierOrder.indexOf(left) - localTierOrder.indexOf(right));
  select.innerHTML = '<option value="all">全部体量</option>';
  tiers.forEach((tier) => {
    const option = document.createElement('option');
    option.value = tier;
    option.textContent = tier;
    option.selected = tier === current;
    select.appendChild(option);
  });
  syncLocalModelFilterControls();
}

function getFilteredLocalModelPacks() {
  const packs = state.system?.builtinModelPacks || [];
  const { query, tier, install, compat, deploy, focus, capability, sort } = state.localModelFilters;
  const filtered = packs.filter((pack) => {
    const haystack = `${pack.name} ${pack.provider} ${pack.modelName} ${pack.version} ${pack.variant}`.toLowerCase();
    if (query && !haystack.includes(query)) {
      return false;
    }
    if (tier !== 'all' && pack.sizeTier !== tier) {
      return false;
    }
    if (install !== 'all' && pack.installState !== install) {
      return false;
    }
    if (compat === 'allowed' && pack.policyMode !== 'on-demand') {
      return false;
    }
    if (compat === 'recommended' && !pack.recommended) {
      return false;
    }
    if (deploy !== 'all' && localDeploymentMeta(pack).key !== deploy) {
      return false;
    }
    if (capability !== 'all' && !localCapabilityTags(pack).includes(capability)) {
      return false;
    }
    return true;
  });
  const ranked = getLowFrictionRankedLocalModelPacks(filtered);
  if (focus === 'flagship') {
    return ranked.filter((pack) => isFlagshipLowFrictionPack(pack));
  }
  if (focus === 'top-12') {
    return ranked.slice(0, 12);
  }
  if (focus === 'starter') {
    return ranked.filter((pack) => numericTierValue(pack.sizeTier) <= 8 && (pack.allowed || pack.recommended));
  }
  ranked.sort((left, right) => {
    if (sort === 'size-asc') {
      return numericTierValue(left.sizeTier) - numericTierValue(right.sizeTier);
    }
    if (sort === 'size-desc') {
      return numericTierValue(right.sizeTier) - numericTierValue(left.sizeTier);
    }
    return compareLocalModelLowFriction(left, right);
  });
  return ranked;
}

function renderLocalModelSummary() {
  const root = q('#local-model-summary');
  if (!root || !state.system) return;
  const filtered = getFilteredLocalModelPacks();
  const stats = getLocalModelPolicyStats(filtered);
  const capabilityCounts = Object.entries(localCapabilityLabels)
    .map(([key, label]) => ({ key, label, count: filtered.filter((pack) => localCapabilityTags(pack).includes(key)).length }))
    .filter((item) => item.count > 0)
    .sort((left, right) => right.count - left.count)
    .slice(0, 3);
  const profile = state.system.systemProfile || {};
  const runtimePolicy = state.system.runtimePolicy || {};
  const localNarrative = buildLocalModelPolicyNarrative(runtimePolicy, stats);
  const primaryDirection = primaryGovernanceDirection(stats);
  const suggested = filtered
    .filter((pack) => pack.policyMode === 'on-demand')
    .slice(0, 3)
    .map((pack) => `<span class="badge">${pack.name}</span>`)
    .join('');
  const filterSummary = localFilterSummaryBadges();
  root.innerHTML = `
    <div class="selector-summary-card local-summary-card">
      <div>
        <strong>${stats.total}</strong>
        <span>当前可见候选</span>
      </div>
      <div>
        <strong>${stats.ready}</strong>
        <span>已就绪</span>
      </div>
      <div>
        <strong>${stats.onDemand}</strong>
          <span>按需启用</span>
      </div>
      <div>
        <strong>${stats.externallyManaged}</strong>
        <span>外部托管</span>
      </div>
      <div>
        <strong>${stats.recommended}</strong>
        <span>当前推荐</span>
      </div>
      <div>
        <strong>${stats.displayOnly}</strong>
        <span>仅展示</span>
      </div>
      <div>
        <strong>${profile.cpuCores || '-'}C / ${profile.memoryMB || '-'}MB</strong>
        <span>当前机器规格</span>
      </div>
    </div>
    <div class="panel-subtle local-machine-note">
      <div class="policy-snapshot">
        <div>
          <span>当前机器口径</span>
          <strong>${runtimePolicy.allowLocalModels ? '按需启用优先' : '仅展示优先'}</strong>
        </div>
        <div>
          <span>当前说明</span>
          <strong>${localNarrative.summary}</strong>
        </div>
        <div>
          <span>当前重点</span>
          <strong>${primaryDirection}</strong>
        </div>
      </div>
      <div class="list-row split-row">
        <div>
          <strong>当前机器适配摘要</strong>
          <p>机器级别 ${profile.tier || '-'}，运行模式 ${runtimeProfileLabel(runtimePolicy.profile)}。任务入口与本地模型库共用同一套判断方式，优先区分按需启用、外部托管与仅展示三类方式。</p>
        </div>
        <div class="stage-list">${suggested || '<span class="badge">当前筛选下暂无可直接推荐候选</span>'}</div>
      </div>
      <div class="stage-list local-filter-summary-strip">${filterSummary.map((item) => `<span class="badge">${item}</span>`).join('') || '<span class="badge">当前为默认筛选视图</span>'}</div>
      <div class="stage-list">${localNarrative.badges.map((item) => `<span class="badge">${item}</span>`).join('')}</div>
      <div class="stage-list local-capability-strip">${capabilityCounts.map((item) => `<span class="badge">${item.label} ${item.count}</span>`).join('') || '<span class="badge">当前筛选下暂无能力标签聚类</span>'}</div>
    </div>
  `;
}

function renderHeroMetrics() {
  const root = q('#hero-metrics');
  if (!root) return;
  const readyProviders = (state.system?.providers || []).filter((provider) => provider.configured).length;
  const totalProviders = (state.system?.providers || []).length;
  const workflowCount = (state.system?.workflowOptions || []).reduce((sum, group) => sum + (group.options || []).length, 0);
  const localStats = getLocalModelPolicyStats(state.system?.builtinModelPacks || []);
  const primaryDirection = primaryGovernanceDirection(localStats);
  root.innerHTML = `
    <div class="metric-card">
      <strong>${readyProviders} / ${totalProviders || '--'}</strong>
      <span>接入平台</span>
    </div>
    <div class="metric-card">
      <strong>${state.models.length || '--'}</strong>
      <span>可用模型</span>
    </div>
    <div class="metric-card">
      <strong>${workflowCount || '--'}</strong>
      <span>任务方案</span>
    </div>
    <div class="metric-card">
      <strong>${primaryDirection}</strong>
      <span>当前重点</span>
    </div>
  `;
}

function renderCapabilityAtlas() {
  const root = q('#capability-atlas');
  if (!root || !state.system) return;
  const workflowGroups = state.system.workflowOptions || [];
  const localStats = getLocalModelPolicyStats(state.system?.builtinModelPacks || []);
  const primaryDirection = primaryGovernanceDirection(localStats);
  const toolsByCategory = (state.system.toolCatalog || []).reduce((acc, tool) => {
    const key = tool.category || 'general';
    if (!acc[key]) {
      acc[key] = [];
    }
    acc[key].push(tool);
    return acc;
  }, {});
  const highlightedFlows = workflowGroups
    .slice(0, 4)
    .map((group) => `
      <div class="list-card capability-card">
        <div class="list-row split-row">
          <div>
            <h3>${group.name}</h3>
            <p>${group.description}</p>
          </div>
          <span class="badge">${(group.options || []).length} 项</span>
        </div>
        <div class="stage-list">${(group.options || []).slice(0, 4).map((option) => `<span class="badge">${option.name}</span>`).join('')}</div>
      </div>
    `)
    .join('');
  const categoryCards = Object.entries(toolsByCategory)
    .map(([category, tools]) => `
      <div class="list-card capability-card">
        <div class="list-row split-row">
          <strong>${toolCategoryLabel(category)}</strong>
          <span class="badge">${tools.length} 项</span>
        </div>
        <p>${tools.slice(0, 3).map((tool) => toolLabels[tool.name] || tool.name).join(' / ')}</p>
        <p class="run-meta">${tools.slice(0, 2).map((tool) => tool.description).join(' · ')}</p>
      </div>
    `)
    .join('');
  const testCards = (state.system.testCatalog || [])
    .slice(0, 4)
    .map((item) => `
      <div class="list-card capability-card compact-card">
        <div class="list-row split-row">
          <strong>${item.name}</strong>
          <span class="badge">${testLayerLabel(item.layer)}</span>
        </div>
        <p>${item.description}</p>
      </div>
    `)
    .join('');

  root.className = 'capability-atlas';
  root.innerHTML = `
    <div class="atlas-intro panel-subtle">
      <div>
        <strong>统一任务中枢</strong>
        <p>星境会优先把模型接入、能力调用、验证与运维建议整理成一条清晰流程，而不是堆砌孤立工具。</p>
      </div>
      <div class="stage-list">
        <span class="badge completed">${primaryDirection}</span>
        <span class="badge">代码与工程</span>
        <span class="badge">办公与研究</span>
        <span class="badge">内容与资产</span>
        <span class="badge">交付与运维</span>
      </div>
    </div>
    <div class="capability-grid">
      <section>
        <div class="section-kicker">核心流程</div>
        <div class="stack-list">${highlightedFlows}</div>
      </section>
      <section>
        <div class="section-kicker">能力簇</div>
        <div class="stack-list">${categoryCards}</div>
      </section>
      <section>
        <div class="section-kicker">验证清单</div>
        <div class="stack-list">${testCards}</div>
      </section>
    </div>
  `;
}

function renderModelProviderFilter() {
  const select = q('#model-provider-filter');
  if (!select) return;
  const current = state.modelFilters.provider;
  const providers = Array.from(new Set(state.models.map((model) => model.provider))).sort((a, b) => a.localeCompare(b));
  select.innerHTML = '<option value="all">全部平台</option>';
  providers.forEach((provider) => {
    const option = document.createElement('option');
    option.value = provider;
    option.textContent = providerLabel(provider);
    option.selected = provider === current;
    select.appendChild(option);
  });
}

function openModelModal(modelID) {
  state.selectedModelId = modelID;
  renderModelModal();
}

function closeModelModal() {
  state.selectedModelId = null;
  renderModelModal();
}

function renderModelModal() {
  const shell = q('#model-modal');
  const title = q('#model-modal-title');
  const body = q('#model-modal-body');
  if (!shell || !title || !body) return;
  const model = state.models.find((item) => item.id === state.selectedModelId);
  if (!model) {
    shell.classList.add('hidden');
    shell.setAttribute('aria-hidden', 'true');
    return;
  }
  shell.classList.remove('hidden');
  shell.setAttribute('aria-hidden', 'false');
  const modelVersion = model.version || model.id.split(':').slice(1).join(':');
  const fullName = model.fullName || `${providerLabel(model.provider)} / ${model.name} / ${modelVersion}`;
  const scoreLabels = fitScoreLabels();
  title.textContent = `${model.name} · ${providerLabel(model.provider)}`;
  body.innerHTML = `
    <div class="list-card">
      <div class="list-row">
        <span class="badge">${regionLabel(model.region)}</span>
        <span class="badge ${model.configured ? 'completed' : 'paused'}">${model.configured ? '已完成接入识别' : '待补充接入'}</span>
      </div>
      <p><strong>${fullName}</strong></p>
      <p>${model.tags.join(' / ')}</p>
      <p class="run-meta">使用参考：${scoreLabels.filter} ${model.filterScore} · ${scoreLabels.review} ${model.reviewScore} · ${scoreLabels.alignment} ${model.alignmentScore}</p>
    </div>
    <div class="list-card">
      <div class="list-row"><strong>接入方式</strong><span class="badge">自动识别</span></div>
      <p>系统会自动检查环境中是否存在对应平台的接入变量。变量一旦可用，模型就会直接进入可接入状态。</p>
      <p class="run-meta">接入变量名 ${model.credentialEnv || '待补充'} · 推荐地址 ${model.recommendedBaseUrl || '沿用官方默认'}</p>
      <div class="actions compact-actions">
        <button type="button" class="ghost-button" data-copy="env">复制变量名</button>
        <button type="button" class="ghost-button" data-copy="base">复制推荐地址</button>
        <a class="button-link" href="${model.website}" target="_blank" rel="noreferrer">前往官网</a>
      </div>
    </div>
    <div class="list-card">
      <div class="list-row"><strong>接入建议</strong><span class="badge">极简流程</span></div>
      <p>${model.configured ? '当前环境已识别接入条件，可直接加入任务流程。' : '当前环境尚未识别接入条件。补充对应变量后刷新页面，系统会自动完成识别，无需重复登记。'}</p>
    </div>
  `;
  body.querySelectorAll('[data-copy]').forEach((button) => {
    button.addEventListener('click', async () => {
      const value = button.dataset.copy === 'env' ? model.credentialEnv : model.recommendedBaseUrl;
      if (!value) return;
      try {
        await navigator.clipboard.writeText(value);
        pushEvent(`已复制${button.dataset.copy === 'env' ? '变量名' : '推荐地址'}，可用于 ${model.name}`);
      } catch (error) {
        pushEvent(`复制失败，请稍后再试：${error.message}`);
      }
    });
  });
}

function renderModels() {
  const list = q('#models-list');
  const summary = q('#model-selector-summary');
  const models = getFilteredModels();
  const scoreLabels = fitScoreLabels();
  list.innerHTML = '';
  renderModelSelectorSummary(summary, models);
  models.forEach((model, index) => {
    const card = document.createElement('div');
    card.className = 'list-card selector-card';
    card.innerHTML = `
      <div class="list-row split-row selector-head">
        <div class="stack-tight selector-title-group">
          <div class="selector-rank">#${String(index + 1).padStart(2, '0')}</div>
          <div class="list-row">
            <h3>${model.name}</h3>
            <span class="badge">${providerLabel(model.provider)}</span>
            <span class="badge">${regionLabel(model.region)}</span>
            <span class="badge ${model.configured ? 'completed' : 'paused'}">${model.configured ? '已完成接入识别' : '待接入'}</span>
          </div>
          <p class="selector-fullname">${model.fullName || `${providerLabel(model.provider)} / ${model.name} / ${model.version || model.id.split(':').slice(1).join(':')}`}</p>
          <p>${model.tags.join(' / ')}</p>
        </div>
        <div class="score-strip">
          <div><span>${scoreLabels.filter}</span><strong>${model.filterScore}</strong></div>
          <div><span>${scoreLabels.review}</span><strong>${model.reviewScore}</strong></div>
          <div><span>${scoreLabels.alignment}</span><strong>${model.alignmentScore}</strong></div>
        </div>
      </div>
      <p class="run-meta">接入变量名 ${model.credentialEnv || '待补充'} · 推荐地址 ${model.recommendedBaseUrl || '沿用官方默认'}</p>
      <div class="selector-footer">
        <div class="list-row selector-state">
          <span class="status-dot ${model.configured ? 'online' : 'offline'}"></span>
          <span>${model.configured ? '当前环境已检测到该平台接入条件，可直接参与调度。' : '当前环境尚未检测到接入条件，补充后会自动识别。'}</span>
        </div>
        <div class="actions compact-actions">
          <button type="button" class="ghost-button" data-action="detail">接入详情</button>
          <a class="button-link" href="${model.website}" target="_blank" rel="noreferrer">官网</a>
        </div>
      </div>
    `;
    card.querySelector('[data-action="detail"]').onclick = () => openModelModal(model.id);
    list.appendChild(card);
  });
  if (models.length === 0) {
    list.innerHTML = '<div class="detail-empty">当前筛选下没有匹配模型，建议放宽条件后再查看。</div>';
  }
}

function renderRuns() {
  const list = q('#runs-list');
  list.innerHTML = '';
  const nextDeploymentSignatures = {};
  state.runs.forEach((run) => {
    const card = document.createElement('div');
    const stages = run.stages.map((stage) => `<span class="badge ${stage.status}">${stageLabel(stage.kind)} · ${statusLabel(stage.status)}</span>`).join('');
    const template = runTemplate(run);
    const templateLabel = runTemplateLabel(run);
    const bundleLabel = runBundleLabel(run);
    const assemblySummary = runAssemblySummary(run);
    const sourceLabel = templateSourceLabel(run.templateSource);
    const deploymentSummary = deriveRunDeploymentSummary(run);
    const failureHeadline = runFailureHeadline(run);
    const failureSummary = runFailureSummary(run);
    const failureContextMeta = runFailureContextMeta(run);
    const failureStatusMeta = runFailureStatusMeta(run);
    const failureTone = unresolvedFailures(run).length ? 'failed' : ((run.failures || []).length ? 'completed' : 'paused');
    const failureHasHistory = Boolean((run.failures || []).length);
    const deploymentSignature = buildRunDeploymentSignature(deploymentSummary);
    const deploymentChanged = Boolean(state.runDeploymentSignatures?.[run.id]) && state.runDeploymentSignatures[run.id] !== deploymentSignature;
    nextDeploymentSignatures[run.id] = deploymentSignature;
    card.className = `list-card task-card task-card-${deploymentSummary.tone}${deploymentChanged ? ' task-card-deployment-updated' : ''}`;
    card.innerHTML = `
      <div class="list-row">
        <h3>${run.title}</h3>
        <span class="badge ${run.status}">${statusLabel(run.status)}</span>
      </div>
      <p>${run.goal.replaceAll('\n', ' ')}</p>
      <div class="stage-list">${stages}</div>
      <div class="stage-list run-route-strip">
        <span class="badge completed">模板 ${templateLabel}</span>
        <span class="badge completed">工具方案 ${bundleLabel}</span>
        ${template?.recommendedBundle?.id === run.analysis?.recommendedBundle?.id ? '<span class="badge completed">模板与预演一致</span>' : '<span class="badge paused">系统自动匹配</span>'}
      </div>
      <p class="run-meta">${assemblySummary}</p>
      <p class="run-meta">意图 ${analysisIntentLabel(run.analysis?.intent)} · 模板来源 ${sourceLabel} · 测试 ${run.selectedTests?.length || run.analysis?.recommendedTests?.length || 0} 项</p>
      <button type="button" class="run-deployment-strip run-deployment-strip-${deploymentSummary.tone}" data-action="view-deployment">
        <div class="list-row split-row">
          <strong>${deploymentSummary.title}</strong>
          <span class="badge ${deploymentSummary.tone}">${deploymentSummary.progressLabel}</span>
        </div>
        <div class="run-deployment-progress"><span style="width: ${deploymentSummary.progressValue}%;"></span></div>
        <p class="run-meta">${deploymentSummary.summary}</p>
      </button>
      <div class="list-card quality-blocker-card quality-blocker-${failureTone}">
        <div class="list-row split-row">
          <strong>失败摘要</strong>
          <span class="badge ${failureTone}">${failureHeadline}</span>
        </div>
        <p>${failureSummary}</p>
        ${failureHasHistory ? `<p class="run-meta">${failureContextMeta}</p>` : ''}
        <p class="run-meta">${failureStatusMeta}</p>
      </div>
      <p class="run-meta">结构状态 ${runOffsetHint(run)}</p>
      <div class="run-actions">
        <button data-action="view">查看详情</button>
        <button data-action="pause">暂停</button>
        <button data-action="resume">恢复</button>
        <button data-action="requirement">补充需求</button>
      </div>
    `;
    const viewButton = card.querySelector('[data-action="view"]');
    const pauseButton = card.querySelector('[data-action="pause"]');
    const resumeButton = card.querySelector('[data-action="resume"]');
    const requirementButton = card.querySelector('[data-action="requirement"]');
    viewButton.onclick = () => selectRun(run.id);
    card.querySelector('[data-action="view-deployment"]')?.addEventListener('click', () => {
      setView('runs');
      selectRun(run.id);
      window.setTimeout(() => focusRunDeploymentSection(), 40);
    });
    pauseButton.onclick = () => mutateRun(run.id, 'pause');
    resumeButton.onclick = () => mutateRun(run.id, 'resume');
    requirementButton.onclick = async () => {
      const extra = window.prompt('请输入补充需求');
      if (!extra) return;
      await mutateRun(run.id, 'requirements', { extra });
    };
    list.appendChild(card);
  });
  state.runDeploymentSignatures = nextDeploymentSignatures;
  if (state.selectedRunId) {
    renderRunDetail();
  }
  renderDevRunBinding();
}

function renderRunDetail() {
  const target = q('#run-detail');
  const run = state.runs.find((item) => item.id === state.selectedRunId);
  if (!run) {
    target.className = 'detail-empty';
    target.textContent = '请选择一项任务，查看阶段进展、验证结果与运行记录。';
    return;
  }
  target.className = 'run-detail-layout';
  const stageItems = run.stages.map((stage) => `<div class="list-card"><div class="list-row"><strong>${stageLabel(stage.kind)}</strong><span class="badge ${stage.status}">${statusLabel(stage.status)}</span></div><p>${stage.summary || '等待开始'}</p></div>`).join('');
  const testItems = (run.testReports || []).map((report) => `<div class="list-card"><div class="list-row"><strong>${report.name}</strong><span class="badge">${testLayerLabel(report.layer)}</span><span class="badge ${report.status}">${report.status === 'passed' ? '已通过' : report.status}</span></div><p>${report.summary}</p><p class="run-meta">耗时 ${Math.round((report.duration || 0) / 1000000)} ms</p></div>`).join('');
  const failureGroups = splitFailureGroups(run.failures || []);
  const validationFailureItems = renderFailureCards(failureGroups.validation);
  const deliveryFailureItems = renderFailureCards(failureGroups.delivery);
  const repairAdviceItems = renderFailureCards(dedupeFailureAdvice(failureGroups.advice), 'advice');
  const resolvedFailureItems = renderFailureCards(failureGroups.resolved, 'resolved');
  const policyDecisionItems = (run.policyDecisions || []).map((decision) => `<div class="list-card policy-decision-card"><div class="list-row"><strong>${policyDecisionAreaLabel(decision.area)}</strong><span class="badge">${policyDecisionActionLabel(decision.action)}</span><span class="badge">${policyDecisionReasonLabel(decision.reason)}</span></div><p>${decision.message}</p></div>`).join('');
  const logItems = (run.logs || []).slice(-6).reverse().map((entry) => `<div class="list-card"><div class="list-row"><strong>${stageLabel(entry.stage)}</strong><span class="badge">${runLogLevelLabel(entry.level)}</span></div><p>${entry.message}</p></div>`).join('');
  const devActivityItems = (run.devActivities || []).slice(0, 6).map((item) => `<div class="list-card"><div class="list-row split-row"><strong>${devActivityLabel(item.kind)}</strong><span class="badge ${item.status === 'completed' ? 'completed' : item.status === 'running' ? 'running' : 'paused'}">${item.status || 'pending'}</span></div><p>${devActivitySummary(item)}</p><p class="run-meta">${devActivityMeta(item)}</p></div>`).join('');
  const analysis = run.analysis || {};
  const template = runTemplate(run);
  const recommendedStages = (analysis.recommendedStages || []).map((stage) => `<span class="badge">${stageLabel(stage)}</span>`).join('');
  const recommendedTools = (analysis.recommendedTools || []).map((tool) => `<span class="badge">${toolLabels[tool] || tool}</span>`).join('');
  const recommendedTests = (analysis.recommendedTests || []).map((item) => `<span class="badge">${item}</span>`).join('');
  const loopBlueprint = (run.loopBlueprint || analysis.loopBlueprint || []).map((step, index) => `<div class="list-card"><div class="list-row"><strong>${index + 1}. ${step.name}</strong><span class="badge">步骤 ${index + 1}</span></div><p>${step.summary}</p></div>`).join('');
  const assembledTools = (run.assembledTools || []).map((tool) => `
    <div class="list-card">
      <div class="list-row split-row">
        <strong>${tool.name}</strong>
        <span class="badge ${tool.localFirst ? 'completed' : 'paused'}">${tool.localFirst ? '本地优先' : '备用入口'}</span>
      </div>
      <p>${tool.summary}</p>
      <p class="run-meta">首选入口 ${toolLabels[tool.preferredProvider] || tool.preferredProvider}</p>
      <div class="stage-list">
        <span class="badge">${toolCategoryLabel(tool.category)}</span>
        <span class="badge">${loadTierLabels[tool.loadTier] || tool.loadTier}</span>
        <span class="badge ${tool.recommended ? 'completed' : 'paused'}">${tool.recommended ? '推荐使用' : '备用入口'}</span>
      </div>
      ${(tool.coverageTags || []).length ? `<p class="run-meta">覆盖范围：${tool.coverageTags.join(' / ')}</p>` : ''}
      <div class="stage-list">${(tool.stageKinds || []).map((stage) => `<span class="badge">${stageLabel(stage)}</span>`).join('')}</div>
    </div>
  `).join('');
  const riskItems = (analysis.risks || []).map((risk) => `<div class="list-card"><p>${risk}</p></div>`).join('');
  const templateStageItems = (template?.defaultStages || []).map((stage) => `<span class="badge">${stageLabel(stage)}</span>`).join('');
  const templateToolItems = (template?.defaultTools || []).map((tool) => `<span class="badge">${toolNameBadge(tool)}</span>`).join('');
  const actualStageItems = (run.stages || []).map((stage) => `<span class="badge ${stage.status}">${stageLabel(stage.kind)}</span>`).join('');
  const actualAssemblyItems = (run.assembledTools || []).slice(0, 6).map((tool) => `<span class="badge ${tool.localFirst ? 'completed' : 'paused'}">${toolLabels[tool.name] || tool.name}</span>`).join('');
  const templateOffsetNotes = templateOffsetItems(run, template);
  const templateOffsets = templateOffsetNotes.map((item) => `<div class="list-card compact-card"><p>${item}</p></div>`).join('');
  const templateSource = templateSourceLabel(run.templateSource);
  const checkpoints = analysis.checkpoints || [];
  const checkpointSummary = checkpointStatusSummary(checkpoints);
  const qualityBlocker = deriveQualityBlocker(run, checkpoints);
  const deploymentOverview = deriveDeploymentOverview(run, checkpoints);
  const qualityBlockerQuickAction = deriveQualityBlockerQuickAction(run, checkpoints, deploymentOverview);
  const deploymentQuickActions = deriveDeploymentQuickActions(run, deploymentOverview);
  const deploymentDependencies = deriveDeploymentDependencies(run, checkpoints);
  const deploymentTarget = deriveDeploymentTargetProfile(deploymentOverview.effectiveRecord?.target);
  const rollbackSuggestionAction = deriveRollbackSuggestionAction(deploymentOverview);
  const showDeploymentTimelinePreviewAction = shouldShowDeploymentTimelinePreviewAction(deploymentOverview);
  const deploymentTimelinePreviewLabel = deploymentTimelinePreviewActionLabel(deploymentOverview);
  const deploymentItems = (run.deployments || []).slice().reverse().map((record, index) => `
    <div class="deployment-timeline-item ${index === 0 ? 'deployment-timeline-item-latest' : ''}" data-deployment-version="${record.version || ''}" ${index === 0 ? 'data-deployment-latest="true"' : ''}>
      <div class="deployment-timeline-line"></div>
      <div class="deployment-timeline-dot deployment-timeline-dot-${deploymentStatusClass(record.status)}"></div>
      <div class="list-card deployment-timeline-card deployment-timeline-card-${deploymentStatusClass(record.status)}">
        <div class="list-row split-row">
          <strong>${record.version || '未记录版本'}</strong>
          <span class="badge ${deploymentStatusClass(record.status)}">${deploymentStatusLabel(record.status)}</span>
        </div>
        <p>${record.summary || '已记录一次稳定部署动作。'}</p>
        <div class="stage-list">
          <span class="badge ${deploymentSourceMeta(record).tone}">来源 ${deploymentSourceMeta(record).label}</span>
          <span class="badge">目标 ${deploymentTargetLabel(record.target)}</span>
          <span class="badge">方式 ${deploymentModeLabel(record.mode)}</span>
          ${index === 0 ? `<span class="badge completed">${record.mode === 'rollback' ? '回退生效' : '当前生效'}</span>` : ''}
          ${index === 0 && record.mode === 'revalidate' ? '<span class="badge running">复验已完成</span>' : ''}
        </div>
        <p class="run-meta">时间 ${formatDateTime(record.createdAt)}</p>
        <div class="actions deployment-timeline-actions">
          <button type="button" class="ghost-button" data-action="rollback-deployment" data-version="${record.version || ''}" ${record.version ? '' : 'disabled'}>回退到此版本</button>
          ${index === 0 && showDeploymentTimelinePreviewAction ? `<button type="button" class="ghost-button" data-action="open-preview-from-timeline">${deploymentTimelinePreviewLabel}</button>` : ''}
        </div>
      </div>
    </div>
  `).join('');
  const checkpointItems = checkpoints.map((item) => `
    <div class="list-card checkpoint-card checkpoint-card-${checkpointStatusClass(item.status)}">
      <div class="list-row split-row">
        <strong>${item.title}</strong>
        <span class="badge ${checkpointStatusClass(item.status)}">${checkpointStatusLabel(item.status)}</span>
      </div>
      <p>${item.summary}</p>
      <div class="stage-list">
        <span class="badge">${checkpointTitleTone(item.id)}</span>
      </div>
      <p class="run-meta">检查要求：${item.gate}</p>
    </div>
  `).join('');
  target.innerHTML = `
    <div class="list-card run-detail-hero">
      <div class="list-row split-row">
        <strong>${run.title}</strong>
        <span class="badge ${run.status}">${statusLabel(run.status)}</span>
      </div>
      <p>${run.goal.replaceAll('\n', '<br/>')}</p>
      <p class="run-meta">自动修复 ${run.autoRepairEnabled ? `已开启（${run.autoRepairMode || 'standard'}）` : '关闭'} · 稳定部署 ${run.remoteDeployEnabled ? '已开启' : '关闭'}</p>
    </div>
    <section class="run-detail-section">
      <div class="section-kicker">任务概览</div>
      <div class="list-card run-detail-summary-card"><p>${analysis.summary || '智能预演摘要待补充。'}</p><p class="run-meta">意图 ${analysisIntentLabel(analysis.intent)} · 类型 ${projectKindLabel(analysis.projectKind)}</p><p class="run-meta">建议阶段</p><div class="stage-list">${recommendedStages || '<span class="badge">待补充</span>'}</div><p class="run-meta">建议能力</p><div class="stage-list">${recommendedTools || '<span class="badge">待补充</span>'}</div><div class="list-row split-row"><strong>${(run.selectedTests || []).length ? '当前验证' : '推荐验证'}</strong><span class="badge">${run.testMode || 'template'}</span></div><div class="stage-list">${(run.selectedTests || []).map((item) => `<span class="badge completed">${item}</span>`).join('') || recommendedTests || '<span class="badge">待补充</span>'}</div></div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">模板结构</div>
      <div class="stack-list">
        <div class="list-card template-detail-card">
          <div class="list-row split-row">
            <strong>${template?.name || run.templateId || '未选择模板'}</strong>
            ${template?.recommendedBundle?.name ? `<span class="badge completed">${template.recommendedBundle.name}</span>` : '<span class="badge paused">系统自动匹配</span>'}
          </div>
          <p>${template?.description || '当前运行单尚未绑定可识别的模板结构，系统会按任务分析结果与当前设置自动匹配。'}</p>
          <p class="run-meta">模板来源 ${templateSource} · 结构状态 ${runOffsetHint(run)}</p>
          ${template?.recommendedBundle?.reason ? `<p class="run-meta">${template.recommendedBundle.reason}</p>` : ''}
          <div class="template-compare-grid">
            <div>
              <span>模板默认阶段</span>
              <div class="stage-list">${templateStageItems || '<span class="badge">待补充</span>'}</div>
            </div>
            <div>
              <span>实际执行阶段</span>
              <div class="stage-list">${actualStageItems || '<span class="badge">待开始</span>'}</div>
            </div>
            <div>
              <span>模板默认工具</span>
              <div class="stage-list">${templateToolItems || '<span class="badge">待补充</span>'}</div>
            </div>
            <div>
              <span>实际能力装配</span>
              <div class="stage-list">${actualAssemblyItems || '<span class="badge">待补充</span>'}</div>
            </div>
          </div>
          ${templateOffsetNotes.length ? `<div class="stack-list template-offset-list">${templateOffsets}</div>` : ''}
        </div>
      </div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">质量总览</div>
      <div class="stack-list">
        <div class="list-card quality-blocker-card quality-blocker-${qualityBlocker.tone}">
          <div class="list-row split-row">
            <strong>当前主阻塞项</strong>
            <span class="badge ${qualityBlocker.tone}">${qualityBlocker.title}</span>
          </div>
          <p>${qualityBlocker.summary}</p>
          <p class="run-meta">建议动作：${qualityBlocker.action}</p>
          ${qualityBlockerQuickAction ? `<div class="actions quality-blocker-actions"><button type="button" class="ghost-button" data-action="${qualityBlockerQuickAction.action}" ${qualityBlockerQuickAction.disabled ? 'disabled' : ''}>${qualityBlockerQuickAction.label}</button></div>` : ''}
        </div>
        ${checkpoints.length ? `<div class="checkpoint-summary-grid">
          <div class="list-card checkpoint-summary-card">
            <span>已完成</span>
            <strong>${checkpointSummary.completed}</strong>
          </div>
          <div class="list-card checkpoint-summary-card">
            <span>进行中</span>
            <strong>${checkpointSummary.inProgress}</strong>
          </div>
          <div class="list-card checkpoint-summary-card">
            <span>待开始</span>
            <strong>${checkpointSummary.pending}</strong>
          </div>
          <div class="list-card checkpoint-summary-card">
            <span>需关注</span>
            <strong>${checkpointSummary.attention}</strong>
          </div>
        </div><div class="stack-list">${checkpointItems}</div>` : '<div class="detail-empty">任务分析完成后，将显示质量总览。</div>'}
      </div>
    </section>
    <section class="run-detail-section" id="run-deployment-section">
      <div class="section-kicker">稳定部署</div>
      <div class="stack-list">
        <div class="deployment-summary-grid">
          <div class="list-card deployment-summary-card">
            <span>部署方式</span>
            <strong>${deployModeLabel(state.system?.runtimePolicy?.deployMode)}</strong>
          </div>
          <div class="list-card deployment-summary-card">
            <span>当前状态</span>
            <strong>${deploymentOverview.statusTitle}</strong>
          </div>
          <div class="list-card deployment-summary-card">
            <span>当前生效版本</span>
            <button type="button" class="deployment-version-link" data-action="view-effective-version" ${deploymentOverview.effectiveVersion === '暂无版本' ? 'disabled' : ''}>${deploymentOverview.effectiveVersion}</button>
          </div>
          <div class="list-card deployment-summary-card">
            <span>目标环境</span>
            <strong>${deploymentOverview.effectiveTarget}</strong>
          </div>
        </div>
        <div class="list-card deployment-status-card deployment-status-${deploymentOverview.statusTone}">
          <div class="list-row split-row">
            <strong>当前进展</strong>
            <span class="badge ${deploymentOverview.statusTone}">${deploymentOverview.statusTitle}</span>
          </div>
          <p>${deploymentOverview.statusSummary}</p>
          <div class="actions deployment-status-actions">
            <button type="button" class="ghost-button" data-action="deploy-run" ${run.remoteDeployEnabled ? '' : 'disabled'}>执行稳定部署</button>
          </div>
        </div>
        ${deploymentOverview.latestRecord ? `<div class="list-card deployment-status-card deployment-status-${deploymentOverview.actionTone}">
          <div class="list-row split-row">
            <strong>最近动作</strong>
            <span class="badge ${deploymentOverview.actionTone}">${deploymentOverview.actionTitle}</span>
          </div>
          <p>${deploymentOverview.actionSummary}</p>
        </div>` : ''}
        <div class="list-card deployment-status-card deployment-status-${deploymentOverview.actionTone}">
          <div class="list-row split-row">
            <strong>下一步建议</strong>
            <span class="badge ${deploymentOverview.actionTone}">建议</span>
          </div>
          <p>${deploymentQuickActions.length ? `当前可直接执行：${deploymentQuickActions.map((item) => item.label).join(' / ')}。` : '当前暂无可直接执行的建议动作。'}</p>
          <div class="actions deployment-status-actions">
            ${deploymentQuickActions.map((item) => `<button type="button" class="ghost-button" data-action="${item.action}" ${item.disabled ? 'disabled' : ''}>${item.label}</button>`).join('')}
          </div>
        </div>
        <div class="list-card deployment-target-card">
          <div class="list-row split-row">
            <strong>上线环境</strong>
            <span class="badge ${deploymentOverview.effectiveRecord ? 'completed' : 'paused'}">${deploymentTarget.title}</span>
          </div>
          <p>${deploymentTarget.summary}</p>
          <div class="deployment-target-grid">
            ${deploymentTarget.items.map((item) => `<div class="list-card compact-card deployment-readiness-card deployment-readiness-${item.status}"><div class="list-row split-row"><strong>${item.label}</strong><span class="badge ${item.status}">${item.value}</span></div></div>`).join('')}
          </div>
        </div>
        <div class="stack-list">
          <div class="list-row split-row">
            <strong>发布条件</strong>
            <span class="badge">条件</span>
          </div>
          <div class="deployment-dependency-grid">
            ${deploymentDependencies.map((item) => `<div class="list-card compact-card deployment-dependency-card deployment-dependency-${item.status}"><div class="list-row split-row"><strong>${item.label}</strong><span class="badge ${item.status}">${item.status === 'completed' ? '已满足' : item.status === 'running' ? '进行中' : item.status === 'failed' ? '需关注' : '待补齐'}</span></div><p class="run-meta">${item.summary}</p></div>`).join('')}
          </div>
        </div>
        <div class="stack-list">
          <div class="list-row split-row">
            <strong>部署准备</strong>
            <span class="badge">进度</span>
          </div>
          <div class="deployment-readiness-grid">
            ${deploymentOverview.readinessItems.map((item) => `<div class="list-card compact-card deployment-readiness-card deployment-readiness-${item.status}"><div class="list-row split-row"><strong>${item.label}</strong><span class="badge ${item.status}">${item.value}</span></div></div>`).join('')}
          </div>
        </div>
        <div class="stack-list">
          <div class="list-row split-row">
            <strong>版本记录</strong>
            <span class="badge">记录</span>
          </div>
          <div class="deployment-timeline">${deploymentItems || '<div class="detail-empty">首次稳定部署完成后，将显示版本记录。</div>'}</div>
        </div>
        ${deploymentOverview.latestRecord ? `<div class="list-card deployment-rollback-card">
          <div class="list-row split-row">
            <strong>回退建议</strong>
            <span class="badge completed">${deploymentOverview.rollbackTitle}</span>
          </div>
          <p>${deploymentOverview.rollbackSummary}</p>
          <p class="run-meta">建议顺序：回退版本 -> 完成复验 -> 查看预览。</p>
          ${rollbackSuggestionAction ? `<div class="actions deployment-rollback-actions"><button type="button" class="ghost-button" data-action="${rollbackSuggestionAction.action}" ${rollbackSuggestionAction.disabled ? 'disabled' : ''}>${rollbackSuggestionAction.label}</button></div>` : ''}
        </div>` : ''}
      </div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">风险提示</div>
      <div class="stack-list">${riskItems || '<div class="detail-empty">当前暂无风险提醒。</div>'}</div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">执行路径</div>
      <div class="stack-list">${loopBlueprint || '<div class="detail-empty">执行路径待补充。</div>'}</div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">能力装配</div>
      <div class="stack-list">${assembledTools || '<div class="detail-empty">当前暂无装配结果，请检查阶段与工具开关。</div>'}</div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">执行轨迹</div>
      <div class="stack-list">${stageItems}</div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">验证结果</div>
      <div class="stack-list">${testItems || '<div class="detail-empty">验证完成后，将显示结果。</div>'}</div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">关注事项</div>
      <div class="stack-list">
        <div>
          <div class="list-row split-row"><strong>验证失败</strong><span class="badge ${failureGroups.validation.length ? 'failed' : 'completed'}">${summarizeFailureGroup(failureGroups.validation.length, '待处理', '暂无问题')}</span></div>
          <div class="stack-list">${validationFailureItems || '<div class="detail-empty">当前没有待处理项。</div>'}</div>
        </div>
        <div>
          <div class="list-row split-row"><strong>部署失败</strong><span class="badge ${failureGroups.delivery.length ? 'failed' : 'completed'}">${summarizeFailureGroup(failureGroups.delivery.length, '待处理', '暂无问题')}</span></div>
          <div class="stack-list">${deliveryFailureItems || '<div class="detail-empty">当前没有待处理项。</div>'}</div>
        </div>
        <div>
          <div class="list-row split-row"><strong>修复建议</strong><span class="badge ${failureGroups.advice.length ? 'running' : 'completed'}">${failureGroups.advice.length ? summarizeFailureGroup(dedupeFailureAdvice(failureGroups.advice).length, '可继续处理', '暂无建议') : '暂无建议'}</span></div>
          <div class="stack-list">${repairAdviceItems || '<div class="detail-empty">当前没有新的处理建议。</div>'}</div>
        </div>
        <div>
          <div class="list-row split-row"><strong>处理记录</strong><span class="badge ${failureGroups.resolved.length ? 'completed' : 'paused'}">${failureGroups.resolved.length ? summarizeFailureGroup(failureGroups.resolved.length, '已保留', '暂无记录') : '暂无记录'}</span></div>
          <div class="stack-list">${resolvedFailureItems || '<div class="detail-empty">当前还没有处理记录。</div>'}</div>
        </div>
      </div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">系统判断</div>
      <div class="stack-list">${policyDecisionItems || '<div class="detail-empty">当前暂无系统判断记录。</div>'}</div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">工作台轨迹</div>
      <div class="stack-list">${devActivityItems || '<div class="detail-empty">当前暂无与该任务关联的文件、终端或预览动作。</div>'}</div>
    </section>
    <section class="run-detail-section">
      <div class="section-kicker">最近动态</div>
      <div class="stack-list">${logItems || '<div class="detail-empty">最近动态待补充。</div>'}</div>
    </section>
  `;
  target.querySelectorAll('[data-action="rollback-deployment"]').forEach((button) => {
    button.addEventListener('click', async () => {
      const version = button.dataset.version;
      if (!version) return;
      const confirmed = window.confirm(`确认回退到版本 ${version} 吗？回退后请先完成复验，再查看预览。`);
      if (!confirmed) return;
      await runDeploymentAction(button, '回退中...', '版本回退失败', async () => {
        await mutateRun(run.id, 'rollback', { version }, { suppressError: true, rethrow: true });
        pushEvent(`已开始回退：${version}`);
      });
    });
  });
  target.querySelectorAll('[data-action="deploy-run"]').forEach((button) => {
    button.addEventListener('click', async (event) => {
      const currentButton = event.currentTarget;
      const confirmed = window.confirm('确认现在执行稳定部署吗？执行后请查看预览并留意服务状态。');
      if (!confirmed) return;
      await runDeploymentAction(currentButton, '部署中...', '稳定部署失败', async () => {
        await mutateRun(run.id, 'deploy', undefined, { suppressError: true, rethrow: true });
        pushEvent('已开始稳定部署');
      });
    });
  });
  target.querySelectorAll('[data-action="revalidate-run"]').forEach((button) => {
    button.addEventListener('click', async (event) => {
      const currentButton = event.currentTarget;
      await runDeploymentAction(currentButton, '复验中...', '复验失败', async () => {
        await mutateRun(run.id, 'revalidate', undefined, { suppressError: true, rethrow: true });
        pushEvent('已开始复验');
      });
    });
  });
  target.querySelectorAll('[data-action="open-preview"]').forEach((button) => {
    button.addEventListener('click', async (event) => {
      const currentButton = event.currentTarget;
      await runDeploymentAction(currentButton, '打开中...', '打开预览失败', async () => {
        await openDeploymentPreview(q('#dev-preview-port')?.value?.trim() || state.dev.previewPort || '8080');
      });
    });
  });
  target.querySelectorAll('[data-action="open-preview-from-timeline"]').forEach((button) => {
    button.addEventListener('click', async (event) => {
      const currentButton = event.currentTarget;
      await runDeploymentAction(currentButton, '打开中...', '打开预览失败', async () => {
        await openDeploymentPreview(q('#dev-preview-port')?.value?.trim() || state.dev.previewPort || '8080');
      });
    });
  });
  target.querySelector('[data-action="view-effective-version"]')?.addEventListener('click', () => {
    focusDeploymentTimelineRecord(deploymentOverview.effectiveVersion === '暂无版本' ? '' : deploymentOverview.effectiveVersion);
  });
}

function renderAnalysisPreview() {
  const target = q('#analysis-preview');
  const analysis = state.analysisPreview;
  if (!analysis) {
    target.className = 'analysis-preview detail-empty';
    target.textContent = '输入一句话需求后，可先看到自动整理的流程、能力与验证建议。';
    return;
  }
  const matchedBundle = previewBundles().find((bundle) => bundle.id === analysis.recommendedBundle?.id) || previewBundles().find((bundle) => bundle.recommended);
  const matchedItems = matchedBundle?.items || [];
  const requirementSpec = analysis.requirementSpec || {};
  const taskQueue = analysis.taskQueue || [];
  const checkpoints = analysis.checkpoints || [];
  const checkpointSummary = checkpointStatusSummary(checkpoints);
  const qualityBlocker = deriveQualityBlocker({ failures: [], testReports: [] }, checkpoints);
  target.className = 'analysis-preview';
  target.innerHTML = `
    <div class="list-card">
      <div class="list-row">
        <strong>${analysisIntentLabel(analysis.intent)}</strong>
        <span class="badge">${projectKindLabel(analysis.projectKind)}</span>
      </div>
      <p>${analysis.summary}</p>
      <p class="run-meta">建议阶段</p>
      <div class="stage-list">${(analysis.recommendedStages || []).map((item) => `<span class="badge">${stageLabel(item)}</span>`).join('')}</div>
      <p class="run-meta">建议能力</p>
      <div class="stage-list">${(analysis.recommendedTools || []).map((item) => `<span class="badge">${item}</span>`).join('')}</div>
      <p class="run-meta">建议验证</p>
      <div class="stage-list">${(analysis.recommendedTests || []).map((item) => `<span class="badge">${item}</span>`).join('')}</div>
      <p class="run-meta">建议方案</p>
      <div class="list-card compact-card">
        <div class="list-row split-row">
          <strong>${analysis.recommendedBundle?.name || '待补充'}</strong>
          ${matchedBundle ? '<span class="badge completed">已可装配</span>' : '<span class="badge paused">待补充</span>'}
        </div>
        <p>${analysis.recommendedBundle?.reason || '系统正在整理更合适的方案。'}</p>
        <div class="stage-list">${matchedItems.slice(0, 4).map((item) => `<span class="badge completed">${item.name}</span>`).join('') || '<span class="badge">将按当前阶段补齐</span>'}</div>
      </div>
      <p class="run-meta">统一闭环</p>
      <div class="stage-list">${(analysis.loopBlueprint || []).map((item, index) => `<span class="badge">${index + 1}. ${item.name}</span>`).join('')}</div>
      <div class="analysis-grid">
        <div>
          <h4>需求规格</h4>
          <div class="analysis-list">
            <div class="list-card compact-card"><p>${requirementSpec.summary || '需求摘要待补充。'}</p></div>
            <div class="list-card compact-card"><p class="run-meta">功能要求</p><div class="stage-list">${(requirementSpec.functionalRequirements || []).map((item) => `<span class="badge">${item}</span>`).join('') || '<span class="badge">待补充</span>'}</div></div>
            <div class="list-card compact-card"><p class="run-meta">技术栈</p><div class="stage-list">${(requirementSpec.techStack || []).map((item) => `<span class="badge">${item}</span>`).join('') || '<span class="badge">沿用现有栈</span>'}</div></div>
          </div>
        </div>
        <div>
          <h4>任务队列</h4>
          <div class="analysis-list">
            ${taskQueue.map((item) => `<div class="list-card compact-card"><div class="list-row split-row"><strong>${item.title}</strong><span class="badge">P${item.priority}</span></div><p>${item.summary}</p><p class="run-meta">${queueAgentLabel(item.agent)} · ${queuePhaseLabel(item.phase)}</p></div>`).join('') || '<div class="detail-empty">任务队列待补充。</div>'}
          </div>
        </div>
        <div>
          <h4>质量总览</h4>
          <div class="analysis-list">
            <div class="list-card compact-card quality-blocker-card quality-blocker-${qualityBlocker.tone}"><div class="list-row split-row"><strong>当前主阻塞项</strong><span class="badge ${qualityBlocker.tone}">${qualityBlocker.title}</span></div><p>${qualityBlocker.summary}</p><p class="run-meta">建议动作：${qualityBlocker.action}</p></div>
            <div class="checkpoint-summary-grid">
              <div class="list-card checkpoint-summary-card"><span>已完成</span><strong>${checkpointSummary.completed}</strong></div>
              <div class="list-card checkpoint-summary-card"><span>进行中</span><strong>${checkpointSummary.inProgress}</strong></div>
              <div class="list-card checkpoint-summary-card"><span>待开始</span><strong>${checkpointSummary.pending}</strong></div>
              <div class="list-card checkpoint-summary-card"><span>需关注</span><strong>${checkpointSummary.attention}</strong></div>
            </div>
            ${checkpoints.map((item) => `<div class="list-card compact-card checkpoint-card checkpoint-card-${checkpointStatusClass(item.status)}"><div class="list-row split-row"><strong>${item.title}</strong><span class="badge ${checkpointStatusClass(item.status)}">${checkpointStatusLabel(item.status)}</span></div><p>${item.summary}</p><div class="stage-list"><span class="badge">${checkpointTitleTone(item.id)}</span></div><p class="run-meta">检查要求：${item.gate}</p></div>`).join('') || '<div class="detail-empty">质量总览待补充。</div>'}
          </div>
        </div>
      </div>
    </div>
  `;
}

function renderWorkflowOptions() {
  const root = q('#workflow-options');
  if (!root) return;
  root.innerHTML = '';
  (state.workflowOptions || []).forEach((group) => {
    const card = document.createElement('div');
    card.className = 'list-card';
      const options = (group.options || []).map((option) => `<div class="list-card"><div class="list-row"><strong>${option.name}</strong><span class="badge">${projectKindLabel(option.kind)}</span><span class="badge ${option.defaultOn ? 'completed' : 'paused'}">${option.defaultOn ? '默认开启' : '默认未启用'}</span></div><p>${option.description}</p></div>`).join('');
    card.innerHTML = `<div class="list-row"><h3>${group.name}</h3></div><p>${group.description}</p><div class="stack-list">${options}</div>`;
    root.appendChild(card);
  });
}

function renderFeatureToggles() {
  const root = q('#feature-toggles');
  if (!root || !state.system) return;
  root.innerHTML = '';
  (state.system.featureToggles || []).forEach((toggle) => {
    const card = document.createElement('div');
    card.className = 'list-card';
    card.innerHTML = `
      <div class="list-row">
        <h3>${toggle.name}</h3>
        <span class="badge ${toggle.enabled ? 'completed' : 'paused'}">${toggle.enabled ? '开启' : '关闭'}</span>
        <span class="badge ${toggle.recommended ? 'completed' : 'paused'}">${toggle.recommended ? '推荐' : '可选'}</span>
        <span class="badge ${toggle.allowed ? 'completed' : 'failed'}">${toggle.allowed ? '允许' : '受限'}</span>
      </div>
      <p>${toggle.description}</p>
      <p class="run-meta">默认 ${toggle.defaultOn ? '开启' : '关闭'}</p>
      ${toggle.warning ? `<div class="warning-block">${toggle.warning}</div>` : ''}
      <div class="actions"><button type="button" ${!toggle.allowed && !toggle.enabled ? 'disabled' : ''}>${toggle.enabled ? '关闭' : '开启'}</button></div>
    `;
    card.querySelector('button').onclick = async () => {
      try {
        await fetchJSON(`/api/settings/toggles/${toggle.id}`, { method: 'POST', body: JSON.stringify({ enabled: !toggle.enabled }) });
        await loadSystemSummary();
        pushEvent(`已更新功能设置：${toggle.name}`);
      } catch (error) {
        pushEvent(`功能设置更新失败，请稍后重试：${error.message}`);
      }
    };
    root.appendChild(card);
  });
}

function renderAtomicTools() {
  const root = q('#atomic-tools');
  const invocationsRoot = q('#tool-invocations');
  if (!root || !state.system) return;
  root.innerHTML = '';
  const registry = state.system.atomicTools || [];
  const categorySummary = registry.reduce((acc, tool) => {
    const key = tool.category || 'general';
    acc[key] = (acc[key] || 0) + 1;
    return acc;
  }, {});
  const categoryStrip = Object.entries(categorySummary)
    .sort((left, right) => right[1] - left[1])
    .map(([category, count]) => `<span class="badge">${toolCategoryLabel(category)} ${count}</span>`)
    .join('');
  if (categoryStrip) {
    root.innerHTML = `<div class="list-card"><div class="list-row split-row"><strong>覆盖概览</strong><span class="badge">${registry.length} 项能力</span></div><p>功能总览按全类别展开，但同类能力只保留本地优先、最强入口和必要备用入口。</p><div class="stage-list">${categoryStrip}</div></div>`;
  }
  registry.forEach((tool) => {
    const card = document.createElement('div');
    card.className = 'list-card';
    const stages = (tool.stageKinds || []).map((stage) => `<span class="badge">${stageLabel(stage)}</span>`).join('');
    const fallbacks = (tool.fallbackProviders || []).map((item) => `<span class="badge">${toolLabels[item] || item}</span>`).join('');
    const coverageSummary = (tool.coverageTags || []).join(' / ');
    card.innerHTML = `
      <div class="list-row split-row">
        <strong>${tool.name}</strong>
        <span class="badge ${capabilityModeClass(tool.activationMode)}">${capabilityModeLabels[tool.activationMode] || tool.activationMode}</span>
      </div>
      <p>${tool.summary}</p>
      <p class="run-meta">首选入口 ${toolLabels[tool.preferredProvider] || tool.preferredProvider} · 已按同类能力整理</p>
      <div class="stage-list">
        <span class="badge">${toolCategoryLabel(tool.category)}</span>
        <span class="badge">${loadTierLabels[tool.loadTier] || tool.loadTier}</span>
        <span class="badge ${tool.localFirst ? 'completed' : 'paused'}">${tool.localFirst ? '本地优先' : '备用入口'}</span>
        <span class="badge ${tool.allowed ? 'completed' : 'failed'}">${tool.allowed ? '当前可用' : '当前受限'}</span>
        <span class="badge ${tool.recommended ? 'completed' : 'paused'}">${tool.recommended ? '推荐入口' : '备用入口'}</span>
      </div>
      <div class="stage-list">${stages}</div>
      <p class="run-meta">${coverageSummary ? `覆盖范围：${coverageSummary}` : '覆盖范围待补充。'}</p>
      <div class="stage-list">${fallbacks || '<span class="badge">无备用入口</span>'}</div>
      <p class="run-meta">${tool.reason}</p>
      <div class="actions"><button type="button" ${tool.allowed ? '' : 'disabled'}>单独调用</button></div>
    `;
    card.querySelector('button').onclick = async () => {
      try {
        const result = await fetchJSON(`/api/tools/${tool.id}/invoke`, { method: 'POST' });
        pushEvent(`已单独调用能力：${result.title}`);
        await loadSystemSummary();
      } catch (error) {
        pushEvent(`能力调用失败：${error.message}`);
      }
    };
    root.appendChild(card);
  });
  if (registry.length === 0) {
    root.innerHTML = '<div class="detail-empty">当前还没有能力注册表。</div>';
  }
  if (!invocationsRoot) return;
  const invocations = state.system.toolInvocations || [];
  invocationsRoot.innerHTML = invocations.map((item) => `
    <div class="list-card">
      <div class="list-row split-row">
        <strong>${item.title}</strong>
        <span class="badge ${checkpointStatusClass(item.status)}">${invocationStatusLabel(item.status)}</span>
      </div>
      <p>${item.summary}</p>
      <p class="run-meta">记录 ${new Date(item.createdAt).toLocaleString()}</p>
    </div>
  `).join('') || '<div class="detail-empty">最近还没有单独调用记录。</div>';
}

function renderModelPacks() {
  const root = q('#model-packs');
  if (!root || !state.system) return;
  root.innerHTML = '';
  renderLocalModelFilterOptions();
  renderLocalModelSummary();
  const packs = getFilteredLocalModelPacks();
  const groupedPacks = packs.reduce((acc, pack) => {
    const key = pack.policyMode || 'display-only';
    if (!acc[key]) {
      acc[key] = [];
    }
    acc[key].push(pack);
    return acc;
  }, {});
  localPolicySections.forEach((group) => {
    const candidates = rankPolicySectionCandidates(group.key, groupedPacks[group.key] || []);
    const section = document.createElement('section');
    section.className = `stack-list local-model-tier local-policy-section local-policy-section-${group.key}`;
    const downloaded = candidates.filter((item) => item.downloaded).length;
    const ready = candidates.filter((item) => item.installState === 'ready').length;
    const recommended = candidates.filter((item) => item.recommended).length;
    const dominantTarget = dominantHostingTarget(candidates);
    const adviceSnapshot = state.localPolicyAdvice[group.key];
    const adviceMatchesFilters = adviceSnapshot && adviceSnapshot.filterSignature === localFilterSignature();
    const governanceSummary = policyGovernanceSummary(group.key, candidates);
    const recommendedBadges = group.key === 'on-demand' ? groupRecommendedBadges(candidates) : '';
    const reasonBadges = group.key === 'display-only'
      ? summarizePolicyReasons(candidates).map((item) => `<span class="badge">${item}</span>`).join('')
      : '';
    const tiers = Array.from(new Set(candidates.map((item) => item.sizeTier).filter(Boolean)))
      .sort((left, right) => numericTierValue(left) - numericTierValue(right))
      .map((tier) => `<span class="badge">${tier}</span>`)
      .join('');
    const emptyMessage = group.key === 'on-demand'
      ? '当前筛选下暂无适合按需启用的候选。建议重置筛选，或先查看仅展示与外部托管分组。'
      : group.key === 'externally-managed'
        ? '当前筛选下暂无适合外部托管的候选。如需查看统一托管方案，请先放宽筛选条件。'
        : '当前筛选下暂无仅展示候选，说明当前视图更偏向可直接启用或外部托管的模型。';
    section.innerHTML = `
      <div class="list-row split-row tier-head local-policy-head">
        <div>
          <h3>${group.title}</h3>
          <p class="run-meta">${group.description}</p>
        </div>
        <div class="stage-list">
          <span class="badge ${localPolicyModeClass(group.key)}">${candidates.length} 个候选</span>
          <span class="badge">${ready} 已就绪</span>
          <span class="badge">${downloaded} 已下载</span>
          <span class="badge">${candidates.length - downloaded} 未下载</span>
          ${group.key === 'on-demand' ? `<span class="badge completed">${recommended} 推荐候选</span>` : ''}
          ${group.key === 'externally-managed' ? `<span class="badge queued">主建议 ${hostingModeLabel(dominantTarget)}</span>` : ''}
        </div>
      </div>
      <div class="stage-list local-policy-tiers">${tiers || '<span class="badge">体量待补充</span>'}</div>
      <div class="local-policy-note local-policy-governance-note"><div class="list-row"><strong>重点摘要</strong><span class="badge ${localPolicyModeClass(group.key)}">${localPolicyModeLabels[group.key] || group.key}</span></div><p>${governanceSummary}</p></div>
      ${adviceSnapshot ? `<div class="local-policy-note"><div class="list-row"><strong>最近托管建议</strong><div class="stage-list">${adviceSnapshot.hostingMode ? `<span class="badge queued">${hostingModeLabel(adviceSnapshot.hostingMode)}</span>` : ''}<span class="badge ${adviceMatchesFilters ? 'completed' : 'paused'}">${adviceMatchesFilters ? '适用于当前筛选' : '来自其他筛选结果'}</span></div></div><p>${adviceSnapshot.summary}</p>${adviceMatchesFilters ? '' : '<p class="run-meta">筛选条件已更新，如需查看新的分组方案，请重新生成托管建议。</p>'}</div>` : ''}
      ${recommendedBadges ? `<div class="stage-list local-policy-strip"><span class="badge completed">推荐候选</span>${recommendedBadges}</div>` : ''}
      ${reasonBadges ? `<div class="stage-list local-policy-strip"><span class="badge paused">限制原因</span>${reasonBadges}</div>` : ''}
      ${group.key === 'externally-managed' ? `<div class="actions local-policy-actions"><button type="button" class="ghost-button" data-action="group-advice">查看分组托管建议</button></div>` : ''}
      ${group.key === 'on-demand' ? `<div class="actions local-policy-actions"><button type="button" class="ghost-button" data-action="group-filter">查看推荐候选</button></div>` : ''}
    `;
    const list = document.createElement('div');
    list.className = 'stack-list';
    if (candidates.length === 0) {
      list.innerHTML = `<div class="detail-empty local-policy-empty">${emptyMessage}</div>`;
    } else {
      candidates.forEach((pack) => {
        list.appendChild(buildLocalModelCard(pack));
      });
    }
    section.querySelector('[data-action="group-advice"]')?.addEventListener('click', async () => {
      try {
        state.advice = await buildExternalHostingAdviceForGroup(candidates);
        state.localPolicyAdvice[group.key] = summarizeAdviceForGroup(state.advice);
        renderModelPacks();
        renderAdvice();
        pushEvent(`已更新分组托管建议，当前主方案为${hostingModeLabel(dominantTarget)}`);
      } catch (error) {
        pushEvent(`分组托管建议暂时无法生成：${error.message}`);
      }
    });
    section.querySelector('[data-action="group-filter"]')?.addEventListener('click', () => {
      state.localModelFilters.compat = 'recommended';
      state.localModelFilters.focus = 'all';
      syncLocalModelFilterControls();
      renderModelPacks();
      pushEvent('已切换到本地推荐候选视图');
    });
    section.appendChild(list);
    root.appendChild(section);
  });
}

function buildLocalModelCard(pack) {
  const deployMeta = localDeploymentMeta(pack);
  const hostingTarget = hostingModeLabel(deployMeta.key);
  const capabilityTags = localCapabilityTags(pack);
  const policyDecision = pack.policyDecision || {};
  const scoreLabels = fitScoreLabels();
  const deploymentMeta = [
    pack.sizeTier || '体量待补充',
    deployMeta.label,
    pack.policyMode === 'externally-managed' ? `推荐托管 ${hostingTarget}` : '',
  ].filter(Boolean).map((item) => `<span class="badge">${item}</span>`).join('');
  const runtimeMeta = [pack.variant, pack.sizeHint, pack.runtimeHint].filter(Boolean).join(' · ');
  const card = document.createElement('div');
  card.className = `list-card local-model-card local-model-card-${pack.policyMode || 'display-only'}`;
  card.innerHTML = `
    <div class="list-row split-row">
      <div class="stack-tight selector-title-group">
        <div class="selector-rank">${providerLabel(pack.provider)}</div>
        <div class="list-row">
          <h3>${pack.name}</h3>
          <span class="badge ${pack.enabled ? 'completed' : 'paused'}">${pack.enabled ? '已启用' : '未启用'}</span>
          <span class="badge ${pack.downloaded ? 'completed' : 'paused'}">${pack.downloaded ? '已下载' : '未下载'}</span>
          <span class="badge ${localPolicyModeClass(pack.policyMode)}">${localPolicyModeLabels[pack.policyMode] || pack.policyMode || '仅展示'}</span>
          <span class="badge ${localInstallStateClass(pack.installState)}">${localInstallStateLabels[pack.installState] || pack.installState}</span>
        </div>
        <p class="selector-fullname">${providerLabel(pack.provider)} / ${pack.modelName} / ${pack.version}</p>
      </div>
      <div class="score-strip compact-score-strip">
        <div><span>${scoreLabels.filter}</span><strong>${pack.filterScore}</strong></div>
        <div><span>${scoreLabels.review}</span><strong>${pack.reviewScore}</strong></div>
        <div><span>${scoreLabels.alignment}</span><strong>${pack.alignmentScore}</strong></div>
      </div>
    </div>
    <p>${pack.description}</p>
    <div class="lifecycle-strip">
      <div>
        <span>当前状态</span>
        <strong>${localInstallStateLabels[pack.installState] || pack.installState}</strong>
      </div>
      <div>
        <span>状态说明</span>
        <strong>${pack.statusDetail}</strong>
      </div>
    </div>
    <div class="stage-list">${deploymentMeta}</div>
    ${runtimeMeta ? `<p class="run-meta">运行信息 ${runtimeMeta}</p>` : ''}
    <div class="stage-list">
      ${capabilityTags.map((tag) => `<span class="badge">${localCapabilityLabels[tag] || tag}</span>`).join('') || '<span class="badge">通用候选</span>'}
    </div>
    <p class="run-meta">推荐配置 ${pack.systemRequirement}</p>
    <p class="run-meta">使用提示 ${pack.policyHint}</p>
    <div class="policy-strip">
      <div>
        <span>当前方式</span>
        <strong>${localPolicyModeLabels[pack.policyMode] || pack.policyMode || '仅展示'}</strong>
      </div>
      <div>
        <span>限制原因</span>
        <strong>${policyDecisionReasonLabel(policyDecision.reason || 'runtime-tier')}</strong>
      </div>
    </div>
    ${policyDecision.message ? `<div class="policy-note">${policyDecision.message}</div>` : ''}
    ${pack.warning ? `<div class="warning-block">${pack.warning}</div>` : ''}
    <div class="actions">
      <button type="button" data-action="toggle" ${canToggleLocalModel(pack) ? '' : 'disabled'}>${localInstallActionLabel(pack)}</button>
      ${pack.policyMode === 'externally-managed' ? '<button type="button" class="ghost-button" data-action="advice">查看托管建议</button>' : ''}
      <button type="button" class="ghost-button" data-action="remove" ${!pack.downloaded || localInstallBusy(pack) ? 'disabled' : ''}>移除</button>
    </div>
  `;
  card.querySelector('[data-action="toggle"]').onclick = async () => {
    try {
      await fetchJSON(`/api/settings/model-packs/${pack.id}`, {
        method: 'POST',
        body: JSON.stringify({ enabled: !pack.enabled || !pack.downloaded, downloaded: true, remove: false }),
      });
      await loadSystemSummary();
      pushEvent(`已更新本地模型：${pack.name}`);
    } catch (error) {
      pushEvent(`本地模型更新失败，请稍后重试：${error.message}`);
    }
  };
  card.querySelector('[data-action="advice"]')?.addEventListener('click', async () => {
    try {
      await buildExternalHostingAdvice(pack);
      pushEvent(`已更新托管建议：${pack.name}`);
    } catch (error) {
      pushEvent(`托管建议暂时无法生成：${error.message}`);
    }
  });
  card.querySelector('[data-action="remove"]').onclick = async () => {
    try {
      await fetchJSON(`/api/settings/model-packs/${pack.id}`, {
        method: 'POST',
        body: JSON.stringify({ enabled: false, downloaded: false, remove: true }),
      });
      await loadSystemSummary();
      pushEvent(`已移除本地模型：${pack.name}`);
    } catch (error) {
      pushEvent(`本地模型移除失败，请稍后重试：${error.message}`);
    }
  };
  return card;
}

function renderAudit() {
  const root = q('#audit-report');
  if (!root) return;
  if (!state.audit) {
    root.className = 'detail-empty';
    root.textContent = '系统体检结果待补充。';
    return;
  }
  root.className = '';
  root.innerHTML = `<div class="list-card"><div class="list-row"><strong>评分 ${state.audit.score}</strong><span class="badge">体检</span></div><p>${state.audit.summary}</p></div><div class="stack-list">${(state.audit.checks || []).map((item) => `<div class="list-card"><div class="list-row"><strong>${item.name}</strong><span class="badge ${item.status === 'passed' ? 'completed' : item.status === 'failed' ? 'failed' : 'paused'}">${auditStatusLabel(item.status)}</span></div><p>${item.summary}</p><p class="run-meta">建议：${item.suggestion}</p></div>`).join('')}</div>`;
}

function renderAdvice() {
  const root = q('#advice-report');
  if (!root) return;
  if (!state.advice) {
    root.className = 'detail-empty';
    root.textContent = '升级建议待补充。';
    return;
  }
  root.className = '';
  const adviceTitle = adviceModeLabel(state.advice.mode);
  const hostingBadge = state.advice.hostingMode ? `<span class="badge queued">${hostingModeLabel(state.advice.hostingMode)}</span>` : '';
  const checklist = (state.advice.checklist || []).map((item) => `<div class="list-card"><p>${item}</p></div>`).join('');
  const consoleSteps = (state.advice.consoleSteps || []).map((item) => `<div class="list-card"><p>${item}</p></div>`).join('');
  const suggestions = (state.advice.suggestions || []).map((item) => `<div class="list-card"><p>${item}</p></div>`).join('');
  root.className = 'advice-report-shell';
  root.innerHTML = `
    <div class="stack-list advice-layout">
      <div class="list-card advice-hero-card">
        <div class="list-row"><strong>${adviceTitle}</strong><span class="badge">建议</span>${hostingBadge}</div>
        <p>${state.advice.summary}</p>
        ${state.advice.promptBundle ? '<p class="run-meta">建议已结合当前机器情况、已接入能力与目标场景整理。</p>' : ''}
      </div>
      ${checklist ? `<section class="advice-section"><div class="section-kicker">实施清单</div><div class="stack-list">${checklist}</div></section>` : ''}
      ${consoleSteps ? `<section class="advice-section"><div class="section-kicker">控制台接入</div><div class="stack-list">${consoleSteps}</div></section>` : ''}
      <section class="advice-section"><div class="section-kicker">方案说明</div><div class="stack-list">${suggestions}</div></section>
    </div>
  `;
}

function renderSystemSummary() {
  const target = q('#system-summary');
  if (!state.system) {
    target.textContent = '正在加载系统概况。';
    return;
  }
  const totals = Object.entries(state.system.runTotals || {}).map(([status, count]) => `<span class="badge ${status}">${statusLabel(status)} ${count}</span>`).join('');
  const modes = (state.system.interfaceModes || []).map((mode) => `<span class="badge">${interfaceModeLabel(mode)}</span>`).join('');
  const tools = (state.system.toolCatalog || []).slice(0, 6).map((tool) => `<span class="badge">${toolLabels[tool.name] || tool.name}</span>`).join('');
  const tests = (state.system.testCatalog || []).slice(0, 6).map((item) => `<span class="badge">${item.name}</span>`).join('');
  const profile = state.system.systemProfile || {};
  const profileFeatures = (profile.recommendedFeatures || []).map((item) => `<span class="badge">${item}</span>`).join('');
  const upgrades = (profile.upgradeSuggestions || []).map((item) => `<div class="list-card"><p>${item}</p></div>`).join('');
  const providerCount = (state.system.providers || []).length;
  const readyProviderCount = (state.system.providers || []).filter((provider) => provider.configured).length;
  const capabilityProfiles = state.system.capabilityProfiles || [];
  const optimizationProfiles = state.system.optimizationProfiles || [];
  const runtimePolicy = state.system.runtimePolicy || {};
  const localStats = getLocalModelPolicyStats(state.system.builtinModelPacks || []);
  const primaryDirection = primaryGovernanceDirection(localStats);
  const capabilityCards = capabilityProfiles.map((item) => `
    <div class="list-card capability-runtime-card">
      <div class="list-row split-row">
        <strong>${item.name}</strong>
        <span class="badge ${capabilityModeClass(item.mode)}">${capabilityModeLabels[item.mode] || item.mode}</span>
      </div>
      <p>${item.summary}</p>
      <p class="run-meta">${item.reason}</p>
      <div class="stage-list">
        <span class="badge">${toolCategoryLabel(item.category)}</span>
        <span class="badge ${item.allowed ? 'completed' : 'failed'}">${item.allowed ? '当前可用' : '当前受限'}</span>
        <span class="badge ${item.recommended ? 'completed' : 'paused'}">${item.recommended ? '当前建议' : '备选方式'}</span>
      </div>
    </div>
  `).join('');
  const optimizationCards = optimizationProfiles.map((item) => `
    <div class="list-card capability-runtime-card">
      <div class="list-row split-row">
        <strong>${item.title}</strong>
        <span class="badge ${optimizationPriorityClass(item.priority)}">${item.priority === 'high' ? '高优先级' : item.priority === 'medium' ? '中优先级' : '低优先级'}</span>
      </div>
      <p>${item.action}</p>
      <p class="run-meta">${item.reason}</p>
    </div>
  `).join('');
  target.innerHTML = `
    <div class="stack-list">
      <div class="list-card">
         <div class="summary-strip">
           <div><strong>${Object.values(state.system.runTotals || {}).reduce((sum, count) => sum + count, 0)}</strong><span>任务总量</span></div>
           <div><strong>${readyProviderCount}</strong><span>已接入平台</span></div>
           <div><strong>${(state.system.toolCatalog || []).length}</strong><span>工具能力</span></div>
         </div>
          <div class="list-row">${totals || '<span class="badge">暂无任务</span>'}</div>
          <p class="run-meta">运行方式 ${schedulerModeLabel(state.system.schedulerMode)} · 部署方式 ${deployModeLabel(state.system.deployMode)}</p>
          <p class="run-meta">当前主机 ${profile.cpuCores || '-'}C / ${profile.memoryMB || '-'}MB · 机器级别 ${profile.tier || '-'}</p>
          <p class="run-meta">建议并发 ${profile.recommendedConcurrency || '-'} · ${profile.strategySummary || '系统会自动根据机器规格调整运行方式。'}</p>
          <div class="stage-list">${modes}</div>
         <p class="run-meta">当前建议能力</p>
          <div class="stage-list">${profileFeatures}</div>
         <p class="run-meta">能力总览</p>
          <div class="stage-list">${tools}</div>
         <p class="run-meta">验证清单</p>
         <div class="stage-list">${tests}</div>
       </div>
        <div class="list-card">
          <div class="list-row"><strong>免 SSH 就绪度</strong><span class="badge">${profile.noSshReadiness || '未知'}</span></div>
          <p>推荐 VPS：${profile.recommendedVps || '-'}</p>
          <p class="run-meta">已接入平台 ${readyProviderCount} / ${providerCount}，详细选择可前往模型总览。</p>
        </div>
        <div class="list-card">
          <div class="list-row split-row"><strong>全系统运行方式</strong><span class="badge">全局建议</span></div>
          <p>核心能力尽量保持稳定可用，较重能力按需启用，更高资源消耗的部分优先放到独立机器或仅保留目录信息。</p>
          <div class="stack-list capability-runtime-list">${capabilityCards}</div>
        </div>
        <div class="list-card">
          <div class="list-row split-row"><strong>自动优化建议</strong><span class="badge">按机器调整</span></div>
          <p>系统会根据当前机器规格给出最适合的运行方式，优先保证稳定、轻量和长期可维护。</p>
          <div class="stack-list capability-runtime-list">${optimizationCards}</div>
        </div>
        <div class="list-card">
          <div class="list-row split-row"><strong>默认使用建议</strong><span class="badge">${runtimeProfileLabel(runtimePolicy.profile)}</span></div>
          <p>${runtimePolicy.summary || '系统尚未生成默认使用建议。'}</p>
          <div class="policy-snapshot system-governance-strip">
            <div>
              <span>本地模型当前重点</span>
              <strong>${primaryDirection}</strong>
            </div>
            <div>
              <span>本地模型方式</span>
              <strong>${runtimePolicy.allowLocalModels ? '按需启用优先' : '仅展示优先'}</strong>
            </div>
          </div>
          <div class="stage-list">
            <span class="badge ${runtimePolicy.allowBackgroundJobs ? 'completed' : 'paused'}">后台任务 ${runtimePolicy.allowBackgroundJobs ? '可按需启用' : '暂不开放'}</span>
            <span class="badge ${runtimePolicy.allowLocalModels ? 'completed' : 'failed'}">本地模型 ${runtimePolicy.allowLocalModels ? '可按需启用' : '仅展示'}</span>
            <span class="badge">活跃运行单 ${runtimePolicy.maxConcurrentRuns || 1}</span>
            <span class="badge">重动作并发 ${runtimePolicy.maxHeavyActions || 0}</span>
            <span class="badge">缓存预算 ${runtimePolicy.cacheBudgetMB || 0}MB</span>
            <span class="badge">验证深度 ${validationDepthLabels[runtimePolicy.validationDepth] || runtimePolicy.validationDepth || '标准闭环'}</span>
          </div>
        </div>
       ${upgrades}
      </div>
    `;
  renderHeroMetrics();
  renderCapabilityAtlas();
  renderFeatureToggles();
  renderAtomicTools();
  renderModelPacks();
  renderRuntimePolicyHint();
}

async function analyzeCurrentInput() {
  const payload = {
    title: q('input[name="title"]').value,
    goal: q('textarea[name="goal"]').value,
  };
  state.analysisPreview = await fetchJSON('/api/intake/analyze', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  state.selectedTests = [...(state.analysisPreview.recommendedTests || [])];
  renderTemplates();
  autoSelectPreferredTemplate(true);
  renderAnalysisPreview();
  renderTestOptions();
  renderAtomicToolPreview();
}

function pushEvent(message) {
  const root = q('#event-log');
  const item = document.createElement('div');
  item.className = 'event-item';
  item.innerHTML = `<p>${message}</p><time>${new Date().toLocaleTimeString()}</time>`;
  root.prepend(item);
}

async function loadTemplates() {
  const payload = await fetchJSON('/api/templates');
  state.templates = payload.items || [];
  renderTemplates();
  renderRecentTemplates();
  const select = q('#template-select');
  if (select && select.options.length > 0) {
    if (!select.value) {
      select.selectedIndex = 0;
    }
    autoSelectPreferredTemplate(false);
  }
}

async function loadModels() {
  const payload = await fetchJSON('/api/models');
  state.models = payload.items || [];
  renderModelProviderFilter();
  renderModels();
  renderHeroMetrics();
}

async function loadRuns() {
  const payload = await fetchJSON('/api/runs');
  state.runs = payload.items || [];
  renderRuns();
  await loadSystemSummary();
}

async function loadSystemSummary() {
  state.system = await fetchJSON('/api/system');
  applyRuntimePolicyDefaults();
  renderSystemSummary();
  renderTestOptions();
  renderAtomicToolPreview();
}

function applyRuntimePolicyDefaults() {
  if (state.toolPolicyApplied || !state.system?.runtimePolicy) {
    syncRuntimePolicyFormDefaults();
    syncTestDefaults();
    return;
  }
  const enabledSet = new Set(state.system.runtimePolicy.defaultEnabledTools || []);
  const disabledSet = new Set(state.system.runtimePolicy.defaultDisabledTools || []);
  state.tools = state.tools.map((tool) => {
    if (enabledSet.has(tool.name)) {
      return { ...tool, enabled: true };
    }
    if (disabledSet.has(tool.name)) {
      return { ...tool, enabled: false };
    }
    return tool;
  });
  state.toolPolicyApplied = true;
  syncRuntimePolicyFormDefaults(true);
  syncTestDefaults(true);
  renderTools();
}

async function loadWorkflowOptions() {
  const payload = await fetchJSON('/api/workflow/options');
  state.workflowOptions = payload.items || [];
  renderWorkflowOptions();
}

async function runAudit() {
  state.audit = await fetchJSON('/api/system/audit');
  renderAudit();
}

async function buildAdvice() {
  state.advice = await fetchJSON('/api/system/advice', { method: 'POST', body: JSON.stringify({ mode: 'ops-review', goal: '提高免 SSH 运维和自动化覆盖度' }) });
  renderAdvice();
}

async function buildExternalHostingAdvice(pack) {
  const hostingTarget = localDeploymentMeta(pack).key;
  state.advice = await fetchJSON('/api/system/advice', {
    method: 'POST',
    body: JSON.stringify({
      mode: 'external-hosting',
      goal: `${pack.name} (${pack.sizeTier}) 当前为 ${localPolicyModeLabels[pack.policyMode] || pack.policyMode}，请生成${hostingModeLabel(hostingTarget)}建议`,
      target: hostingTarget,
    }),
  });
  renderAdvice();
}

async function mutateRun(id, action, payload, options = {}) {
  const { suppressError = false, rethrow = false } = options;
  try {
    await fetchJSON(`/api/runs/${id}/${action}`, {
      method: 'POST',
      body: payload ? JSON.stringify(payload) : undefined,
    });
    await loadRuns();
  } catch (error) {
    if (!suppressError) {
      pushEvent(`任务操作失败，请稍后重试：${error.message}`);
    }
    if (rethrow) {
      throw error;
    }
  }
}

function selectRun(id) {
  state.selectedRunId = id;
  renderRunDetail();
  renderDevRunBinding();
}

function initStages() {
  const select = q('#stages-select');
  stageOptions.forEach((stage) => {
    const option = document.createElement('option');
    option.value = stage;
    option.textContent = stage;
    if (['intent', 'context', 'plan', 'implement', 'result', 'test'].includes(stage)) {
      option.selected = true;
    }
    select.appendChild(option);
  });
  select.addEventListener('change', () => {
    renderRuntimePolicyHint();
    renderAtomicToolPreview();
  });
}

function initMenu() {
  document.querySelectorAll('.menu-item').forEach((button) => {
    button.addEventListener('click', () => setView(button.dataset.view));
  });
}

function initRunForm() {
  q('#template-select')?.addEventListener('change', (event) => {
    state.templatePinned = true;
    state.templateSource = 'manual-select';
    rememberRecentTemplate(event.target.value);
    applyTemplateSelection(event.target.value);
    pushEvent(`已切换模板：${templateById(event.target.value)?.name || event.target.value}`);
  });
  q('#test-mode-select')?.addEventListener('change', (event) => {
    state.testMode = event.target.value;
    pushEvent(`测试模式已切换为：${event.target.selectedOptions[0]?.textContent || event.target.value}`);
  });
  q('#select-all-tests')?.addEventListener('click', () => {
    state.selectedTests = (state.system?.testCatalog || []).map((item) => item.name);
    renderTestOptions();
    pushEvent('已全选测试项');
  });
  q('#select-recommended-tests')?.addEventListener('click', () => {
    state.selectedTests = [...(state.analysisPreview?.recommendedTests || [])];
    renderTestOptions();
    pushEvent('已切换为推荐测试项');
  });
  q('input[name="autoRepairEnabled"]')?.addEventListener('change', () => {
    syncRuntimePolicyFormDefaults();
  });
  q('input[name="remoteDeployEnabled"]')?.addEventListener('change', (event) => {
    const select = q('#stages-select');
    const deployOption = Array.from(select?.options || []).find((option) => option.value === 'deploy');
    if (!event.target.checked && deployOption) {
      deployOption.selected = false;
      renderRuntimePolicyHint();
    }
  });
  q('#run-form').addEventListener('submit', async (event) => {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const payload = {
      title: form.get('title'),
      goal: form.get('goal'),
      templateId: form.get('templateId'),
      templateSource: state.templateSource,
      testMode: form.get('testMode'),
      selectedTests: state.selectedTests,
      autoRepairEnabled: form.get('autoRepairEnabled') === 'on',
      autoRepairMode: form.get('autoRepairMode'),
      remoteDeployEnabled: form.get('remoteDeployEnabled') === 'on',
      stages: Array.from(q('#stages-select').selectedOptions).map((option) => option.value),
      tools: state.tools,
    };
    try {
      const run = await fetchJSON('/api/runs', {
        method: 'POST',
        body: JSON.stringify(payload),
      });
      state.selectedRunId = run.id;
      pushEvent(`任务已创建：${run.title}`);
      setView('runs');
      await loadRuns();
      event.currentTarget.reset();
      state.analysisPreview = null;
      state.templateSource = '';
      state.testMode = 'template';
      state.selectedTests = [];
      state.templatePinned = false;
      rememberRecentTemplate(run.templateId);
      syncRuntimePolicyFormDefaults(true);
      renderTemplates();
      renderRecentTemplates();
      autoSelectPreferredTemplate(false);
      renderAnalysisPreview();
      renderTestOptions();
    } catch (error) {
      pushEvent(`任务创建失败，请稍后重试：${error.message}`);
    }
  });

  q('#refresh-runs').addEventListener('click', loadRuns);
  q('#analyze-run').addEventListener('click', async () => {
    try {
      await analyzeCurrentInput();
      pushEvent('智能预演已刷新');
    } catch (error) {
      pushEvent(`智能预演暂时不可用：${error.message}`);
    }
  });
}

function initDevWorkbench() {
  q('#refresh-dev-files')?.addEventListener('click', async () => {
    try {
      await loadDevFiles(state.dev.currentDir || '.');
      pushEvent('文件浏览器已刷新');
    } catch (error) {
      pushEvent(`刷新文件浏览器失败：${error.message}`);
    }
  });
  q('#save-dev-file')?.addEventListener('click', async () => {
    try {
      await saveDevFile();
    } catch (error) {
      pushEvent(`保存文件失败：${error.message}`);
    }
  });
  q('#rollback-dev-file')?.addEventListener('click', async () => {
    try {
      await rollbackDevFile();
    } catch (error) {
      pushEvent(`回滚文件失败：${error.message}`);
    }
  });
  q('#run-dev-command')?.addEventListener('click', async () => {
    try {
      await runDevCommand(false);
    } catch (error) {
      pushEvent(`执行命令失败：${error.message}`);
    }
  });
  q('#run-dev-command-bg')?.addEventListener('click', async () => {
    try {
      await runDevCommand(true);
    } catch (error) {
      pushEvent(`启动后台命令失败：${error.message}`);
    }
  });
  q('#open-dev-preview')?.addEventListener('click', async () => {
    try {
      await openDeploymentPreview(q('#dev-preview-port')?.value?.trim() || '8080');
    } catch (error) {
      pushEvent(`打开预览失败：${error.message}`);
    }
  });
}

function initControlActions() {
  q('#run-audit')?.addEventListener('click', async () => {
    await runAudit();
    pushEvent('系统体检已更新');
  });
  q('#build-advice')?.addEventListener('click', async () => {
    await buildAdvice();
    pushEvent('升级建议已更新');
  });
}

function initModes() {
  document.body.dataset.theme = state.theme;
  const toggle = q('#theme-toggle');
  const syncThemeToggle = () => {
    if (!toggle) return;
    const isDay = state.theme === 'day';
    toggle.classList.toggle('is-day', isDay);
    toggle.setAttribute('aria-pressed', String(isDay));
  };
  syncThemeToggle();
  toggle?.addEventListener('click', () => {
    state.theme = state.theme === 'day' ? 'night' : 'day';
    document.body.dataset.theme = state.theme;
    syncThemeToggle();
  });
}

function initModelFilters() {
  q('#model-search')?.addEventListener('input', (event) => {
    state.modelFilters.query = event.target.value.trim().toLowerCase();
    renderModels();
  });
  q('#model-provider-filter')?.addEventListener('change', (event) => {
    state.modelFilters.provider = event.target.value;
    renderModels();
  });
  q('#model-region-filter')?.addEventListener('change', (event) => {
    state.modelFilters.region = event.target.value;
    renderModels();
  });
  q('#model-readiness-filter')?.addEventListener('change', (event) => {
    state.modelFilters.readiness = event.target.value;
    renderModels();
  });
  q('#local-model-search')?.addEventListener('input', (event) => {
    state.localModelFilters.query = event.target.value.trim().toLowerCase();
    renderModelPacks();
  });
  q('#local-tier-filter')?.addEventListener('change', (event) => {
    state.localModelFilters.tier = event.target.value;
    renderModelPacks();
  });
  q('#local-install-filter')?.addEventListener('change', (event) => {
    state.localModelFilters.install = event.target.value;
    renderModelPacks();
  });
  q('#local-compat-filter')?.addEventListener('change', (event) => {
    state.localModelFilters.compat = event.target.value;
    renderModelPacks();
  });
  q('#local-deploy-filter')?.addEventListener('change', (event) => {
    state.localModelFilters.deploy = event.target.value;
    renderModelPacks();
  });
  q('#local-focus-filter')?.addEventListener('change', (event) => {
    state.localModelFilters.focus = event.target.value;
    renderModelPacks();
  });
  q('#local-capability-filter')?.addEventListener('change', (event) => {
    state.localModelFilters.capability = event.target.value;
    renderModelPacks();
  });
  q('#local-sort-filter')?.addEventListener('change', (event) => {
    state.localModelFilters.sort = event.target.value;
    renderModelPacks();
  });
  q('#reset-local-filters')?.addEventListener('click', () => {
    resetLocalModelFilters();
    renderModelPacks();
    pushEvent('本地模型筛选已重置');
  });
}

function initModelModal() {
  q('#close-model-modal')?.addEventListener('click', closeModelModal);
  q('#model-modal')?.addEventListener('click', (event) => {
    if (event.target.id === 'model-modal') {
      closeModelModal();
    }
  });
}

function initEvents() {
  const source = new EventSource('/api/events');
  source.onmessage = async (event) => {
    const payload = JSON.parse(event.data);
    pushEvent(payload.message);
    await loadRuns();
  };
}

async function boot() {
  initMenu();
  initModes();
  initModelFilters();
  initModelModal();
  initStages();
  initRunForm();
  initDevWorkbench();
  initControlActions();
  renderTools();
  renderAnalysisPreview();
  renderAudit();
  renderAdvice();
  renderDevEditor();
  renderDevTerminal();
  renderDevPreview();
  await Promise.all([loadTemplates(), loadModels(), loadRuns(), loadWorkflowOptions(), loadDevFiles('.'), loadDevSessions()]);
  renderDevRunBinding();
  window.setInterval(() => {
    loadDevSessions().catch(() => {});
  }, 3000);
  initEvents();
}

boot().catch((error) => pushEvent(`页面初始化失败，请刷新后重试：${error.message}`));
