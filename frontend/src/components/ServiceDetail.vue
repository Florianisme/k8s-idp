<!-- frontend/src/components/ServiceDetail.vue -->
<template>
  <v-container fluid>
    <div class="d-flex align-center ga-2 mb-1">
      <h1 class="text-h5">{{ service.name }}</h1>
      <v-chip v-if="service.hasSpec" color="success" size="small" variant="tonal">OpenAPI</v-chip>
    </div>

    <p v-if="service.description" class="text-body-2 text-medium-emphasis mb-3">
      {{ service.description }}
    </p>

    <div class="d-flex flex-wrap ga-2 mb-4">
      <v-chip v-if="service.owner" prepend-icon="mdi-account-group" variant="outlined" size="small">
        {{ service.owner }}
      </v-chip>
      <v-chip v-if="service.namespace" prepend-icon="mdi-kubernetes" variant="outlined" size="small">
        {{ service.namespace }}
      </v-chip>
      <v-btn
        v-if="service.sourceUrl"
        :href="service.sourceUrl"
        target="_blank"
        prepend-icon="mdi-source-repository"
        variant="tonal"
        size="small"
        color="primary"
      >Source Code</v-btn>
    </div>

    <v-divider class="mb-4" />

    <SwaggerViewer v-if="service.hasSpec" :service-id="service.id" />
    <v-alert v-else type="info" variant="tonal">
      No OpenAPI spec configured. Add the <code>k8s-idp/openapi-path</code> label to enable API docs.
    </v-alert>
  </v-container>
</template>

<script setup>
import SwaggerViewer from './SwaggerViewer.vue'
defineProps({ service: { type: Object, required: true } })
</script>
