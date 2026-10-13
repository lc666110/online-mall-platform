<script setup lang="ts">
import type { Product } from "../../../shared/types";
import { images, money } from "../../../shared/helpers";
defineProps<{ product: Product }>();
</script>
<template>
  <RouterLink :to="`/products/${product.id}`" class="product-card"
    ><div class="product-image">
      <img
        v-if="images(product.images)[0]"
        :src="images(product.images)[0]"
        :alt="product.name"
        loading="lazy"
      /><span v-else class="image-placeholder">MORI</span
      ><span class="category-badge">{{ product.category.name }}</span>
    </div>
    <div class="product-copy">
      <h3>{{ product.name }}</h3>
      <p>{{ product.variants[0]?.specification }}</p>
      <strong>{{
        money(Math.min(...product.variants.map((v) => v.price)))
      }}</strong
      ><span class="arrow">↗</span>
    </div></RouterLink
  >
</template>
