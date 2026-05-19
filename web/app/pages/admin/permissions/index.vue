<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Permissions</h1>
        <p class="text-sm text-slate-500">Manage permission definitions</p>
      </div>
      <UButton icon="i-heroicons-plus" @click="openCreateModal">Create Permission</UButton>
    </header>

    <UCard>
      <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="Search by name or namespace..." class="w-80 mb-4" @input="debouncedFetch" />

      <UTable :rows="permissions" :columns="columns" :loading="loading">
        <template #actions-data="{ row }">
          <UTooltip text="Delete">
            <UButton size="xs" variant="ghost" color="red" icon="i-heroicons-trash" @click="handleDelete(row.id)" />
          </UTooltip>
        </template>
      </UTable>
    </UCard>

    <UModal v-model="isCreateOpen">
      <UCard>
        <template #header>
          <h3 class="text-base font-semibold">New Permission</h3>
        </template>
        <form @submit.prevent="handleCreate" class="space-y-4">
          <UFormGroup label="Name">
            <UInput v-model="form.name" placeholder="e.g. users:read" />
          </UFormGroup>
          <UFormGroup label="Namespace">
            <UInput v-model="form.namespace" placeholder="default" />
          </UFormGroup>
          <UFormGroup label="Description">
            <UInput v-model="form.description" placeholder="Read user data" />
          </UFormGroup>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="isCreateOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="saving">Create</UButton>
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
const isCreateOpen = ref(false)
const permissions = ref([])
const search = ref('')
const form = reactive({ name: '', namespace: 'default', description: '' })

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'namespace', label: 'Namespace' },
  { key: 'description', label: 'Description' },
  { key: 'actions', label: '' }
]

let fetchTimeout = null
function debouncedFetch() { clearTimeout(fetchTimeout); fetchTimeout = setTimeout(fetchPermissions, 300) }

function openCreateModal() {
  form.name = ''
  form.namespace = 'default'
  form.description = ''
  isCreateOpen.value = true
}

async function fetchPermissions() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page_size: '100' })
    if (search.value) params.set('query', search.value)
    const res = await api.fetch(`/v1/admin/permissions?${params}`)
    permissions.value = res.permissions || []
  } catch { toast.add({ title: 'Error fetching permissions', color: 'red' }) }
  finally { loading.value = false }
}

async function handleCreate() {
  saving.value = true
  try {
    await api.fetch('/v1/admin/permissions', { method: 'POST', body: form })
    toast.add({ title: 'Permission created', color: 'green' })
    isCreateOpen.value = false
    fetchPermissions()
  } catch { toast.add({ title: 'Creation failed', color: 'red' }) }
  finally { saving.value = false }
}

async function handleDelete(id) {
  if (!confirm('Delete this permission?')) return
  try {
    await api.fetch(`/v1/admin/permissions/${id}`, { method: 'DELETE' })
    toast.add({ title: 'Permission deleted', color: 'green' })
    fetchPermissions()
  } catch { toast.add({ title: 'Deletion failed', color: 'red' }) }
}

onMounted(fetchPermissions)
</script>
