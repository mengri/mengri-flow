<script setup lang="ts">
interface Props {
  size?: 'sm' | 'md' | 'lg'
}

withDefaults(defineProps<Props>(), {
  size: 'md',
})

const sizeMap = {
  sm: { icon: 'h-6 w-6', primary: 'text-sm font-bold', secondary: 'text-xs font-semibold' },
  md: { icon: 'h-8 w-8', primary: 'text-lg font-bold', secondary: 'text-base font-semibold' },
  lg: { icon: 'h-10 w-10', primary: 'text-xl font-bold', secondary: 'text-lg font-semibold' },
}
</script>

<template>
  <div class="app-logo inline-flex items-center gap-2.5 no-underline hover:no-underline">
    <img src="/logo.svg" :class="[sizeMap[size].icon]" alt="Mengri Flow" />
    <span v-if="size !== 'sm'" class="app-logo-text">
      <span :class="[sizeMap[size].primary, 'text-gray-900 leading-tight']">Mengri</span>
      <span :class="[sizeMap[size].secondary, 'text-primary-600 leading-tight']">Flow</span>
    </span>
  </div>
</template>

<style scoped>
.app-logo {
  cursor: pointer;
  user-select: none;
}

.app-logo-text {
  display: flex;
  align-items: baseline;
  gap: 1px;
}
</style>
