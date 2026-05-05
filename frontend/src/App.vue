<!-- frontend/src/App.vue -->
<template>
  <v-app>
    <v-navigation-drawer permanent width="300">
      <v-toolbar color="primary" density="compact" flat>
        <v-toolbar-title class="text-body-1 font-weight-bold">Developer Portal</v-toolbar-title>
        <v-btn icon="mdi-refresh" size="small" :loading="loading" @click="fetchServices" />
      </v-toolbar>
      <ServiceList :services="services" :selected-id="selectedId" @select="selectedId = $event" />
    </v-navigation-drawer>

    <v-main>
      <ServiceDetail v-if="selectedService" :service="selectedService" />
      <v-container v-else class="d-flex align-center justify-center fill-height">
        <div class="text-center text-medium-emphasis">
          <v-icon size="64" class="mb-4">mdi-api</v-icon>
          <div>Select a service to view its documentation</div>
        </div>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import ServiceList from './components/ServiceList.vue'
import ServiceDetail from './components/ServiceDetail.vue'

const services = ref([])
const selectedId = ref(null)
const loading = ref(false)
let timer = null

const selectedService = computed(() =>
  services.value.find(s => s.id === selectedId.value) ?? null
)

async function fetchServices() {
  loading.value = true
  try {
    const res = await fetch('/api/services')
    services.value = await res.json()
    if (selectedId.value && !services.value.find(s => s.id === selectedId.value)) {
      selectedId.value = null
    }
  } catch (e) {
    console.error('Failed to fetch services:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => { fetchServices(); timer = setInterval(fetchServices, 30_000) })
onUnmounted(() => clearInterval(timer))
</script>
