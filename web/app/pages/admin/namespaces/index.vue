<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Namespaces</h1>
        <p class="text-sm text-slate-500">Manage tenant isolation and security policies</p>
      </div>
      <UButton icon="i-heroicons-plus" @click="openCreateModal">Create Namespace</UButton>
    </header>

    <UCard>
      <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="Search namespaces..." class="max-w-xs mb-4" @input="debouncedFetch" />

      <UTable :rows="namespaces" :columns="columns" :loading="loading">
        <template #config-data="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge size="xs" :color="row.config?.mfa_required ? 'orange' : 'gray'" variant="soft">MFA {{ row.config?.mfa_required ? 'ON' : 'OFF' }}</UBadge>
            <template v-if="row.config?.password_policy">
              <UBadge size="xs" color="primary" variant="soft">min {{ row.config.password_policy.min_length || 0 }} chars</UBadge>
              <UBadge v-if="row.config.password_policy.require_uppercase" size="xs" color="primary" variant="soft">A-Z</UBadge>
              <UBadge v-if="row.config.password_policy.require_lowercase" size="xs" color="primary" variant="soft">a-z</UBadge>
              <UBadge v-if="row.config.password_policy.require_number" size="xs" color="primary" variant="soft">0-9</UBadge>
              <UBadge v-if="row.config.password_policy.require_special" size="xs" color="primary" variant="soft">!@#</UBadge>
              <UBadge v-if="row.config.password_policy.password_history" size="xs" color="primary" variant="soft">history {{ row.config.password_policy.password_history }}</UBadge>
            </template>
            <UBadge v-for="p in row.config?.allowed_social_providers" :key="p" size="xs" color="blue" variant="soft">{{ p }}</UBadge>
            <UBadge v-if="row.config?.ip_allowlist?.length" size="xs" color="green" variant="soft">IP allowlist</UBadge>
            <UBadge v-if="row.config?.ip_denylist?.length" size="xs" color="red" variant="soft">IP denylist</UBadge>
            <UBadge v-if="row.config?.webhook_url" size="xs" color="purple" variant="soft">Webhook</UBadge>
          </div>
        </template>
        <template #actions-data="{ row }">
          <div class="flex gap-1">
            <UTooltip text="Edit Config">
              <UButton size="xs" variant="ghost" icon="i-heroicons-pencil-square" @click="openEditModal(row)" />
            </UTooltip>
            <UTooltip text="Delete">
              <UButton size="xs" variant="ghost" color="red" icon="i-heroicons-trash" @click="handleDelete(row.id)" />
            </UTooltip>
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Create/Edit Modal -->
    <UModal v-model="isModalOpen">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold">{{ editingId ? 'Edit Namespace Config' : 'New Namespace' }}</h3>
            <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="isModalOpen = false" />
          </div>
        </template>

        <form @submit.prevent="editingId ? handleUpdateConfig() : handleCreate()" class="space-y-4 py-2">
          <UFormGroup label="Name" v-if="!editingId">
            <UInput v-model="form.name" placeholder="e.g. acme-corp" />
          </UFormGroup>
          <UFormGroup label="Security">
            <UCheckbox v-model="form.config.mfa_required" label="Require MFA for all users" />
          </UFormGroup>
          <UFormGroup label="Social Providers">
            <USelectMenu v-model="form.config.allowed_social_providers" multiple :options="['google', 'github']" />
          </UFormGroup>
          <UFormGroup label="Password Policy">
            <div class="space-y-2">
              <UInput v-model.number="form.config.password_policy.min_length" type="number" placeholder="Min length" />
              <UCheckbox v-model="form.config.password_policy.require_uppercase" label="Require uppercase" />
              <UCheckbox v-model="form.config.password_policy.require_lowercase" label="Require lowercase" />
              <UCheckbox v-model="form.config.password_policy.require_number" label="Require number" />
              <UCheckbox v-model="form.config.password_policy.require_special" label="Require special character" />
              <UInput v-model.number="form.config.password_policy.password_history" type="number" placeholder="Password history count" />
            </div>
          </UFormGroup>
          <div class="flex justify-end gap-2 mt-4">
            <UButton variant="ghost" @click="isModalOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="saving">{{ editingId ? 'Save' : 'Create' }}</UButton>
          </div>
        </form>
      </UCard>
    </UModal>
  </div>
</template>

<script setup>
import { useApi } from '~/composables/useApi'

definePageMeta({ middleware: 'auth' })

const api = useApi()
const toast = useToast()

const loading = ref(true)
const saving = ref(false)
const isModalOpen = ref(false)
const namespaces = ref([])
const editingId = ref(null)
const search = ref('')

const defaultConfig = () => ({
  mfa_required: false,
  allowed_social_providers: [],
  password_policy: { min_length: 8, require_uppercase: false, require_lowercase: false, require_number: false, require_special: false, password_history: 0 },
})

const form = reactive({ name: '', config: defaultConfig() })

const columns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'id', label: 'ID' },
  { key: 'config', label: 'Security Policy' },
  { key: 'actions', label: '' }
]

let fetchTimeout = null
function debouncedFetch() {
  clearTimeout(fetchTimeout)
  fetchTimeout = setTimeout(fetchNamespaces, 300)
}

function openCreateModal() {
  editingId.value = null
  form.name = ''
  Object.assign(form.config, defaultConfig())
  isModalOpen.value = true
}

function openEditModal(row) {
  editingId.value = row.id
  form.name = row.name
  Object.assign(form.config, defaultConfig(), row.config || {})
  if (row.config?.password_policy) Object.assign(form.config.password_policy, row.config.password_policy)
  isModalOpen.value = true
}

async function fetchNamespaces() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page_size: '100' })
    if (search.value) params.set('query', search.value)
    const res = await api.fetch(`/v1/admin/namespaces?${params}`)
    namespaces.value = res.namespaces || []
  } catch { toast.add({ title: 'Error fetching namespaces', color: 'red' }) }
  finally { loading.value = false }
}

async function handleCreate() {
  saving.value = true
  try {
    await api.fetch('/v1/admin/namespaces', { method: 'POST', body: form })
    toast.add({ title: 'Namespace created', color: 'green' })
    isModalOpen.value = false
    fetchNamespaces()
  } catch { toast.add({ title: 'Creation failed', color: 'red' }) }
  finally { saving.value = false }
}

async function handleUpdateConfig() {
  saving.value = true
  try {
    await api.fetch(`/v1/admin/namespaces/${editingId.value}/config`, { method: 'PATCH', body: { config: form.config } })
    toast.add({ title: 'Config updated', color: 'green' })
    isModalOpen.value = false
    fetchNamespaces()
  } catch { toast.add({ title: 'Update failed', color: 'red' }) }
  finally { saving.value = false }
}

async function handleDelete(id) {
  if (!confirm('Are you sure? This action is permanent.')) return
  try {
    await api.fetch(`/v1/admin/namespaces/${id}`, { method: 'DELETE' })
    toast.add({ title: 'Namespace deleted', color: 'green' })
    fetchNamespaces()
  } catch { toast.add({ title: 'Deletion failed', color: 'red' }) }
}

onMounted(fetchNamespaces)
</script>
