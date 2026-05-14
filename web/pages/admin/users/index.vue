<template>
  <div class="space-y-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">Users</h1>
        <p class="text-sm text-slate-500">Manage user identities and access control</p>
      </div>
    </header>

    <UCard>
      <div class="flex items-center justify-between mb-4">
        <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="Search users..." class="max-w-xs" />
        <USelectMenu v-model="selectedNamespace" :options="namespaces" placeholder="All Namespaces" class="w-48" />
      </div>

      <UTable :rows="filteredUsers" :columns="columns" :loading="loading">
        <template #username-data="{ row }">
          <div class="flex items-center gap-2">
            <UAvatar :alt="row.username" size="xs" />
            <span class="font-medium text-slate-900 dark:text-white">{{ row.username }}</span>
          </div>
        </template>
        
        <template #status-data="{ row }">
          <UBadge :color="row.is_banned ? 'red' : 'green'" variant="soft" size="xs">
            {{ row.is_banned ? 'Banned' : 'Active' }}
          </UBadge>
        </template>

        <template #actions-data="{ row }">
          <UDropdown :items="actions(row)">
            <UButton color="gray" variant="ghost" icon="i-heroicons-ellipsis-horizontal-20-solid" />
          </UDropdown>
        </template>
      </UTable>
    </UCard>
  </div>
</template>

<script setup>
import { useApi } from '~/composables/useApi'

definePageMeta({
  middleware: 'auth'
})

const api = useApi()
const toast = useToast()

const loading = ref(true)
const search = ref('')
const selectedNamespace = ref(null)
const users = ref([])
const namespaces = ref(['default'])

const columns = [
  { key: 'username', label: 'User' },
  { key: 'email', label: 'Email' },
  { key: 'namespace', label: 'Namespace' },
  { key: 'status', label: 'Status' },
  { key: 'actions', label: '' }
]

const filteredUsers = computed(() => {
  return users.value.filter(u => {
    const matchesSearch = u.username.toLowerCase().includes(search.value.toLowerCase()) || 
                         u.email.toLowerCase().includes(search.value.toLowerCase())
    const matchesNamespace = !selectedNamespace.value || u.namespace === selectedNamespace.value
    return matchesSearch && matchesNamespace
  })
})

const actions = (row) => [
  [{
    label: 'View Details',
    icon: 'i-heroicons-eye',
    click: () => console.log('View', row.id)
  }],
  [{
    label: row.is_banned ? 'Unban User' : 'Ban User',
    icon: row.is_banned ? 'i-heroicons-check-circle' : 'i-heroicons-no-symbol',
    color: row.is_banned ? 'green' : 'red',
    click: () => handleToggleBan(row)
  }]
]

async function fetchUsers() {
  loading.value = true
  try {
    const res = await api.fetch('/v1/admin/users?page_size=100')
    users.value = res.users || []
  } catch (err) {
    toast.add({ title: 'Error fetching users', color: 'red' })
  } finally {
    loading.value = false
  }
}

async function handleToggleBan(user) {
  try {
    await api.fetch(`/v1/admin/users/${user.id}/${user.is_banned ? 'unban' : 'ban'}`, { method: 'POST' })
    toast.add({ title: user.is_banned ? 'User unbanned' : 'User banned', color: 'green' })
    fetchUsers()
  } catch (err) {
    toast.add({ title: 'Action failed', color: 'red' })
  }
}

onMounted(fetchUsers)
</script>
