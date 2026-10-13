<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { NCard, NAlert, NSpin, NButton } from "naive-ui";
import type { EChartsOption } from "echarts";
import type { Dashboard } from "../../../shared/types";
import { money } from "../../../shared/helpers";
import { api } from "../api";
import Chart from "../components/Chart.vue";
const data = ref<Dashboard | null>(null),
  error = ref(""),
  loading = ref(false);
async function load() {
  loading.value = true;
  error.value = "";
  try {
    data.value = await api.request<Dashboard>("/admin/dashboard");
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}
const sales = computed<EChartsOption>(() => ({
  color: ["#398265", "#c3a278"],
  tooltip: { trigger: "axis" },
  legend: { top: 0, data: ["销售额（元）", "订单数"] },
  grid: { left: 65, right: 40, top: 45, bottom: 35 },
  xAxis: {
    type: "category",
    data: data.value?.trend.map((x) => x.date.slice(0, 10)),
  },
  yAxis: [{ type: "value" }, { type: "value" }],
  series: [
    {
      name: "销售额（元）",
      type: "line",
      smooth: true,
      areaStyle: { opacity: 0.1 },
      data: data.value?.trend.map((x) => x.sales / 100),
    },
    {
      name: "订单数",
      type: "bar",
      yAxisIndex: 1,
      barMaxWidth: 18,
      data: data.value?.trend.map((x) => x.orders),
    },
  ],
}));
const growth = computed<EChartsOption>(() => ({
  color: ["#608bb4"],
  tooltip: { trigger: "axis" },
  grid: { left: 40, right: 20, top: 25, bottom: 35 },
  xAxis: {
    type: "category",
    data: data.value?.growth.map((x) => x.date.slice(0, 10)),
  },
  yAxis: { type: "value", minInterval: 1 },
  series: [
    {
      name: "新增用户",
      type: "line",
      smooth: true,
      areaStyle: { opacity: 0.08 },
      data: data.value?.growth.map((x) => x.users),
    },
  ],
}));
const ranking = computed<EChartsOption>(() => ({
  color: ["#528569"],
  tooltip: { trigger: "axis" },
  grid: { left: 130, right: 25, top: 20, bottom: 30 },
  xAxis: { type: "value", minInterval: 1 },
  yAxis: {
    type: "category",
    inverse: true,
    data: data.value?.ranking.map((x) => x.name),
  },
  series: [
    {
      name: "销量",
      type: "bar",
      barMaxWidth: 22,
      data: data.value?.ranking.map((x) => x.quantity),
    },
  ],
}));
const hot = computed<EChartsOption>(() => ({
  color: ["#3f765d", "#7c9c71", "#bcc6a1", "#d9b98b", "#afc5c1"],
  tooltip: { trigger: "item", formatter: "{b}: {c} 元 ({d}%)" },
  legend: { bottom: 0 },
  series: [
    {
      name: "商品销售额",
      type: "pie",
      radius: ["40%", "65%"],
      center: ["50%", "44%"],
      data: data.value?.ranking.map((x) => ({
        name: x.name,
        value: x.sales / 100,
      })),
    },
  ],
}));
onMounted(load);
</script>
<template>
  <div class="view-title">
    <div>
      <p class="eyebrow">BUSINESS OVERVIEW</p>
      <h1>经营概览</h1>
      <p>每一次增长，都有数据可循。</p>
    </div>
    <NButton @click="load" :loading="loading">刷新数据</NButton>
  </div>
  <NAlert v-if="error" type="error">{{ error }}</NAlert
  ><NSpin :show="loading"
    ><template v-if="data"
      ><div class="metric-grid">
        <NCard
          v-for="m in [
            ['累计销售额', money(data.sales), '已确认、发货与完成订单'],
            ['订单总数', data.orders, '全部订单'],
            ['注册用户', data.users, '当前累计用户'],
            ['商品总数', data.products, '包含已下架商品'],
          ]"
          :key="String(m[0])"
          ><span class="metric-label">{{ m[0] }}</span
          ><strong class="metric-value">{{ m[1] }}</strong
          ><span class="metric-note">{{ m[2] }}</span></NCard
        >
      </div>
      <div class="chart-grid">
        <NCard title="销售额与订单趋势 · 最近 30 天"
          ><Chart :option="sales" /></NCard
        ><NCard title="用户增长趋势 · 最近 30 天"
          ><Chart :option="growth" /></NCard
        ><NCard title="商品销量排行"
          ><Chart :option="ranking" />
          <p v-if="!data.ranking.length" class="muted">
            确认订单后将展示实际销量。
          </p></NCard
        ><NCard title="热销商品 · 销售额占比"
          ><Chart :option="hot"
        /></NCard></div></template
  ></NSpin>
</template>
