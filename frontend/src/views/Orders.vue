<script setup lang="ts">
import { ref, onMounted } from "vue";
import { api } from "../api";
import type { Order, Page } from "../../../shared/types";
import { money, statusLabel } from "../../../shared/helpers";
import StateMessage from "../components/StateMessage.vue";
const orders = ref<Order[]>([]),
  loading = ref(false),
  error = ref(""),
  page = ref(1),
  total = ref(0);
async function load() {
  loading.value = true;
  try {
    const r = await api.request<Page<Order>>(`/orders?page=${page.value}`);
    orders.value = r.items;
    total.value = r.total;
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
onMounted(load);
</script>
<template>
  <div class="page-title">
    <p class="eyebrow">MY ORDERS</p>
    <h1>每一份期待，都有迹可循。</h1>
  </div>
  <StateMessage
    :loading="loading"
    :error="error"
    :empty="!loading && !orders.length"
  />
  <article v-for="o in orders" :key="o.id" class="panel order-card">
    <div class="order-heading">
      <span>订单 {{ o.number }}</span
      ><span class="badge">{{ statusLabel(o.status) }}</span>
    </div>
    <p v-for="i in o.items" :key="i.id">
      {{ i.name }} · {{ i.specification }} × {{ i.quantity }}
    </p>
    <div class="order-heading">
      <strong>{{ money(o.total) }}</strong
      ><RouterLink class="button secondary" :to="`/orders/${o.id}`"
        >查看订单</RouterLink
      >
    </div>
  </article>
  <div v-if="total > 12" class="pagination">
    <button
      :disabled="page === 1"
      @click="
        page--;
        load();
      "
    >
      上一页</button
    ><span>{{ page }}</span
    ><button
      :disabled="page * 12 >= total"
      @click="
        page++;
        load();
      "
    >
      下一页
    </button>
  </div>
</template>
