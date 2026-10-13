<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
import * as echarts from "echarts/core";
import { LineChart, BarChart, PieChart } from "echarts/charts";
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
import type { EChartsOption } from "echarts";
echarts.use([
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer,
]);
const props = defineProps<{ option: EChartsOption }>();
const element = ref<HTMLDivElement>();
let chart: echarts.ECharts | undefined, observer: ResizeObserver | undefined;
onMounted(() => {
  if (element.value) {
    chart = echarts.init(element.value);
    chart.setOption(props.option);
    observer = new ResizeObserver(() => chart?.resize());
    observer.observe(element.value);
  }
});
watch(
  () => props.option,
  (option) => chart?.setOption(option, true),
  { deep: true },
);
onBeforeUnmount(() => {
  observer?.disconnect();
  chart?.dispose();
});
</script>
<template>
  <div ref="element" class="chart" role="img" aria-label="经营数据图表"></div>
</template>
