<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import {
  NCard,
  NButton,
  NDataTable,
  NAlert,
  NModal,
  NInput,
  NSelect,
  NTag,
  NPagination,
  useMessage,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";
import type { Order, Page } from "../../../shared/types";
import { money, statusLabel } from "../../../shared/helpers";
import { api } from "../api";
const list = ref<Order[]>([]),
  selected = ref<Order | null>(null),
  show = ref(false),
  status = ref<string | null>(null),
  carrier = ref(""),
  tracking = ref(""),
  page = ref(1),
  total = ref(0),
  busy = ref(false),
  loading = ref(false),
  error = ref("");
const message = useMessage();
async function load() {
  loading.value = true;
  try {
    const p = await api.request<Page<Order>>(
      `/admin/orders?page=${page.value}`,
    );
    list.value = p.items;
    total.value = p.total;
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
async function edit(o: Order) {
  error.value = "";
  try {
    selected.value = await api.request<Order>(`/admin/orders/${o.id}`);
    status.value = null;
    carrier.value = o.carrier;
    tracking.value = o.tracking;
    show.value = true;
  } catch (e) {
    error.value = (e as Error).message;
  }
}
function options(o: Order) {
  return (
    (
      {
        pending: ["confirmed", "cancelled"],
        confirmed: ["shipped", "cancelled"],
        shipped: ["completed"],
      } as Record<string, string[]>
    )[o.status]?.map((value) => ({ value, label: statusLabel(value) })) ?? []
  );
}
async function save() {
  busy.value = true;
  try {
    if (!status.value) throw new Error("请选择目标状态");
    await api.request(`/admin/orders/${selected.value?.id}/status`, "PATCH", {
      status: status.value,
      carrier: carrier.value,
      tracking: tracking.value,
    });
    show.value = false;
    message.success("订单已更新");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
const columns: DataTableColumns<Order> = [
  { title: "订单号", key: "number", minWidth: 220 },
  { title: "用户 ID", key: "user_id" },
  { title: "金额", key: "total", render: (o) => money(o.total) },
  {
    title: "状态",
    key: "status",
    render: (o) =>
      h(NTag, { size: "small", bordered: false }, () => statusLabel(o.status)),
  },
  { title: "收货人", key: "recipient" },
  {
    title: "创建时间",
    key: "created_at",
    render: (o) => new Date(o.created_at).toLocaleString(),
  },
  {
    title: "操作",
    key: "actions",
    render: (o) =>
      h(
        NButton,
        { size: "small", onClick: () => edit(o) },
        () => "查看 / 处理",
      ),
  },
];
onMounted(load);
</script>
<template>
  <div class="view-title">
    <div>
      <p class="eyebrow">ORDER OPERATIONS</p>
      <h1>订单与发货</h1>
      <p>确认订单、填写物流并追踪履约。</p>
    </div>
    <NButton @click="load">刷新</NButton>
  </div>
  <NAlert v-if="error && !show" type="error">{{ error }}</NAlert
  ><NCard
    ><NDataTable
      :data="list"
      :columns="columns"
      :loading="loading"
      :scroll-x="1100" /><NPagination
      v-model:page="page"
      :item-count="total"
      :page-size="12"
      @update:page="load"
      style="margin-top: 20px" /></NCard
  ><NModal
    v-model:show="show"
    preset="card"
    title="订单处理"
    style="width: min(700px, 95vw)"
    ><template v-if="selected"
      ><p>
        <b>{{ selected.number }}</b> · {{ statusLabel(selected.status) }}
      </p>
      <p>
        {{ selected.recipient }} / {{ selected.phone }} / {{ selected.address }}
      </p>
      <div v-for="i in selected.items" :key="i.id" class="order-item">
        {{ i.name }} · {{ i.specification }} × {{ i.quantity }}
        <b>{{ money(i.price * i.quantity) }}</b>
      </div>
      <p>合计 {{ money(selected.total) }}</p>
      <NAlert v-if="error" type="error">{{ error }}</NAlert>
      <form v-if="options(selected).length" @submit.prevent="save">
        <label
          >变更状态<NSelect
            v-model:value="status"
            :options="options(selected)"
            placeholder="选择目标状态" /></label
        ><template v-if="status === 'shipped'"
          ><label
            >物流公司<NInput
              v-model:value="carrier"
              placeholder="物流公司"
              :maxlength="80" /></label
          ><label
            >运单号<NInput
              v-model:value="tracking"
              placeholder="运单号"
              :maxlength="100" /></label></template
        ><NButton
          attr-type="submit"
          type="primary"
          :loading="busy"
          style="margin-top: 20px"
          >保存订单状态</NButton
        >
      </form>
      <p v-else class="muted">此订单已结束，无可用状态操作。</p></template
    ></NModal
  >
</template>
