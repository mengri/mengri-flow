<template>
  <div class="stat-card" :class="[`variant-${variant}`, `color-${color}`]">
    <!-- 统计卡片头部 -->
    <div class="stat-header">
      <div class="stat-icon-wrapper">
        <slot name="icon">
          <component
            v-if="icon"
            :is="icon"
            class="stat-icon"
          />
          <div v-else class="stat-placeholder">
            {{ getInitials(title) }}
          </div>
        </slot>
      </div>
      
      <div class="stat-info">
        <h3 class="stat-title">{{ title }}</h3>
        <div v-if="subtitle" class="stat-subtitle">
          {{ subtitle }}
        </div>
      </div>
    
      <div class="stat-actions">
        <button
          v-if="showRefresh"
          class="stat-action"
          @click="$emit('refresh')"
          aria-label="Refresh"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
        
        <button
          v-if="showMore"
          class="stat-action"
          @click="$emit('more')"
          aria-label="More options"
        >
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
          </svg>
        </button>
      </div>
    </div>
    
    <!-- 统计值 -->
    <div class="stat-value-area">
      <div class="stat-value-wrapper">
        <span class="stat-value">{{ formattedValue }}</span>
        <span v-if="unit" class="stat-unit">{{ unit }}</span>
      </div>
      
      <!-- 变化指示器 -->
      <div v-if="change !== undefined" class="stat-change" :class="changeClass">
        <component
          :is="changeIcon"
          class="stat-change-icon"
        />
        <span class="stat-change-value">{{ formattedChange }}</span>
        <span v-if="changePeriod" class="stat-change-period">/ {{ changePeriod }}</span>
      </div>
    </div>
    
    <!-- 统计描述 -->
    <div v-if="description" class="stat-description">
      {{ description }}
    </div>
    
    <!-- 统计进度条（可选） -->
    <div v-if="progress !== undefined" class="stat-progress">
      <div class="progress-track">
        <div
          class="progress-fill"
          :style="{ width: `${Math.min(progress, 100)}%` }"
          :class="progressClass"
        />
      </div>
      <div v-if="progressLabel" class="progress-label">
        {{ progressLabel }}
      </div>
    </div>
    
    <!-- 统计趋势图（可选） -->
    <div v-if="trendData && showTrend" class="stat-trend">
      <canvas ref="trendChart" :height="trendHeight" />
    </div>
    
    <!-- 统计详情链接 -->
    <slot name="footer">
      <a
        v-if="viewDetailsLink"
        :href="viewDetailsLink"
        class="stat-footer-link"
        @click.prevent="$emit('view-details')"
      >
        <span>View details</span>
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </a>
    </slot>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import Chart from 'chart.js/auto'

// Props
const props = withDefaults(defineProps<{
  title: string
  subtitle?: string
  value: string | number
  unit?: string
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'danger' | 'info'
  variant?: 'default' | 'compact' | 'chart'
  icon?: any
  change?: number
  changeType?: 'increase' | 'decrease' | 'neutral'
  changePeriod?: string
  description?: string
  progress?: number
  progressLabel?: string
  trendData?: number[]
  showTrend?: boolean
  trendHeight?: number
  showRefresh?: boolean
  showMore?: boolean
  viewDetailsLink?: string
  loading?: boolean
}>(), {
  color: 'primary',
  variant: 'default',
  changeType: 'neutral',
  trendHeight: 40,
  showTrend: false,
  showRefresh: false,
  showMore: false,
  loading: false,
})

// Emits
const emit = defineEmits<{
  'refresh': []
  'more': []
  'view-details': []
}>()

// Refs
const trendChart = ref<HTMLCanvasElement | null>(null)
let chartInstance: Chart | null = null

// Computed
const formattedValue = computed(() => {
  const value = props.value
  
  if (typeof value === 'number') {
    if (value >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`
    } else if (value >= 1000) {
      return `${(value / 1000).toFixed(1)}K`
    }
    return value.toLocaleString()
  }
  
  return value
})

const formattedChange = computed(() => {
  if (props.change === undefined) return ''
  
  const absChange = Math.abs(props.change)
  const prefix = props.changeType === 'increase' ? '+' : props.changeType === 'decrease' ? '-' : ''
  
  if (absChange >= 1000000) {
    return `${prefix}${(absChange / 1000000).toFixed(1)}M`
  } else if (absChange >= 1000) {
    return `${prefix}${(absChange / 1000).toFixed(1)}K`
  }
  return `${prefix}${absChange.toFixed(1)}`
})

const changeClass = computed(() => {
  if (props.changeType === 'increase') {
    return 'change-increase'
  } else if (props.changeType === 'decrease') {
    return 'change-decrease'
  }
  return 'change-neutral'
})

const changeIcon = computed(() => {
  if (props.changeType === 'increase') {
    return 'TrendingUpIcon'
  } else if (props.changeType === 'decrease') {
    return 'TrendingDownIcon'
  }
  return 'MinusIcon'
})

const progressClass = computed(() => {
  const progress = props.progress || 0
  
  if (progress >= 90) return 'progress-danger'
  if (progress >= 75) return 'progress-warning'
  if (progress >= 50) return 'progress-info'
  return 'progress-success'
})

// Methods
const getInitials = (text: string) => {
  return text
    .split(' ')
    .map(word => word[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)
}

const renderChart = () => {
  if (!trendChart.value || !props.trendData || !props.showTrend) return
  
  // 销毁现有图表
  if (chartInstance) {
    chartInstance.destroy()
  }
  
  const ctx = trendChart.value.getContext('2d')
  if (!ctx) return
  
  // 创建图表数据
  const data = props.trendData
  const labels = Array.from({ length: data.length }, (_, i) => i + 1)
  
  const lineColor = getComputedStyle(document.documentElement)
    .getPropertyValue(`--color-${props.color}-500`) || '#3b82f6'
  
  const gradient = ctx.createLinearGradient(0, 0, 0, props.trendHeight)
  gradient.addColorStop(0, `${lineColor}20`)
  gradient.addColorStop(1, `${lineColor}00`)
  
  chartInstance = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        data,
        borderColor: lineColor,
        backgroundColor: gradient,
        borderWidth: 2,
        tension: 0.3,
        fill: true,
        pointRadius: 0,
        pointHoverRadius: 3,
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          enabled: false
        }
      },
      scales: {
        x: {
          display: false
        },
        y: {
          display: false,
          beginAtZero: true
        }
      }
    }
  })
}

// Lifecycle
onMounted(() => {
  if (props.showTrend) {
    nextTick(() => {
      renderChart()
    })
  }
})

onUnmounted(() => {
  if (chartInstance) {
    chartInstance.destroy()
  }
})

// Watchers
watch(() => props.trendData, () => {
  if (props.showTrend) {
    nextTick(() => {
      renderChart()
    })
  }
})

watch(() => props.showTrend, (show) => {
  if (show) {
    nextTick(() => {
      renderChart()
    })
  } else if (chartInstance) {
    chartInstance.destroy()
    chartInstance = null
  }
})
</script>

<style scoped>
.stat-card {
  @apply bg-white border border-gray-200 rounded-lg p-5 transition-all duration-200 hover:shadow-sm;
}

/* 变体样式 */
.stat-card.variant-compact {
  @apply p-3;
}

.stat-card.variant-chart {
  @apply p-0 overflow-hidden;
}

.stat-card.variant-chart .stat-header {
  @apply px-5 pt-5;
}

.stat-card.variant-chart .stat-value-area {
  @apply px-5 pb-2;
}

.stat-card.variant-chart .stat-trend {
  @apply mt-2;
}

/* 颜色变体 */
.stat-card.color-primary {
  --stat-color: var(--color-primary-500);
}

.stat-card.color-secondary {
  --stat-color: var(--color-secondary-500);
}

.stat-card.color-success {
  --stat-color: var(--color-success-500);
}

.stat-card.color-warning {
  --stat-color: var(--color-warning-500);
}

.stat-card.color-danger {
  --stat-color: var(--color-danger-500);
}

.stat-card.color-info {
  --stat-color: var(--color-info-500);
}

/* 头部样式 */
.stat-header {
  @apply flex items-start justify-between mb-3 gap-2;
}

.stat-icon-wrapper {
  @apply h-10 w-10 rounded-lg flex items-center justify-center flex-shrink-0;
  background-color: color-mix(in srgb, var(--stat-color) 10%, transparent);
  color: var(--stat-color);
}

.stat-icon {
  @apply h-5 w-5;
}

.stat-placeholder {
  @apply text-sm font-semibold;
}

.stat-info {
  @apply flex-1 min-w-0;
}

.stat-title {
  @apply text-sm font-semibold text-gray-900 truncate;
}

.stat-subtitle {
  @apply text-xs text-gray-500 mt-0.5;
}

.stat-actions {
  @apply flex gap-1 flex-shrink-0;
}

.stat-action {
  @apply p-1 rounded text-gray-400 hover:text-gray-600 hover:bg-gray-100;
}

/* 数值区域样式 */
.stat-value-area {
  @apply mb-2;
}

.stat-value-wrapper {
  @apply flex items-baseline gap-1;
}

.stat-value {
  @apply text-2xl font-bold text-gray-900;
}

.stat-unit {
  @apply text-sm text-gray-500 ml-1;
}

/* 变化指示器样式 */
.stat-change {
  @apply inline-flex items-center gap-1 mt-1 text-xs font-medium px-1.5 py-0.5 rounded-full;
}

.stat-change.change-increase {
  @apply bg-green-50 text-green-700;
}

.stat-change.change-decrease {
  @apply bg-red-50 text-red-700;
}

.stat-change.change-neutral {
  @apply bg-gray-100 text-gray-600;
}

.stat-change-icon {
  @apply h-3 w-3;
}

.stat-change-period {
  @apply text-gray-500;
}

/* 描述样式 */
.stat-description {
  @apply text-sm text-gray-600 mt-1;
}

/* 进度条样式 */
.stat-progress {
  @apply mt-3 space-y-1;
}

.progress-track {
  @apply h-2 bg-gray-200 rounded-full overflow-hidden;
}

.progress-fill {
  @apply h-full rounded-full transition-all duration-500;
}

.progress-fill.progress-success {
  background-color: var(--color-success-500);
}

.progress-fill.progress-info {
  background-color: var(--color-info-500);
}

.progress-fill.progress-warning {
  background-color: var(--color-warning-500);
}

.progress-fill.progress-danger {
  background-color: var(--color-danger-500);
}

.progress-label {
  @apply text-xs text-gray-500 text-right;
}

/* 趋势图样式 */
.stat-trend {
  @apply w-full mt-4;
  height: v-bind('trendHeight + "px"');
}

/* 底部链接样式 */
.stat-footer-link {
  @apply inline-flex items-center gap-1 mt-4 text-sm text-primary-600 hover:text-primary-800 no-underline;
}

/* 紧凑样式适配 */
.stat-card.variant-compact .stat-value {
  @apply text-xl;
}

.stat-card.variant-compact .stat-trend {
  @apply mt-2;
  height: v-bind('Math.max(20, trendHeight - 20) + "px"');
}

/* 加载状态 */
.stat-card.loading .stat-value,
.stat-card.loading .stat-change,
.stat-card.loading .stat-description {
  @apply animate-pulse bg-gray-200 text-transparent rounded;
}

.stat-card.loading .stat-value {
  @apply h-8 w-16;
}

.stat-card.loading .stat-change {
  @apply h-5 w-12;
}

.stat-card.loading .stat-description {
  @apply h-4 w-24;
}

/* 悬停效果 */
.stat-card:hover {
  @apply border-gray-300;
}

.stat-card:hover .stat-footer-link {
  @apply text-primary-700;
}

/* 可点击卡片 */
.stat-card.interactive {
  @apply cursor-pointer hover:shadow-md hover:border-primary-300 hover:-translate-y-1;
}

/* 响应式调整 */
@media (max-width: 640px) {
  .stat-card:not(.variant-compact) {
    @apply p-4;
  }
  
  .stat-value {
    @apply text-xl;
  }
}
</style>