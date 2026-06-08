<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Security"
      title="Audit Logs"
      subtitle="View security events and user activity"
    />

    <UCard>
      <div class="flex flex-wrap gap-2 mb-4">
        <UInput v-model="filters.namespace" placeholder="Namespace" class="w-40" @input="debouncedFetch" />
        <UInput v-model="filters.user_id" placeholder="User ID" class="w-48" @input="debouncedFetch" />
        <UInput v-model="filters.event" placeholder="Event type" class="w-48" @input="debouncedFetch" />
        <UInput v-model="filters.from" type="datetime-local" class="w-52" @change="fetchLogs" />
        <UInput v-model="filters.to" type="datetime-local" class="w-52" @change="fetchLogs" />
      </div>

      <UTable :rows="logs" :columns="columns" :loading="loading">
        <template #event-data="{ row }">
          <UBadge :color="eventColor(row.event)" variant="soft" size="xs">{{ row.event }}</UBadge>
        </template>
        <template #timestamp-data="{ row }">
          <span class="text-xs font-mono">{{ formatTime(row.timestamp) }}</span>
        </template>
        <template #metadata_json-data="{ row }">
          <div class="max-w-[280px]">
            <JsonView v-if="row.metadata_json" :data="row.metadata_json" label="metadata" />
            <span v-else class="text-xs text-slate-500">—</span>
          </div>
        </template>
      </UTable>

      <div class="flex justify-between items-center mt-4" v-if="totalPages > 1">
        <span class="text-sm text-slate-500">{{ totalCount }} events total</span>
        <UPagination v-model="page" :page-count="pageSize" :total="totalCount" @update:model-value="fetchLogs" />
      </div>
    </UCard>
  </div>
</template>

<script setup>
import { useApi } from '~/composables/useApi'

definePageMeta({ middleware: 'auth' })

const api = useApi()
const toast = useToast()

const loading = ref(true)
const logs = ref([])
const page = ref(1)
const pageSize = 50
const totalCount = ref(0)
const totalPages = ref(0)
const filters = reactive({ namespace: 'default', user_id: '', event: '', from: '', to: '' })

const columns = [
  { key: 'event', label: 'Event' },
  { key: 'user_id', label: 'User ID' },
  { key: 'namespace', label: 'Namespace' },
  { key: 'ip', label: 'IP' },
  { key: 'user_agent', label: 'User Agent' },
  { key: 'metadata_json', label: 'Metadata' },
  { key: 'timestamp', label: 'Time' },
]

function eventColor(event) {
  if (event?.includes('success')) return 'green'
  if (event?.includes('failed')) return 'red'
  if (event?.includes('mfa')) return 'orange'
  return 'gray'
}

function formatTime(ts) {
  if (!ts) return '-'
  return new Date(ts).toLocaleString()
}

let fetchTimeout = null
function debouncedFetch() { clearTimeout(fetchTimeout); fetchTimeout = setTimeout(() => { page.value = 1; fetchLogs() }, 300) }

async function fetchLogs() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page_size: String(pageSize), page: String(page.value) })
    if (filters.namespace) params.set('namespace', filters.namespace)
    if (filters.user_id) params.set('user_id', filters.user_id)
    if (filters.event) params.set('event', filters.event)
    if (filters.from) params.set('from', new Date(filters.from).toISOString())
    if (filters.to) params.set('to', new Date(filters.to).toISOString())
    const res = await api.fetch(`/v1/admin/audit?${params}`)
    logs.value = res.logs || []
    totalCount.value = res.total_count || 0
    totalPages.value = res.total_pages || 0
  } catch { toast.add({ title: 'Error fetching audit logs', color: 'red' }) }
  finally { loading.value = false }
}

onMounted(fetchLogs)
</script>
