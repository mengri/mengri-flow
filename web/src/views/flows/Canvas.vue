<template>
  <FlowCanvas :flow-id="flowId" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import FlowCanvas from '@/components/flow/FlowCanvas.vue'

const route = useRoute()
const flowId = computed(() => route.params.id as string)
</script>
