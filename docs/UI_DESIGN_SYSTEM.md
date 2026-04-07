# Mengri Flow - UI设计系统

## 🎨 设计基础与品牌系统

### 色彩系统
**品牌色彩**：基于科技、信任、高效的蓝色系，结合现代化渐变设计
```css
/* 设计令牌 - Tailwind配置增强 */
:root {
  /* 主要品牌色系 */
  --color-primary-50: #eff6ff;
  --color-primary-100: #dbeafe;
  --color-primary-200: #bfdbfe;
  --color-primary-300: #93c5fd;
  --color-primary-400: #60a5fa;
  --color-primary-500: #3b82f6; /* 主品牌色 */
  --color-primary-600: #2563eb;
  --color-primary-700: #1d4ed8;
  --color-primary-800: #1e40af;
  --color-primary-900: #1e3a8a;
  --color-primary-950: #172554;

  /* 次要色系 - 青色 */
  --color-secondary-50: #f0fdfa;
  --color-secondary-100: #ccfbf1;
  --color-secondary-200: #99f6e4;
  --color-secondary-300: #5eead4;
  --color-secondary-400: #2dd4bf;
  --color-secondary-500: #14b8a6; /* 次要品牌色 */
  --color-secondary-600: #0d9488;
  --color-secondary-700: #0f766e;
  --color-secondary-800: #115e59;
  --color-secondary-900: #134e4a;
  --color-secondary-950: #042f2e;

  /* 语义色系 */
  --color-success-500: #10b981; /* 成功状态 */
  --color-warning-500: #f59e0b; /* 警告状态 */
  --color-error-500: #ef4444;   /* 错误状态 */
  --color-info-500: #3b82f6;    /* 信息状态 */

  /* 中性色系 */
  --color-gray-50: #f9fafb;
  --color-gray-100: #f3f4f6;
  --color-gray-200: #e5e7eb;
  --color-gray-300: #d1d5db;
  --color-gray-400: #9ca3af;
  --color-gray-500: #6b7280;
  --color-gray-600: #4b5563;
  --color-gray-700: #374151;
  --color-gray-800: #1f2937;
  --color-gray-900: #111827;
  --color-gray-950: #030712;

  /* 背景与表面色 */
  --color-background: #ffffff;
  --color-surface: #ffffff;
  --color-elevated: #f9fafb;
  --color-overlay: rgba(0, 0, 0, 0.5);
}

/* 深色主题 */
[data-theme="dark"] {
  --color-primary-500: #60a5fa;
  --color-secondary-500: #5eead4;
  --color-background: #111827;
  --color-surface: #1f2937;
  --color-elevated: #374151;
  --color-overlay: rgba(0, 0, 0, 0.7);
}
```

### 字体系统
**字体家族**：现代、清晰的无衬线字体系统
```css
:root {
  /* 字体栈 */
  --font-family-primary: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  --font-family-mono: 'JetBrains Mono', 'Cascadia Code', 'Fira Code', 'SF Mono', Monaco, 'Courier New', monospace;
  
  /* 字体比例 - 基于 16px 基础（1rem）的 1.125 比例因子 */
  --font-size-xs: 0.75rem;      /* 12px - 辅助文本、标签 */
  --font-size-sm: 0.875rem;     /* 14px - 正文、表单标签 */
  --font-size-base: 1rem;       /* 16px - 常规正文 */
  --font-size-lg: 1.125rem;     /* 18px - 小标题 */
  --font-size-xl: 1.25rem;      /* 20px - 子标题 */
  --font-size-2xl: 1.5rem;      /* 24px - 标题 */
  --font-size-3xl: 1.875rem;    /* 30px - 大标题 */
  --font-size-4xl: 2.25rem;     /* 36px - 页面标题 */
  --font-size-5xl: 3rem;        /* 48px - 展示级标题 */
  --font-size-6xl: 3.75rem;     /* 60px - 超大标题 */

  /* 行高系统 */
  --line-height-tight: 1.25;    /* 紧凑 - UI元素 */
  --line-height-normal: 1.5;    /* 正常 - 正文 */
  --line-height-relaxed: 1.75;  /* 宽松 - 长篇幅内容 */
  
  /* 字重 */
  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;
}
```

### 间距系统
**8点网格系统**：基于 8px 的基础单位，确保一致的视觉节奏
```css
:root {
  /* 间距比例 - 基于 8px (0.5rem) */
  --space-1: 0.25rem;   /* 4px - 紧密间距 */
  --space-2: 0.5rem;    /* 8px - 基础单位 */
  --space-3: 0.75rem;   /* 12px - 小间距 */
  --space-4: 1rem;      /* 16px - 常规间距 */
  --space-5: 1.25rem;   /* 20px */
  --space-6: 1.5rem;    /* 24px - 中间距 */
  --space-8: 2rem;      /* 32px */
  --space-10: 2.5rem;   /* 40px */
  --space-12: 3rem;     /* 48px - 大间距 */
  --space-16: 4rem;     /* 64px - 超大间距 */
  --space-20: 5rem;     /* 80px */
  --space-24: 6rem;     /* 96px */
  --space-32: 8rem;     /* 128px */
}
```

### 边框与圆角
```css
:root {
  /* 边框宽度 */
  --border-width-thin: 1px;
  --border-width-medium: 2px;
  --border-width-thick: 3px;

  /* 圆角系统 */
  --radius-none: 0;
  --radius-xs: 0.125rem;   /* 2px - 按钮、输入框 */
  --radius-sm: 0.25rem;    /* 4px - 卡片、徽章 */
  --radius-md: 0.375rem;   /* 6px - 中等组件 */
  --radius-lg: 0.5rem;     /* 8px - 卡片、模态框 */
  --radius-xl: 0.75rem;    /* 12px - 大组件 */
  --radius-2xl: 1rem;      /* 16px - 展示组件 */
  --radius-full: 9999px;   /* 圆形元素 */
}
```

### 阴影与海拔
```css
:root {
  /* 阴影系统 - 模拟物理高度 */
  --shadow-xs: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-sm: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1);
  --shadow-xl: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1);
  --shadow-2xl: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  
  /* 内阴影 */
  --shadow-inner: inset 0 2px 4px 0 rgba(0, 0, 0, 0.06);
}
```

## 🏗️ 组件设计规范

### 按钮组件
```vue
<!-- 使用示例 -->
<template>
  <!-- 主要按钮 -->
  <m-button variant="primary" size="md" :loading="isLoading">
    Create Account
  </m-button>
  
  <!-- 次要按钮 -->
  <m-button variant="secondary" size="sm">
    Cancel
  </m-button>
  
  <!-- 危险操作按钮 -->
  <m-button variant="danger" size="xs">
    Delete
  </m-button>
  
  <!-- 文字按钮 -->
  <m-button variant="text" size="md">
    Learn More
  </m-button>
  
  <!-- 图标按钮 -->
  <m-button variant="icon" size="md">
    <PlusIcon />
  </m-button>
</template>
```

### 表单组件
```vue
<!-- 使用示例 -->
<template>
  <!-- 标签与输入框 -->
  <m-form-group label="Company Name" required>
    <m-input 
      v-model="companyName"
      placeholder="Enter company name"
      :error="formErrors.companyName"
    />
    <m-form-hint v-if="formErrors.companyName">
      {{ formErrors.companyName }}
    </m-form-hint>
    <m-form-hint v-else>
      Enter your company's legal name
    </m-form-hint>
  </m-form-group>
  
  <!-- 选择器 -->
  <m-form-group label="Country">
    <m-select 
      v-model="selectedCountry"
      :options="countryOptions"
      placeholder="Select a country"
    />
  </m-form-group>
  
  <!-- 复选框组 -->
  <m-form-group label="Notification Preferences">
    <m-checkbox-group v-model="notifications">
      <m-checkbox value="email">Email notifications</m-checkbox>
      <m-checkbox value="sms">SMS notifications</m-checkbox>
      <m-checkbox value="push">Push notifications</m-checkbox>
    </m-checkbox-group>
  </m-form-group>
</template>
```

### 卡片与容器
```vue
<!-- 使用示例 -->
<template>
  <!-- 基础卡片 -->
  <m-card>
    <template #header>
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900">Account Overview</h3>
        <m-button variant="text" size="sm">View Details</m-button>
      </div>
    </template>
    <template #default>
      <div class="space-y-4">
        <!-- 卡片内容 -->
      </div>
    </template>
    <template #footer>
      <div class="flex justify-end space-x-2">
        <m-button variant="secondary" size="sm">Cancel</m-button>
        <m-button variant="primary" size="sm">Save Changes</m-button>
      </div>
    </template>
  </m-card>
  
  <!-- 统计卡片 -->
  <m-stat-card 
    title="Active Users"
    :value="1234"
    :change="12"
    change-type="increase"
    icon="UserGroupIcon"
  />
</template>
```

### 数据展示组件
```vue
<!-- 使用示例 -->
<template>
  <!-- 表格 -->
  <m-table 
    :columns="columns"
    :data="userList"
    :loading="isLoading"
    :pagination="pagination"
    @row-click="handleRowClick"
    @sort-change="handleSortChange"
  >
    <template #action="{ row }">
      <div class="flex space-x-2">
        <m-button variant="text" size="xs" @click="editUser(row)">
          Edit
        </m-button>
        <m-button variant="danger-text" size="xs" @click="deleteUser(row)">
          Delete
        </m-button>
      </div>
    </template>
  </m-table>
  
  <!-- 空状态 -->
  <m-empty-state
    v-if="!userList.length && !isLoading"
    title="No users found"
    description="Get started by creating a new user account"
    icon="UserAddIcon"
  >
    <template #action>
      <m-button variant="primary" @click="createUser">
        Add User
      </m-button>
    </template>
  </m-empty-state>
</template>
```

## 🎯 布局系统

### 响应式网格布局
```css
/* 使用 Tailwind 网格类 */
.grid-container {
  @apply grid gap-4;
}

/* 移动端：单列 */
@media (max-width: 639px) {
  .grid-container {
    @apply grid-cols-1;
  }
}

/* 平板端：2列 */
@media (min-width: 640px) and (max-width: 1023px) {
  .grid-container {
    @apply grid-cols-2;
  }
}

/* 桌面端：3-4列 */
@media (min-width: 1024px) {
  .grid-container {
    @apply grid-cols-3;
  }
}

/* 大桌面端：4列 */
@media (min-width: 1280px) {
  .grid-container {
    @apply grid-cols-4;
  }
}
```

### 页面布局结构
```vue
<!-- 应用布局组件 -->
<template>
  <!-- 根容器 -->
  <div class="min-h-screen bg-gray-50 flex flex-col">
    
    <!-- 顶部导航 -->
    <m-topbar :user="currentUser" @logout="handleLogout" />
    
    <div class="flex flex-1">
      <!-- 侧边栏 -->
      <m-sidebar :menu-items="menuItems" />
      
      <!-- 主内容区 -->
      <main class="flex-1 p-6 overflow-auto">
        <!-- 页面标题区 -->
        <div v-if="title" class="mb-6">
          <h1 class="text-2xl font-bold text-gray-900">{{ title }}</h1>
          <p v-if="subtitle" class="text-gray-600 mt-1">{{ subtitle }}</p>
        </div>
        
        <!-- 面包屑导航 -->
        <m-breadcrumb v-if="breadcrumbs.length" :items="breadcrumbs" class="mb-6" />
        
        <!-- 内容容器 -->
        <div class="content-area">
          <slot />
        </div>
      </main>
    </div>
    
    <!-- 页脚 -->
    <m-footer />
  </div>
</template>
```

## 📱 响应式设计策略

### 断点系统
```css
/* 断点定义 */
--breakpoint-sm: 640px;   /* 小型设备 */
--breakpoint-md: 768px;   /* 中型设备 */
--breakpoint-lg: 1024px;  /* 大型设备 */
--breakpoint-xl: 1280px;  /* 超大设备 */
--breakpoint-2xl: 1536px; /* 特大设备 */

/* 移动优先媒体查询 */
@mixin mobile {
  @media (max-width: 639px) {
    @content;
  }
}

@mixin tablet {
  @media (min-width: 640px) {
    @content;
  }
}

@mixin desktop {
  @media (min-width: 1024px) {
    @content;
  }
}
```

### 组件响应式行为
| 组件 | 移动端 | 平板端 | 桌面端 |
|------|--------|--------|--------|
| 导航栏 | 汉堡菜单 | 精简导航 | 完整导航 |
| 表格 | 卡片列表 | 简化表格 | 完整表格 |
| 模态框 | 全屏模态 | 居中模态 | 标准模态 |
| 表单 | 单列布局 | 双列布局 | 响应式多列 |

## ♿ 可访问性设计

### WCAG 2.1 AA 合规性
1. **色彩对比度**: 文本最小 4.5:1，大文本 3:1
2. **键盘导航**: 所有功能可通过键盘操作
3. **屏幕阅读器**: 语义化HTML和ARIA标签
4. **焦点管理**: 清晰焦点指示器和逻辑顺序

### 具体实现
```vue
<template>
  <!-- 使用语义化标签 -->
  <header role="banner">
    <nav role="navigation" aria-label="Main navigation">
      <!-- 导航内容 -->
    </nav>
  </header>
  
  <!-- 为图片提供Alt文本 -->
  <img 
    src="/logo.png" 
    alt="Mengri Flow Logo - Low-code platform for workflow automation"
    role="img"
  />
  
  <!-- 表单可访问性 -->
  <label for="username" class="sr-only">Username</label>
  <input 
    id="username"
    type="text"
    aria-describedby="username-help"
    aria-required="true"
  />
  <div id="username-help" class="text-sm text-gray-600">
    Enter your username or email address
  </div>
  
  <!-- ARIA live区域用于动态内容 -->
  <div 
    aria-live="polite" 
    aria-atomic="true"
    class="sr-only"
  >
    {{ screenReaderAnnouncement }}
  </div>
</template>
```

## 🌈 动效与微交互

### 过渡动画
```css
/* 过渡时间 */
--transition-fast: 150ms ease;
--transition-normal: 300ms ease;
--transition-slow: 500ms ease;

/* 动画曲线 */
--ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
--ease-out: cubic-bezier(0, 0, 0.2, 1);
--ease-in: cubic-bezier(0.4, 0, 1, 1);
```

### 组件动效
```vue
<script setup>
import { ref, watch } from 'vue'

// 加载状态动画
const isLoading = ref(false)
const startLoading = () => {
  isLoading.value = true
  // 加载完成后的逻辑
}
</script>

<template>
  <!-- 按钮点击反馈 -->
  <m-button 
    @click="startLoading" 
    :loading="isLoading"
    class="transition-all duration-300 hover:scale-105 active:scale-95"
  >
    <span v-if="!isLoading">Submit</span>
    <span v-else>Processing...</span>
  </m-button>
  
  <!-- 列表项进入动画 -->
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
  transition: all 300ms ease;
}

.list-enter-from,
.list-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
```

## 🎨 视觉设计令牌更新

现在更新项目的 Tailwind 配置以匹配设计系统：



<｜DSML｜function_calls>
<｜DSML｜invoke name="replace_in_file">
<｜DSML｜parameter name="filePath" string="true">c:\work\mengri\mengri-flow\web\tailwind.config.js