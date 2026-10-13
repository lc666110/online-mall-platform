<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import {
  NButton,
  NDataTable,
  NInput,
  NInputNumber,
  NSelect,
  NSwitch,
  NModal,
  NCard,
  NAlert,
  NPopconfirm,
  NSpace,
  NPagination,
  useMessage,
} from "naive-ui";
import type { DataTableColumns } from "naive-ui";
import type { Product, Category, Page } from "../../../shared/types";
import { money, images } from "../../../shared/helpers";
import { api } from "../api";
interface EditableVariant {
  id: number;
  sku: string;
  specification: string;
  price: number;
  stock: number;
  expected_stock?: number;
}
interface Form {
  id: number;
  name: string;
  category_id: number | null;
  description: string;
  images: string;
  active: boolean;
  variants: EditableVariant[];
}
const list = ref<Product[]>([]),
  categories = ref<Category[]>([]),
  query = ref(""),
  page = ref(1),
  total = ref(0),
  loading = ref(false),
  saving = ref(false),
  error = ref(""),
  show = ref(false),
  form = ref<Form>(blank());
const message = useMessage();
function blank(): Form {
  return {
    id: 0,
    name: "",
    category_id: null,
    description: "",
    images: "",
    active: true,
    variants: [
      { id: 0, sku: "", specification: "默认规格", price: 100, stock: 0 },
    ],
  };
}
async function load() {
  loading.value = true;
  error.value = "";
  try {
    const p = await api.request<Page<Product>>(
      `/admin/products?q=${encodeURIComponent(query.value)}&page=${page.value}`,
    );
    list.value = p.items;
    total.value = p.total;
    categories.value = await api.request<Category[]>("/admin/categories");
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
function edit(p?: Product) {
  form.value = p
    ? {
        id: p.id,
        name: p.name,
        category_id: p.category_id,
        description: p.description,
        images: images(p.images).join("\n"),
        active: p.active,
        variants: p.variants.map((v) => ({
          id: v.id,
          sku: v.sku,
          specification: v.specification,
          price: v.price,
          stock: v.stock,
          expected_stock: v.stock,
        })),
      }
    : blank();
  show.value = true;
}
async function save() {
  saving.value = true;
  error.value = "";
  try {
    if (
      !form.value.name ||
      !form.value.category_id ||
      form.value.variants.some((v) => !v.sku)
    )
      throw new Error("请填写商品名称、分类和所有 SKU");
    await api.request(
      form.value.id ? `/admin/products/${form.value.id}` : "/admin/products",
      form.value.id ? "PUT" : "POST",
      {
        ...form.value,
        images: form.value.images
          .split("\n")
          .map((x) => x.trim())
          .filter(Boolean),
      },
    );
    show.value = false;
    message.success("商品已保存");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    saving.value = false;
  }
}
async function remove(p: Product) {
  try {
    await api.request(`/admin/products/${p.id}`, "DELETE");
    message.success("商品已删除");
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  }
}
async function toggle(p: Product) {
  try {
    await api.request(`/admin/products/${p.id}/status`, "PATCH", {
      active: !p.active,
    });
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  }
}
const columns: DataTableColumns<Product> = [
  { title: "商品", key: "name", minWidth: 180 },
  { title: "分类", key: "category", render: (p) => p.category.name },
  {
    title: "价格",
    key: "price",
    render: (p) => money(Math.min(...p.variants.map((v) => v.price))),
  },
  {
    title: "库存",
    key: "stock",
    render: (p) => p.variants.reduce((s, v) => s + v.stock, 0),
  },
  {
    title: "状态",
    key: "active",
    render: (p) => (p.active ? "销售中" : "已下架"),
  },
  {
    title: "操作",
    key: "actions",
    width: 230,
    render: (p) =>
      h(NSpace, {}, () => [
        h(NButton, { size: "small", onClick: () => edit(p) }, () => "编辑"),
        h(NButton, { size: "small", onClick: () => toggle(p) }, () =>
          p.active ? "下架" : "上架",
        ),
        h(
          NPopconfirm,
          { onPositiveClick: () => remove(p) },
          {
            trigger: () =>
              h(
                NButton,
                { size: "small", type: "error", quaternary: true },
                () => "删除",
              ),
            default: () => "确认删除此商品？有交易记录的商品须下架保留。",
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
      <p class="eyebrow">PRODUCT CATALOG</p>
      <h1>商品管理</h1>
      <p>维护商品、规格、价格与库存。</p>
    </div>
    <NButton type="primary" @click="edit()">新增商品</NButton>
  </div>
  <NAlert v-if="error && !show" type="error" style="margin-bottom: 20px">{{
    error
  }}</NAlert
  ><NCard
    ><form
      class="toolbar"
      @submit.prevent="
        page = 1;
        load();
      "
    >
      <NInput
        v-model:value="query"
        placeholder="搜索商品名称"
        style="max-width: 320px"
      /><NButton attr-type="submit">搜索</NButton>
    </form>
    <NDataTable
      :columns="columns"
      :data="list"
      :loading="loading"
      :scroll-x="950"
      :row-key="(p: Product) => p.id" /><NPagination
      v-model:page="page"
      :item-count="total"
      :page-size="12"
      @update:page="load"
      style="margin-top: 20px" /></NCard
  ><NModal
    v-model:show="show"
    preset="card"
    :title="form.id ? '编辑商品' : '新增商品'"
    style="width: min(900px, 95vw)"
    ><form @submit.prevent="save">
      <NAlert v-if="error" type="error">{{ error }}</NAlert>
      <div class="form-grid">
        <label
          >商品名称<NInput
            v-model:value="form.name"
            placeholder="商品名称"
            :maxlength="160" /></label
        ><label
          >分类<NSelect
            v-model:value="form.category_id"
            :options="categories.map((c) => ({ label: c.name, value: c.id }))"
            placeholder="选择分类" /></label
        ><label>上架状态<NSwitch v-model:value="form.active" /></label>
      </div>
      <label
        >商品描述<NInput
          v-model:value="form.description"
          type="textarea"
          placeholder="商品描述"
          :maxlength="20000" /></label
      ><label
        >图片地址（每行一个 HTTPS 或 /images/ 地址）<NInput
          v-model:value="form.images"
          type="textarea"
          placeholder="图片地址"
      /></label>
      <h3>规格与库存 · 价格单位为分</h3>
      <div v-for="(v, i) in form.variants" :key="i" class="variant-editor">
        <label>SKU<NInput v-model:value="v.sku" placeholder="SKU" /></label
        ><label
          >规格<NInput
            v-model:value="v.specification"
            placeholder="规格" /></label
        ><label
          >价格（分）<NInputNumber
            v-model:value="v.price"
            :min="1"
            :max="100000000"
            :precision="0" /></label
        ><label
          >库存<NInputNumber
            v-model:value="v.stock"
            :min="0"
            :max="1000000"
            :precision="0" /></label
        ><NButton
          v-if="!v.id && form.variants.length > 1"
          @click="form.variants.splice(i, 1)"
          >移除</NButton
        >
      </div>
      <NButton
        @click="
          form.variants.push({
            id: 0,
            sku: '',
            specification: '',
            price: 100,
            stock: 0,
          })
        "
        :disabled="form.variants.length >= 50"
        >添加规格</NButton
      >
      <p class="muted">
        已有规格保留以维护历史订单；库存修改为指定的绝对值；若销售改变库存，须刷新后重新调整。
      </p>
      <NSpace justify="end"
        ><NButton @click="show = false">取消</NButton
        ><NButton attr-type="submit" type="primary" :loading="saving"
          >保存商品</NButton
        ></NSpace
      >
    </form></NModal
  >
</template>
