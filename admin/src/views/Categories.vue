<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import {
  NCard,
  NInput,
  NButton,
  NDataTable,
  NAlert,
  NModal,
  NPopconfirm,
  NSpace,
  useMessage,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";
import type { Category } from "../../../shared/types";
import { api } from "../api";
const list = ref<Category[]>([]),
  id = ref(0),
  name = ref(""),
  show = ref(false),
  busy = ref(false),
  loading = ref(false),
  error = ref("");
const message = useMessage();
async function load() {
  loading.value = true;
  error.value = "";
  try {
    list.value = await api.request<Category[]>("/admin/categories");
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
function edit(c?: Category) {
  id.value = c?.id ?? 0;
  name.value = c?.name ?? "";
  show.value = true;
  error.value = "";
}
async function save() {
  busy.value = true;
  try {
    await api.request(
      id.value ? `/admin/categories/${id.value}` : "/admin/categories",
      id.value ? "PUT" : "POST",
      { name: name.value },
    );
    show.value = false;
    message.success("分类已保存");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
async function remove(c: Category) {
  try {
    await api.request(`/admin/categories/${c.id}`, "DELETE");
    message.success("分类已删除");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  }
}
const columns: DataTableColumns<Category> = [
  { title: "ID", key: "id" },
  { title: "分类名称", key: "name" },
  {
    title: "操作",
    key: "actions",
    render: (c) =>
      h(NSpace, {}, () => [
        h(NButton, { size: "small", onClick: () => edit(c) }, () => "编辑"),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(c) },
          {
            trigger: () =>
              h(NButton, { size: "small", type: "error" }, () => "删除"),
            default: () => "确认删除空分类？",
          },
        ),
      ]),
  },
];
onMounted(load);
</script>
<template>
  <div class="view-title">
    <div>
      <p class="eyebrow">CATEGORIES</p>
      <h1>分类管理</h1>
    </div>
    <NButton type="primary" @click="edit()">新增分类</NButton>
  </div>
  <NAlert v-if="error && !show" type="error">{{ error }}</NAlert
  ><NCard
    ><NDataTable :loading="loading" :data="list" :columns="columns" /></NCard
  ><NModal
    v-model:show="show"
    preset="card"
    title="商品分类"
    style="width: 450px"
    ><form @submit.prevent="save">
      <NAlert v-if="error" type="error">{{ error }}</NAlert
      ><label
        >分类名称<NInput
          v-model:value="name"
          placeholder="分类名称"
          :maxlength="80" /></label
      ><NButton
        type="primary"
        attr-type="submit"
        :loading="busy"
        style="margin-top: 20px"
        >保存分类</NButton
      >
    </form></NModal
  >
</template>
