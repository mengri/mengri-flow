<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import AppNavbar from '@/components/ui/AppNavbar.vue'

const route = useRoute()

const currentYear = computed(() => new Date().getFullYear())
const showFooter = computed(() => route.meta?.showFooter !== false)
</script>

<template>
  <div class="app-layout">
    <!-- Top navigation (includes breadcrumb) -->
    <AppNavbar />

    <!-- Main content -->
    <main class="main-content">
      <div class="content-container">
        <RouterView />
      </div>
    </main>

    <!-- Footer -->
    <footer v-if="showFooter" class="app-footer">
      <div class="content-container">
        <span class="text-sm text-gray-500">
          <span class="font-semibold text-gray-700">Mengri Flow</span> &copy; {{ currentYear }}
        </span>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #f6f8fa;
}

/* Main content */
.main-content {
  flex: 1;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.content-container {
  max-width: 1280px;
  margin: 0 auto;
  padding: 24px;
}

/* Footer */
.app-footer {
  padding: 16px 24px;
  border-top: 1px solid #d8dee4;
  background: #fff;
  flex-shrink: 0;
  text-align: center;
}

/* Responsive */
@media (max-width: 767px) {
  .content-container {
    padding: 16px;
  }
}
</style>
