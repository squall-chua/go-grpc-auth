<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Roles</h1>
        <p class="text-sm text-slate-500">Manage RBAC role definitions</p>
      </div>
      <UButton icon="i-heroicons-plus" @click="openCreateModal">Create Role</UButton>
    </header>

    <UCard>
      <div class="flex items-center gap-2 mb-4">
        <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="Search by name or namespace..." class="w-80" @input="debouncedFetch" />
      </div>

      <UTable :rows="roles" :columns="columns" :loading="loading">
        <template #permissions-data="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge v-for="p in row.permissions" :key="p" color="gray" variant="soft" size="xs">{{ p }}</UBadge>
            <span v-if="!row.permissions?.length" class="text-sm text-slate-500">None</span>
          </div>
        </template>
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
          <h3 class="text-base font-semibold">New Role</h3>
        </template>
        <form @submit.prevent="handleCreate" class="space-y-4">
          <UFormGroup label="Name">
            <UInput v-model="form.name" placeholder="e.g. editor" />
          </UFormGroup>
          <UFormGroup label="Namespace">
            <UInput v-model="form.namespace" placeholder="default" />
          </UFormGroup>
          <UFormGroup label="Permissions (comma-separated)">
            <UInput v-model="permissionsInput" placeholder="users:read, users:write" />
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
const roles = ref([])
const search = ref('')
const permissionsInput = ref('')
const form = reactive({ name: '', namespace: 'default', permissions: [] })

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'namespace', label: 'Namespace' },
  { key: 'permissions', label: 'Permissions' },
  { key: 'actions', label: '' }
]

let fetchTimeout = null
function debouncedFetch() { clearTimeout(fetchTimeout); fetchTimeout = setTimeout(fetchRoles, 300) }

function openCreateModal() {
  form.name = ''
  form.namespace = 'default'
  permissionsInput.value = ''
  isCreateOpen.value = true
}

async function fetchRoles() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page_size: '100' })
    if (search.value) params.set('query', search.value)
    const res = await api.fetch(`/v1/admin/roles?${params}`)
    roles.value = res.roles || []
  } catch { toast.add({ title: 'Error fetching roles', color: 'red' }) }
  finally { loading.value = false }
}

async function handleCreate() {
  saving.value = true
  form.permissions = permissionsInput.value.split(',').map(s => s.trim()).filter(Boolean)
  try {
    await api.fetch('/v1/admin/roles', { method: 'POST', body: form })
    toast.add({ title: 'Role created', color: 'green' })
    isCreateOpen.value = false
    fetchRoles()
  } catch { toast.add({ title: 'Creation failed', color: 'red' }) }
  finally { saving.value = false }
}

async function handleDelete(id) {
  if (!confirm('Delete this role?')) return
  try {
    await api.fetch(`/v1/admin/roles/${id}`, { method: 'DELETE' })
    toast.add({ title: 'Role deleted', color: 'green' })
    fetchRoles()
  } catch { toast.add({ title: 'Deletion failed', color: 'red' }) }
}

onMounted(fetchRoles)
</script>
