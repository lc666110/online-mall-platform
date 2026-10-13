<script setup lang="ts">
defineProps<{ loading?: boolean; error?: string; empty?: boolean }>();
</script>
<template>
  <div
    v-if="loading || error || empty"
    class="state-message"
    :role="error ? 'alert' : 'status'"
  >
    <span v-if="loading">正在加载好物…</span
    ><span v-else-if="error">{{ error }}</span
    ><span v-else>这里还没有内容，去探索喜欢的好物吧。</span><slot />
  </div>
</template>
