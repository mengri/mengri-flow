import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'
import 'element-plus/dist/index.css'
import router from './router'
import i18n from './i18n'
import App from './App.vue'
import './assets/css/app.css'
import './assets/styles/main.css'

const app = createApp(App)

// 获取当前语言设置
const savedLocale = localStorage.getItem('locale') || 'en'

// Element Plus 语言映射
const elementLocales: Record<string, any> = {
  zh: zhCn,
  en: en,
}

app.use(createPinia())
app.use(router)
app.use(i18n)
app.use(ElementPlus, { locale: elementLocales[savedLocale] || en })

app.mount('#app')
