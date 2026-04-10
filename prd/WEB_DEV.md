# 前端开发指南

## 技术栈

Vue 3 (Composition API) | TypeScript | Vite | Pinia | Tailwind CSS | Element Plus

## 开发命令

```bash
cd web

# 开发模式 (热重载，端口 3000)
npm run dev

# 构建生产版本 → web/dist/
npm run build

# 预览构建结果
npm run preview

# Linting
npm run lint
```

## 目录结构

```
web/src/
  api/                API 调用函数 (每个领域一个文件)
  composables/        可复用逻辑 hooks (useUser, useAuth 等)
  components/common/  共享基础组件
  stores/             Pinia 状态管理
  types/              TypeScript 类型定义
  utils/              Axios 实例、工具函数
  views/              页面组件
  router/             Vue Router 配置
```

## 代码规范

- **始终使用** `<script setup lang="ts">`，禁止使用 `any`
- **复杂逻辑**放在 `src/composables/`，不在组件内联实现
- **状态管理**：Pinia setup stores (`defineStore('name', () => {...})`)
- **Props 类型**：`defineProps<{ visible: boolean }>()`
- **Emits 类型**：`defineEmits<{ 'update:visible': [value: boolean] }>()`
- **UI 框架**：Element Plus + Tailwind CSS
- **Axios**：位于 `src/utils/request.ts`，baseURL 为 `/api/v1`，拦截器检查 `code !== 0`

### 命名约定

- 变量/函数：camelCase
- 组件/类型：PascalCase
- CSS/资源：kebab-case

## API 响应格式

所有接口返回统一格式：

```json
{
  "code": 0,
  "data": {},
  "msg": "success"
}
```

`code === 0` 表示成功。Axios 拦截器全局处理非零错误码。