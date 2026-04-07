<template>
  <component
    :is="componentType"
    :type="nativeType"
    :class="buttonClasses"
    :disabled="disabled || loading"
    :aria-label="ariaLabel"
    :aria-busy="loading"
    @click="handleClick"
    v-bind="{
      ...$attrs,
      ...(isLink ? { to, replace, activeClass, exactActiveClass } : {})
    }"
  >
    <!-- 加载状态 -->
    <div v-if="loading" class="button-loading">
      <slot name="loading">
        <svg class="button-spinner" fill="none" viewBox="0 0 24 24">
          <circle class="spinner-track" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" opacity="0.25" />
          <path class="spinner-arc" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </slot>
    </div>

    <!-- 图标区域 -->
    <span v-if="icon || $slots.icon" class="button-icon" :class="{ 'button-icon-only': !$slots.default }">
      <slot name="icon">
        <component v-if="icon" :is="icon" class="icon-svg" />
      </slot>
    </span>

    <!-- 文本内容 -->
    <span v-if="$slots.default" class="button-content">
      <slot />
    </span>

    <!-- 尾部图标/徽章 -->
    <span v-if="$slots.suffix || badge || suffixIcon" class="button-suffix">
      <slot name="suffix">
        <component
          v-if="suffixIcon"
          :is="suffixIcon"
          class="suffix-icon"
        />
        <span v-if="badge" class="button-badge">{{ badge }}</span>
      </slot>
    </span>

    <!-- 工具提示 -->
    <div
      v-if="tooltip && !disabled"
      class="button-tooltip"
      role="tooltip"
      :aria-label="tooltip"
    >
      {{ tooltip }}
    </div>
  </component>
</template>

<script setup lang="ts">
import { computed, useAttrs } from 'vue'

// Props
const props = withDefaults(defineProps<{
  // 基础类型
  variant?: 'primary' | 'secondary' | 'tertiary' | 'danger' | 'success' | 'warning' | 'info' | 'text' | 'link'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  shape?: 'default' | 'pill' | 'square' | 'circle'
  loading?: boolean
  disabled?: boolean
  block?: boolean
  // 内容
  icon?: any
  suffixIcon?: any
  badge?: string | number
  // 行为
  nativeType?: 'button' | 'submit' | 'reset'
  href?: string
  target?: string
  to?: string | object
  replace?: boolean
  activeClass?: string
  exactActiveClass?: string
  // 可访问性
  ariaLabel?: string
  tooltip?: string
}>(), {
  variant: 'primary',
  size: 'md',
  shape: 'default',
  nativeType: 'button'
})

// Emits
const emit = defineEmits<{
  'click': [event: MouseEvent]
}>()

// Attrs
const attrs = useAttrs()

// Computed
const componentType = computed(() => {
  if (props.to) return 'router-link'
  if (props.href) return 'a'
  return 'button'
})

const isLink = computed(() => props.to || props.href)

const buttonClasses = computed(() => [
  'm-button',
  `variant-${props.variant}`,
  `size-${props.size}`,
  `shape-${props.shape}`,
  {
    'block-full': props.block,
    'loading': props.loading,
    'disabled': props.disabled,
    'icon-only': !attrs.default && (props.icon || attrs.icon),
    'has-badge': props.badge !== undefined,
  }
])

// Methods
const handleClick = (event: MouseEvent) => {
  if (props.loading || props.disabled) {
    event.preventDefault()
    event.stopPropagation()
    return
  }
  
  emit('click', event)
  
  // 处理原生链接点击
  if (props.href && componentType.value === 'a') {
    if (props.target === '_blank') {
      window.open(props.href, props.target)
    } else {
      window.location.href = props.href
    }
  }
}
</script>

<style scoped>
.m-button {
  @apply relative inline-flex items-center justify-center font-medium text-center no-underline cursor-pointer select-none border transition-all duration-200;
  @apply focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2;
  user-select: none;
  touch-action: manipulation;
}

/* 大小样式 */
.m-button.size-xs {
  @apply text-xs px-2 py 1.5;
  min-height: 1.75rem;
  gap: 0.25rem;
}

.m-button.size-sm {
  @apply text-sm px-3 py-2;
  min-height: 2rem;
  gap: 0.375rem;
}

.m-button.size-md {
  @apply text-sm px-4 py-2.5;
  min-height: 2.5rem;
  gap: 0.5rem;
}

.m-button.size-lg {
  @apply text-base px-5 py-3;
  min-height: 3rem;
  gap: 0.625rem;
}

.m-button.size-xl {
  @apply text-base px-6 py-3;
  min-height: 3.5rem;
  gap: 0.75rem;
}

/* 形状样式 */
.m-button.shape-default {
  @apply rounded;
}

.m-button.shape-pill {
  @apply rounded-full;
}

.m-button.shape-square {
  aspect-ratio: 1 / 1;
}

.m-button.shape-circle {
  @apply rounded-full;
  aspect-ratio: 1 / 1;
}

/* 变体样式 */
.m-button.variant-primary {
  @apply bg-primary-600 text-white border-primary-600;
  @apply hover:bg-primary-700 hover:border-primary-700;
  @apply active:bg-primary-800 active:border-primary-800;
  @apply focus-visible:ring-primary-500;
}

.m-button.variant-secondary {
  @apply bg-gray-100 text-gray-900 border-gray-200;
  @apply hover:bg-gray-200 hover:border-gray-300;
  @apply active:bg-gray-300 active:border-gray-400;
  @apply focus-visible:ring-gray-400;
}

.m-button.variant-tertiary {
  @apply bg-transparent text-primary-600 border-gray-300;
  @apply hover:bg-primary-50 hover:border-primary-500 hover:text-primary-700;
  @apply active:bg-primary-100 active:border-primary-600 active:text-primary-800;
  @apply focus-visible:ring-primary-500;
}

.m-button.variant-danger {
  @apply bg-danger-600 text-white border-danger-600;
  @apply hover:bg-danger-700 hover:border-danger-700;
  @apply active:bg-danger-800 active:border-danger-800;
  @apply focus-visible:ring-danger-500;
}

.m-button.variant-success {
  @apply bg-success-600 text-white border-success-600;
  @apply hover:bg-success-700 hover:border-success-700;
  @apply active:bg-success-800 active:border-success-800;
  @apply focus-visible:ring-success-500;
}

.m-button.variant-warning {
  @apply bg-warning-600 text-white border-warning-600;
  @apply hover:bg-warning-700 hover:border-warning-700;
  @apply active:bg-warning-800 active:border-warning-800;
  @apply focus-visible:ring-warning-500;
}

.m-button.variant-info {
  @apply bg-info-600 text-white border-info-600;
  @apply hover:bg-info-700 hover:border-info-700;
  @apply active:bg-info-800 active:border-info-800;
  @apply focus-visible:ring-info-500;
}

.m-button.variant-text {
  @apply bg-transparent text-gray-700 border-transparent;
  @apply hover:bg-gray-100 hover:text-gray-900;
  @apply active:bg-gray-200 active:text-gray-900;
  @apply focus-visible:ring-gray-400;
}

.m-button.variant-link {
  @apply bg-transparent text-primary-600 border-transparent underline underline-offset-2 decoration-from-font px-0;
  @apply hover:text-primary-700 hover:decoration-2;
  @apply active:text-primary-800 active:decoration-2;
  @apply focus-visible:ring-primary-500 focus-visible:no-underline;
}

/* 特殊状态 */
.m-button.disabled,
.m-button:disabled {
  @apply opacity-50 cursor-not-allowed pointer-events-none;
}

.m-button.loading {
  @apply pointer-events-none;
}

/* 块级按钮 */
.m-button.block-full {
  @apply w-full;
}

/* 内容布局 */
.m-button.icon-only {
  @apply px-0;
  aspect-ratio: 1 / 1;
}

.m-button.icon-only.size-xs {
  @apply w-6 h-6;
}

.m-button.icon-only.size-sm {
  @apply w-8 h-8;
}

.m-button.icon-only.size-md {
  @apply w-10 h-10;
}

.m-button.icon-only.size-lg {
  @apply w-12 h-12;
}

.m-button.icon-only.size-xl {
  @apply w-14 h-14;
}

.button-loading {
  @apply absolute inset-0 flex items-center justify-center bg-inherit;
}

.button-spinner {
  @apply h-4 w-4 animate-spin;
}

.spinner-track {
  stroke: currentColor;
  opacity: 0.25;
}

.spinner-arc {
  fill: currentColor;
}

/* 图标样式 */
.button-icon {
  @apply flex items-center justify-center flex-shrink-0;
}

.button-icon.icon-only {
  @apply mx-auto;
}

.icon-svg {
  @apply h-4 w-4;
}

.m-button.size-xs .icon-svg {
  @apply h-3 w-3;
}

.m-button.size-sm .icon-svg {
  @apply h-4 w-4;
}

.m-button.size-md .icon-svg {
  @apply h-5 w-5;
}

.m-button.size-lg .icon-svg {
  @apply h-6 w-6;
}

.m-button.size-xl .icon-svg {
  @apply h-7 w-7;
}

/* 内容区域 */
.button-content {
  @apply flex-1 whitespace-nowrap overflow-hidden text-ellipsis;
}

/* 尾部内容 */
.button-suffix {
  @apply flex items-center justify-center flex-shrink-0 ml-1;
}

.suffix-icon {
  @apply h-4 w-4;
}

.button-badge {
  @apply ml-1 px-1.5 py-0.5 text-xs font-medium rounded-full bg-primary-100 text-primary-800 min-w-[1.25rem] flex items-center justify-center;
}

/* 方形和圆形按钮的内容调整 */
.m-button.shape-square .button-content,
.m-button.shape-circle .button-content {
  @apply hidden;
}

.m-button.shape-square .button-icon,
.m-button.shape-circle .button-icon {
  @apply m-auto;
}

/* 工具提示 */
.button-tooltip {
  @apply absolute bottom-full left-1/2 transform -translate-x-1/2 mb-1 px,2 py,1 text-xs font-medium bg-gray-900 text-white rounded whitespace-nowrap opacity-0 transition-opacity duration-200 pointer-events-none z-50;
  @apply before:absolute before:top-full before:left-1/2 before:-translate-x-1/2 before:border-4 before:border-transparent before:border-t-gray-900;
}

.m-button:hover .button-tooltip {
  @apply opacity-100;
}

/* 加载状态下的文字隐藏 */
.m-button.loading .button-content,
.m-button.loading .button-icon:not(.button-loading),
.m-button.loading .button-suffix {
  @apply invisible;
}

/* 响应式调整 */
@media (max-width: 640px) {
  .m-button:not(.icon-only) {
    @apply px-3;
  }
  
  .m-button.size-lg:not(.icon-only),
  .m-button.size-xl:not(.icon-only) {
    @apply px-4;
  }
}

/* 颜色系统变量覆盖 */
.m-button.variant-primary {
  --button-color: var(--color-primary-600);
  --button-hover-color: var(--color-primary-700);
  --button-active-color: var(--color-primary-800);
}

.m-button.variant-secondary {
  --button-color: var(--color-gray-100);
  --button-hover-color: var(--color-gray-200);
  --button-active-color: var(--color-gray-300);
}

.m-button.variant-danger {
  --button-color: var(--color-danger-600);
  --button-hover-color: var(--color-danger-700);
  --button-active-color: var(--color-danger-800);
}

/* 深色主题适配 */
@media (prefers-color-scheme: dark) {
  .m-button.variant-secondary {
    @apply bg-gray-800 text-gray-200 border-gray-700;
    @apply hover:bg-gray-700 hover:border-gray-600;
    @apply active:bg-gray-600 active:border-gray-500;
  }
  
  .m-button.variant-text {
    @apply text-gray-300;
    @apply hover:bg-gray-800 hover:text-gray-100;
    @apply active:bg-gray-700 active:text-gray-100;
  }
}

/* 减少动效支持 */
@media (prefers-reduced-motion: reduce) {
  .m-button {
    transition: none;
  }
  
  .button-tooltip {
    transition: none;
  }
}

/* 高对比度支持 */
@media (prefers-contrast: high) {
  .m-button {
    border-width: 2px;
  }
  
  .m-button:focus-visible {
    outline: 3px solid currentColor;
    outline-offset: 2px;
  }
}
</style>