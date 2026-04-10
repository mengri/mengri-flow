<template>
  <el-dropdown
    trigger="click"
    @command="handleLanguageChange"
    class="language-switcher"
  >
    <el-button
      type="default"
      size="small"
      class="language-btn"
      :title="$t('common.language')"
    >
      <el-icon class="language-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <path d="M2 12h20"/>
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
        </svg>
      </el-icon>
      <span class="language-text">{{ currentLanguageLabel }}</span>
      <el-icon class="arrow-icon"><ArrowDown /></el-icon>
    </el-button>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="lang in languages"
          :key="lang.value"
          :command="lang.value"
          :class="{ 'is-active': locale === lang.value }"
        >
          <span class="lang-flag">{{ lang.flag }}</span>
          <span class="lang-label">{{ lang.label }}</span>
          <el-icon v-if="locale === lang.value" class="check-icon"><Check /></el-icon>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { ArrowDown, Check } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';

const { locale, t } = useI18n();

const languages = [
  { value: 'en', label: 'English', flag: '🇺🇸' },
  { value: 'zh', label: '中文', flag: '🇨🇳' },
];

const currentLanguageLabel = computed(() => {
  const lang = languages.find(l => l.value === locale.value);
  return lang?.label || 'English';
});

const handleLanguageChange = (lang: string) => {
  if (locale.value === lang) return;
  
  locale.value = lang;
  localStorage.setItem('locale', lang);
  
  // 更新 Element Plus 语言
  // 注意：Element Plus 语言需要单独处理，可以在 App.vue 中监听 locale 变化
  
  ElMessage.success(
    lang === 'zh' ? '已切换到中文' : 'Switched to English'
  );
};
</script>

<style scoped>
.language-switcher {
  display: inline-flex;
}

.language-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  font-size: 14px;
}

.language-icon {
  font-size: 16px;
  width: 16px;
  height: 16px;
}

.language-text {
  min-width: 50px;
  text-align: left;
}

.arrow-icon {
  font-size: 12px;
  margin-left: 2px;
}

:deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  min-width: 140px;
}

:deep(.el-dropdown-menu__item.is-active) {
  color: var(--el-color-primary);
  background-color: var(--el-color-primary-light-9);
}

.lang-flag {
  font-size: 16px;
}

.lang-label {
  flex: 1;
}

.check-icon {
  font-size: 14px;
  color: var(--el-color-primary);
}
</style>
