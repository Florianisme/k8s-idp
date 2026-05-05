<!-- frontend/src/components/ServiceList.vue -->
<template>
  <div>
    <v-text-field
      v-model="filter"
      prepend-inner-icon="mdi-magnify"
      placeholder="Filter services…"
      variant="outlined"
      density="compact"
      hide-details
      class="ma-2"
    />
    <v-list nav>
      <v-list-item
        v-for="svc in filtered"
        :key="svc.id"
        :active="svc.id === selectedId"
        active-color="primary"
        @click="$emit('select', svc.id)"
      >
        <v-list-item-title>{{ svc.name }}</v-list-item-title>
        <v-list-item-subtitle>{{ svc.owner || 'No owner' }}</v-list-item-subtitle>
        <template #append>
          <v-chip v-if="svc.hasSpec" size="x-small" color="success" variant="tonal">API</v-chip>
        </template>
      </v-list-item>
      <v-list-item v-if="filtered.length === 0" disabled>
        <v-list-item-title class="text-medium-emphasis">No services found</v-list-item-title>
      </v-list-item>
    </v-list>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  services: { type: Array, default: () => [] },
  selectedId: { type: String, default: null }
})
defineEmits(['select'])

const filter = ref('')
const filtered = computed(() => {
  const q = filter.value.toLowerCase()
  if (!q) return props.services
  return props.services.filter(s =>
    s.name.toLowerCase().includes(q) ||
    (s.owner || '').toLowerCase().includes(q) ||
    (s.description || '').toLowerCase().includes(q)
  )
})
</script>
