import { createI18n } from 'vue-i18n';

// 导入语言包
import en from './locales/en.json';
import zh from './locales/zh.json';

const i18n = createI18n({
  legacy: false, // 使用 Composition API 风格
  globalInjection: true, // 全局注入 $t 方法
  locale: localStorage.getItem('locale') || 'en', // 默认语言
  fallbackLocale: 'en', // 如果没找到语言包，使用兜底语言
  messages: {
    en,
    zh,
  },
});

export default i18n;