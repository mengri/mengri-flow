<template>
  <div>
    <label for="lang">Choose Language:</label>
    <select id="lang" @change="onChange" v-model="selectedLang">
      <option value="en">English</option>
      <option value="zh">中文</option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { locale } = useI18n();
const selectedLang = ref(locale.value);

const onChange = (event: Event) => {
  const lang = (event.target as HTMLSelectElement).value;
  locale.value = lang;
  localStorage.setItem('locale', lang);
};
</script>