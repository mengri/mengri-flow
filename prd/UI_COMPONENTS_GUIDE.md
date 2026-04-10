# Mengri Flow UI 组件库指南

## 🎯 概述

Mengri Flow UI 是一个现代化、美观、易用的设计系统和组件库，基于 Vue 3 + Element Plus + TailwindCSS 构建。本文档提供完整的组件使用指南和最佳实践。

## 🏗️ 设计系统架构

### 核心技术栈
- **Vue 3**: 主流前端框架
- **Element Plus**: UI组件库基础
- **TailwindCSS**: 原子化CSS框架
- **TypeScript**: 类型安全支持
- **Chart.js**: 数据可视化

### 设计原则
1. **一致性**: 所有组件遵循统一设计规范
2. **可访问性**: 符合WCAG 2.1 AA标准
3. **响应式**: 完美适配所有设备尺寸
4. **可复用性**: 组件高度可配置和可组合
5. **性能**: 优化加载和渲染性能

## 🎨 核心组件介绍

### MengriTopbar - 顶部导航栏
现代化的顶部导航栏，集成了用户菜单、通知、搜索和主题切换功能。

```vue
<script setup>
import MengriTopbar from '@/components/ui/MengriTopbar.vue'

const menuItems = [
  { path: '/dashboard', label: 'Dashboard', icon: 'HomeIcon' },
  { path: '/workflows', label: 'Workflows', icon: 'ArrowsRightLeftIcon' },
  { path: '/analytics', label: 'Analytics', icon: 'ChartBarIcon' },
  { path: '/settings', label: 'Settings', icon: 'CogIcon' },
]
</script>

<template>
  <MengriTopbar
    :menu-items="menuItems"
    :show-search="true"
    @toggle-sidebar="toggleSidebar"
    @logout="handleLogout"
  />
</template>
```

**特性：**
- ✅ 响应式设计（移动端汉堡菜单）
- ✅ 用户下拉菜单
- ✅ 通知中心和徽章
- ✅ 实时搜索建议
- ✅ 主题切换支持
- ✅ 键盘导航支持

### MengriSidebar - 侧边栏导航
可折叠的侧边栏导航，支持多级菜单和工作区切换。

```vue
<script setup>
import MengriSidebar from '@/components/ui/MengriSidebar.vue'

const navigation = [
  {
    title: 'Workspace',
    items: [
      { path: '/dashboard', label: 'Dashboard', icon: 'HomeIcon' },
      { path: '/workflows', label: 'Workflows', icon: 'ArrowsRightLeftIcon' },
      { 
        path: '/automation', 
        label: 'Automation', 
        icon: 'CogIcon',
        children: [
          { path: '/automation/rules', label: 'Rules' },
          { path: '/automation/templates', label: 'Templates' },
        ]
      },
    ]
  },
  {
    title: 'Management',
    items: [
      { path: '/users', label: 'Users', icon: 'UsersIcon', badge: 3 },
      { path: '/integrations', label: 'Integrations', icon: 'PuzzleIcon' },
    ]
  }
]
</script>

<template>
  <MengriSidebar
    :navigation="navigation"
    @open-settings="openSettings"
    @workspace-change="switchWorkspace"
  />
</template>
```

**特性：**
- ✅ 可折叠设计
- ✅ 多级菜单支持
- ✅ 工作区切换
- ✅ 徽章通知
- ✅ 平滑动画过渡
- ✅ 移动端优化

### MButton - 增强按钮组件
基于Element Plus Button的增强版本，提供更多变体和功能。

```vue
<script setup>
import MButton from '@/components/ui/MButton.vue'
import { PlusIcon, DownloadIcon } from '@heroicons/vue/24/outline'
</script>

<template>
  <!-- 主要按钮 -->
  <MButton variant="primary" @click="handleClick">
    Create New
  </MButton>
  
  <!-- 带图标按钮 -->
  <MButton variant="secondary" :icon="DownloadIcon">
    Export
  </MButton>
  
  <!-- 加载状态 -->
  <MButton variant="primary" :loading="isLoading">
    Processing...
  </MButton>
  
  <!-- 带徽章 -->
  <MButton variant="info" badge="3">
    Notifications
  </MButton>
  
  <!-- 链接按钮 -->
  <MButton variant="link" href="/help" target="_blank">
    Help Center
  </MButton>
</template>
```

**变体类型：**
- `primary`: 主要操作按钮
- `secondary`: 次要操作按钮
- `tertiary`: 第三级按钮（带边框）
- `danger`: 危险操作（删除、取消）
- `success`: 成功操作
- `warning`: 警告操作
- `info`: 信息操作
- `text`: 文本按钮
- `link`: 链接样式按钮

**尺寸：** `xs` | `sm` | `md` | `lg` | `xl`
**形状：** `default` | `pill` | `square` | `circle`

### MStatCard - 统计卡片
现代化的统计卡片，支持图表、进度条和趋势显示。

```vue
<script setup>
import MStatCard from '@/components/ui/MStatCard.vue'
import { UserGroupIcon, ArrowTrendingUpIcon } from '@heroicons/vue/24/outline'

const trendData = [30, 45, 60, 55, 70, 80, 75]
</script>

<template>
  <!-- 基础统计卡片 -->
  <MStatCard
    title="Active Users"
    :value="1245"
    subtitle="Last 30 days"
    color="primary"
    :change="15"
    change-type="increase"
  />
  
  <!-- 带图表的统计卡片 -->
  <MStatCard
    title="Performance"
    :value="88"
    unit="%"
    icon="TrendingUpIcon"
    color="success"
    :progress="88"
    progress-label="Target: 90%"
    :trend-data="trendData"
    show-trend
  />
  
  <!-- 紧凑卡片 -->
  <MStatCard
    title="Response Time"
    :value="120"
    unit="ms"
    variant="compact"
    :change="-5"
    change-type="decrease"
    show-refresh
    @refresh="refreshData"
  />
</template>
```

**特性：**
- ✅ 多种颜色主题
- ✅ 进度条可视化
- ✅ 趋势图表集成
- ✅ 变化指示器
- ✅ 加载状态
- ✅ 刷新功能

### DashboardView - 仪表板视图
完整的工作流仪表板，展示关键指标和实时数据。

```vue
<script setup>
import DashboardView from '@/views/DashboardView.vue'

// 仪表板自动加载用户数据和统计信息
</script>

<template>
  <DashboardView />
</template>
```

**功能模块：**
1. **快捷操作**: 主要功能快速入口
2. **统计卡片**: 关键指标概览
3. **最近活动**: 用户操作时间线
4. **工作流运行**: 近期的执行记录
5. **资源使用**: CPU、内存、数据库监控
6. **即将触发器**: 计划的自动化任务

## 📱 响应式设计

### 断点系统
```css
/* 断点定义 */
xs: 480px    /* 超小屏幕 */
sm: 640px    /* 小屏幕 */
md: 768px    /* 中屏幕 */
lg: 1024px   /* 大屏幕 */
xl: 1280px   /* 超大屏幕 */
2xl: 1536px  /* 特大屏幕 */
```

### 组件响应式行为

| 组件 | 移动端 (< 640px) | 平板端 (640-1024px) | 桌面端 (> 1024px) |
|------|-------------------|---------------------|-------------------|
| **Topbar** | 汉堡菜单 | 精简导航 | 完整导航 |
| **Sidebar** | 底部导航 | 可折叠侧边栏 | 完整侧边栏 |
| **Dashboard** | 单列布局 | 两列布局 | 三列网格 |
| **Tables** | 卡片列表 | 响应式表格 | 完整表格 |
| **Forms** | 单列表单 | 双列表单 | 三列表单 |

### 响应式工具类
```vue
<template>
  <!-- 移动端隐藏，桌面端显示 -->
  <div class="hidden md:block">
    Desktop only content
  </div>
  
  <!-- 桌面端隐藏，移动端显示 -->
  <div class="md:hidden">
    Mobile only content
  </div>
</template>
```

## ♿ 可访问性

### WCAG 2.1 AA 合规性

**色彩对比度:**
- 普通文本: 最小 4.5:1 比例
- 大文本 (18px及以上): 最小 3:1 比例
- UI组件和图标: 最小 3:1 比例

**键盘导航:**
```html
<!-- 使用语义化HTML -->
<nav role="navigation" aria-label="Main navigation">
  <!-- 导航项 -->
</nav>

<!-- 提供键盘快捷键 -->
<button aria-keyshortcuts="Alt+1">Home</button>

<!-- 管理焦点 -->
<div role="dialog" aria-labelledby="dialog-title">
  <h2 id="dialog-title">Dialog Title</h2>
  <!-- 内容 -->
</div>
```

**屏幕阅读器支持:**
```vue
<template>
  <!-- 使用ARIA标签 -->
  <button 
    aria-label="Open user menu"
    aria-haspopup="true"
    :aria-expanded="menuOpen"
  >
    <UserIcon />
  </button>
  
  <!-- 为图片提供Alt文本 -->
  <img 
    src="/logo.png" 
    alt="Mengri Flow Logo - Low-code workflow automation platform"
    role="img"
  />
  
  <!-- 表单可访问性 -->
  <label for="email" class="sr-only">Email Address</label>
  <input 
    id="email"
    type="email"
    aria-describedby="email-help"
    aria-required="true"
  />
  <div id="email-help" class="text-sm text-gray-600">
    We'll never share your email with anyone else.
  </div>
</template>
```

**焦点管理:**
- ✅ 所有交互元素可通过键盘访问
- ✅ 焦点指示清晰可见
- ✅ 焦点顺序逻辑合理
- ✅ 模态框正确处理焦点捕获

**屏幕阅读器专用内容:**
```vue
<template>
  <!-- 隐藏只对屏幕阅读器可见 -->
  <span class="sr-only">This text is only for screen readers</span>
  
  <!-- ARIA实时区域 -->
  <div 
    aria-live="polite" 
    aria-atomic="true"
    class="sr-only"
  >
    {{ screenReaderAnnouncement }}
  </div>
</template>
```

### 包容性设计

**触摸目标尺寸:**
```css
/* 最小触摸目标 44px x 44px */
.btn {
  min-height: 44px;
  min-width: 44px;
  padding: 12px 16px;
}
```

**动画敏感度:**
```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

**文本缩放:**
```css
/* 支持最大 200% 文本缩放 */
:root {
  font-size: clamp(16px, 1vw, 20px);
}

.container {
  /* 使用相对单位 */
  padding: 1rem;
  margin: 1rem;
  
  /* 避免固定像素值 */
  max-width: 90rem; /* 不是 1440px */
}
```

## 🌈 动效与交互

### 过渡系统
```css
:root {
  --transition-fast: 150ms ease;
  --transition-normal: 300ms ease;
  --transition-slow: 500ms ease;
  --ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-out: cubic-bezier(0, 0, 0.2, 1);
  --ease-in: cubic-bezier(0.4, 0, 1, 1);
}
```

### 组件动效示例
```vue
<template>
  <!-- 悬停效果 -->
  <div class="card hover:scale-[1.02] transition-transform duration-300">
    Card content
  </div>
  
  <!-- 点击反馈 -->
  <button 
    class="active:scale-95 transition-transform duration-100"
    @click="handleClick"
  >
    Click Me
  </button>
  
  <!-- 列表项动画 -->
  <transition-group 
    name="list" 
    tag="ul"
    class="space-y-2"
  >
    <li 
      v-for="item in items" 
      :key="item.id"
      class="list-item"
    >
      {{ item.name }}
    </li>
  </transition-group>
</template>

<style scoped>
.list-enter-active,
.list-leave-active {
  transition: all 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

.list-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.list-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

.list-move {
  transition: transform 300ms cubic-bezier(0.4, 0, 0.2, 1);
}
</style>
```

### 加载状态
```vue
<template>
  <!-- 骨架屏 -->
  <div class="skeleton h-6 w-3/4 rounded"></div>
  <div class="skeleton h-4 w-1/2 rounded mt-2"></div>
  <div class="skeleton h-4 w-2/3 rounded mt-2"></div>
  
  <!-- 按钮加载状态 -->
  <MButton :loading="isLoading">
    {{ isLoading ? 'Processing...' : 'Submit' }}
  </MButton>
</template>
```

## 🎯 最佳实践

### 1. 组件使用指南

**避免样式冲突:**
```vue
<!-- 正确: 使用 scoped 样式 -->
<style scoped>
.button {
  /* 组件私有样式 */
}
</style>

<!-- 错误: 使用全局样式 -->
<style>
.button {
  /* 可能影响其他组件 */
}
</style>
```

**正确导入图标:**
```vue
<script setup>
// 正确: 按需导入
import { PlusIcon, TrashIcon } from '@heroicons/vue/24/outline'

// 避免: 全部导入
// import * as Icons from '@heroicons/vue/24/outline'
</script>
```

### 2. 性能优化

**懒加载组件:**
```vue
<script setup>
import { defineAsyncComponent } from 'vue'

// 懒加载大组件
const LargeChart = defineAsyncComponent(() =>
  import('@/components/LargeChart.vue')
)
</script>
```

**图片优化:**
```vue
<template>
  <!-- 使用合适的格式 -->
  <img 
    :src="avatar" 
    loading="lazy"
    decoding="async"
    alt="User avatar"
    :srcset="avatar + ' 1x, ' + avatar2x + ' 2x'"
  />
</template>
```

### 3. 状态管理

**使用Pinia Store:**
```ts
// stores/statistics.ts
export const useStatisticsStore = defineStore('statistics', {
  state: () => ({
    activeWorkflows: 0,
    successRate: 0,
    isLoading: false,
  }),
  
  actions: {
    async fetchStatistics() {
      this.isLoading = true
      try {
        const response = await fetch('/api/statistics')
        const data = await response.json()
        Object.assign(this, data)
      } catch (error) {
        console.error('Failed to fetch statistics:', error)
      } finally {
        this.isLoading = false
      }
    }
  }
})
```

### 4. 错误处理

**全局错误边界:**
```vue
<template>
  <ErrorBoundary>
    <DashboardView />
  </ErrorBoundary>
</template>

<script setup>
import { onErrorCaptured, ref } from 'vue'

const error = ref(null)

onErrorCaptured((err) => {
  error.value = err
  // 发送错误到监控服务
  logError(err)
  return false // 阻止错误继续向上传播
})
</script>
```

## 🔧 开发工作流

### 组件开发流程
1. **需求分析**: 明确组件功能和API设计
2. **原型设计**: Figma或类似工具制作设计稿
3. **组件实现**: 编写Vue组件
4. **测试验证**: 单元测试和可访问性测试
5. **文档编写**: 使用指南和示例代码
6. **代码审查**: PR审查和样式检查
7. **发布部署**: 版本管理和变更日志

### 调试工具

**TailwindCSS调试:**
```bash
# 查看生成的CSS
npx tailwindcss -o output.css

# 检查未使用的CSS
npx tailwindcss -c tailwind.config.js -i input.css -o output.css --watch --purge
```

**Chrome DevTools技巧:**
1. **元素状态**: 检查悬停、焦点状态
2. **响应式测试**: Device Mode 测试断点
3. **性能分析**: Lighthouse 可访问性评分
4. **颜色对比度**: 使用 Color Picker 工具

### 代码规范检查
```bash
# TypeScript检查
npm run type-check

# ESLint检查
npm run lint

# 样式检查
npm run stylelint

# 可访问性测试
npm run a11y
```

## 📚 资源与参考

### 设计资源
- **Figma设计稿**: [link-to-figma]
- **设计令牌**: `/docs/UI_DESIGN_SYSTEM.md`
- **图标库**: Heroicons, Lucide Icons

### 开发资源
- **组件文档**: `/docs/UI_COMPONENTS_GUIDE.md`
- **API文档**: `/docs/API_REFERENCE.md`
- **示例代码**: `/examples/`

### 测试工具
- **单元测试**: Vitest
- **E2E测试**: Playwright
- **可访问性**: axe-core
- **视觉回归**: Percy

---

**更新日志:**
- **v1.0.0**: 初始版本，包含核心组件库
- **v1.1.0**: 增加可访问性改进
- **v1.2.0**: 优化响应式设计
- **v1.3.0**: 集成数据可视化组件

**贡献指南:**
请参考 [`CONTRIBUTING.md`](CONTRIBUTING.md) 了解如何为UI组件库贡献代码。

**支持:**
- 文档问题: 创建Issue
- 功能请求: 使用Feature Request模板
- 错误报告: 包含复现步骤的Bug报告