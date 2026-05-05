<!-- frontend/src/components/SwaggerViewer.vue -->
<template>
  <div ref="container" />
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import SwaggerUI from 'swagger-ui-dist/swagger-ui-bundle'
import 'swagger-ui-dist/swagger-ui.css'

const props = defineProps({
  serviceId: { type: String, required: true }
})

const container = ref(null)
let ui = null

function mount() {
  if (container.value) container.value.innerHTML = ''
  ui = SwaggerUI({
    domNode: container.value,
    url: `/api/services/${encodeURIComponent(props.serviceId)}/spec`,
    presets: [SwaggerUI.presets.apis],
    layout: 'BaseLayout',
    tryItOutEnabled: true,
    deepLinking: false,
    defaultModelsExpandDepth: -1,
  })
}

onMounted(mount)
watch(() => props.serviceId, mount)
onBeforeUnmount(() => { ui = null })
</script>
