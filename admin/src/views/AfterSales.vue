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
  NPagination,
  useMessage,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";
import type { AfterSale, Page } from "../../../shared/types";
import { money, statusLabel } from "../../../shared/helpers";
import { api } from "../api";
const list = ref<AfterSale[]>([]),
  selected = ref<AfterSale | null>(null),
  show = ref(false),
  status = ref("approved"),
  note = ref(""),
  page = ref(1),
  total = ref(0),
  busy = ref(false),
  loading = ref(false),
  error = ref("");
const message = useMessage();
async function load() {
  loading.value = true;
  error.value = "";
  try {
    const p = await api.request<Page<AfterSale>>(
      `/admin/after-sales?page=${page.value}`,
    );
    list.value = p.items;
    total.value = p.total;
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
function edit(s: AfterSale) {
  selected.value = s;
  show.value = true;
  status.value = "approved";
  note.value = "";
  error.value = "";
}
async function save() {
  busy.value = true;
  try {
    await api.request(`/admin/after-sales/${selected.value?.id}`, "PATCH", {
      status: status.value,
      note: note.value,
    });
    show.value = false;
    message.success("售后已处理");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
const columns: DataTableColumns<AfterSale> = [
  { title: "申请 ID", key: "id" },
  { title: "订单", key: "order", render: (s) => s.order.number },
  { title: "金额", key: "total", render: (s) => money(s.order.total) },
  { title: "申请原因", key: "reason", minWidth: 200 },
  { title: "状态", key: "status", render: (s) => statusLabel(s.status) },
  { title: "处理备注", key: "note" },
  {
    title: "操作",
    key: "actions",
    render: (s) =>
      h(
        NButton,
        {
          size: "small",
          disabled: s.status !== "pending",
          onClick: () => edit(s),
        },
        () => "处理售后",
      ),
  },
];
onMounted(load);
</script>
<template>
  <div class="view-title">
    <div>
      <p class="eyebrow">AFTER-SALES SERVICE</p>
      <h1>售后管理</h1>
      <p>审核全单退货申请，批准后恢复库存并结束订单。</p>
    </div>
  </div>
  <NAlert v-if="error && !show" type="error">{{ error }}</NAlert
  ><NCard
    ><NDataTable
      :loading="loading"
      :data="list"
      :columns="columns"
      :scroll-x="1000" /><NPagination
      v-model:page="page"
      :item-count="total"
      :page-size="12"
      @update:page="load"
      style="margin-top: 20px" /></NCard
  ><NModal
    v-model:show="show"
    preset="card"
    title="处理售后"
    style="width: 520px"
    ><form v-if="selected" @submit.prevent="save">
      <p>订单：{{ selected.order.number }}</p>
      <p>原因：{{ selected.reason }}</p>
      <NAlert type="info"
        >批准表示已线下核实退货及资金处理，系统仅恢复库存和记录处理结果，不执行真实退款。</NAlert
      ><NAlert v-if="error" type="error">{{ error }}</NAlert
      ><label
        >处理结果<NSelect
          v-model:value="status"
          :options="[
            { label: '批准', value: 'approved' },
            { label: '拒绝', value: 'rejected' },
          ]" /></label
      ><label
        >处理备注<NInput
          v-model:value="note"
          type="textarea"
          placeholder="处理备注"
          :maxlength="1000" /></label
      ><NButton
        attr-type="submit"
        type="primary"
        :loading="busy"
        style="margin-top: 20px"
        >提交处理结果</NButton
      >
    </form></NModal
  >
</template>
