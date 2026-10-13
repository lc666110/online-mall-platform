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
  NSwitch,
  NPagination,
  useMessage,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";
import type { User, Page } from "../../../shared/types";
import { api } from "../api";
const list = ref<User[]>([]),
  form = ref<User | null>(null),
  page = ref(1),
  total = ref(0),
  show = ref(false),
  busy = ref(false),
  loading = ref(false),
  error = ref("");
const message = useMessage();
async function load() {
  loading.value = true;
  error.value = "";
  try {
    const p = await api.request<Page<User>>(`/admin/users?page=${page.value}`);
    list.value = p.items;
    total.value = p.total;
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
function edit(u: User) {
  form.value = { ...u };
  show.value = true;
  error.value = "";
}
async function save() {
  busy.value = true;
  try {
    await api.request(`/admin/users/${form.value?.id}`, "PATCH", form.value);
    show.value = false;
    message.success("用户资料与权限已保存");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    busy.value = false;
  }
}
const columns: DataTableColumns<User> = [
  { title: "ID", key: "id" },
  { title: "昵称", key: "name" },
  { title: "邮箱", key: "email", minWidth: 200 },
  { title: "角色", key: "role" },
  {
    title: "状态",
    key: "enabled",
    render: (u) => (u.enabled ? "启用" : "禁用"),
  },
  {
    title: "操作",
    key: "actions",
    render: (u) =>
      h(NButton, { size: "small", onClick: () => edit(u) }, () => "管理用户"),
  },
];
onMounted(load);
</script>
<template>
  <div class="view-title">
    <div>
      <p class="eyebrow">CUSTOMERS & ACCESS</p>
      <h1>用户与权限</h1>
      <p>仅管理员可访问，运营人员不能管理用户权限。</p>
    </div>
  </div>
  <NAlert v-if="error && !show" type="error">{{ error }}</NAlert
  ><NCard
    ><NDataTable
      :loading="loading"
      :data="list"
      :columns="columns"
      :scroll-x="850" /><NPagination
      v-model:page="page"
      :item-count="total"
      :page-size="12"
      @update:page="load"
      style="margin-top: 20px" /></NCard
  ><NModal
    v-model:show="show"
    preset="card"
    title="用户资料与权限"
    style="width: 500px"
    ><form v-if="form" @submit.prevent="save">
      <p>{{ form.email }}</p>
      <NAlert v-if="error" type="error">{{ error }}</NAlert
      ><label
        >昵称<NInput
          v-model:value="form.name"
          placeholder="昵称"
          :maxlength="80" /></label
      ><label
        >电话<NInput
          v-model:value="form.phone"
          placeholder="电话"
          :maxlength="30" /></label
      ><label
        >角色<NSelect
          v-model:value="form.role"
          :options="[
            { label: '普通用户', value: 'user' },
            { label: '运营人员', value: 'staff' },
            { label: '管理员', value: 'admin' },
          ]" /></label
      ><label>账号启用<NSwitch v-model:value="form.enabled" /></label
      ><NButton
        attr-type="submit"
        type="primary"
        :loading="busy"
        style="margin-top: 20px"
        >保存用户</NButton
      >
    </form></NModal
  >
</template>
