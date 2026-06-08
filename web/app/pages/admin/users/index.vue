<template>
  <div class="space-y-6">
    <PageHeader
      eyebrow="Administration"
      title="Users"
      subtitle="Manage user identities and access control"
    />

    <UCard>
      <div class="flex items-center justify-between mb-4">
        <UInput v-model="search" icon="i-heroicons-magnifying-glass" placeholder="Search users..." class="max-w-xs" @input="debouncedFetch" />
        <div class="flex gap-2">
          <USelectMenu v-model="selectedStatus" :options="statusOptions" placeholder="All Statuses" class="w-40" @change="fetchUsers" />
          <UInput v-model="selectedNamespace" placeholder="Namespace" class="w-40" @input="debouncedFetch" />
        </div>
      </div>

      <UTable :rows="users" :columns="columns" :loading="loading">
        <template #username-data="{ row }">
          <div class="flex items-center gap-2">
            <UAvatar :alt="row.username" size="xs" />
            <span class="font-medium text-slate-900 dark:text-white">{{ row.username }}</span>
          </div>
        </template>

        <template #status-data="{ row }">
          <UBadge :color="statusColor(row.status)" variant="soft" size="xs">
            {{ statusLabel(row.status) }}
          </UBadge>
        </template>

        <template #roles-data="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge v-for="role in row.roles" :key="role" color="primary" variant="subtle" size="xs">{{ role }}</UBadge>
          </div>
        </template>

        <template #permissions-data="{ row }">
          <div class="flex flex-wrap gap-1">
            <UBadge v-for="perm in row.permissions" :key="perm" color="gray" variant="soft" size="xs">{{ perm }}</UBadge>
          </div>
        </template>

        <template #actions-data="{ row }">
          <div class="flex gap-1">
            <UTooltip text="View Details">
              <UButton size="xs" variant="ghost" icon="i-heroicons-eye" @click="openDetail(row)" />
            </UTooltip>
            <UTooltip :text="statusLabel(row.status) === 'Banned' ? 'Activate' : 'Ban'">
              <UButton size="xs" variant="ghost" :color="statusLabel(row.status) === 'Banned' ? 'green' : 'red'" :icon="statusLabel(row.status) === 'Banned' ? 'i-heroicons-check-circle' : 'i-heroicons-no-symbol'" @click="selectedUser = row; updateStatus(statusLabel(row.status) === 'Banned' ? 'USER_STATUS_ACTIVE' : 'USER_STATUS_BANNED')" />
            </UTooltip>
          </div>
        </template>
      </UTable>

      <div class="flex justify-between items-center mt-4" v-if="totalPages > 1">
        <span class="text-sm text-slate-500">{{ totalCount }} users total</span>
        <UPagination v-model="page" :page-count="pageSize" :total="totalCount" @update:model-value="fetchUsers" />
      </div>
    </UCard>

    <!-- User Detail Modal -->
    <UModal v-model="isDetailOpen">
      <UCard v-if="selectedUser">
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold">User: {{ selectedUser.username }}</h3>
            <UButton color="gray" variant="ghost" icon="i-heroicons-x-mark-20-solid" @click="isDetailOpen = false" />
          </div>
        </template>

        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div><span class="text-slate-500">ID:</span> {{ selectedUser.id }}</div>
            <div><span class="text-slate-500">Email:</span> {{ selectedUser.email }}</div>
            <div><span class="text-slate-500">Namespace:</span> {{ selectedUser.namespace }}</div>
            <div>
              <span class="text-slate-500">Status:</span>
              <UBadge :color="statusColor(selectedUser.status)" variant="soft" size="xs" class="ml-1">{{ statusLabel(selectedUser.status) }}</UBadge>
            </div>
          </div>

          <UDivider />

          <div>
            <h4 class="font-medium mb-2">Roles</h4>
            <div class="flex flex-wrap gap-2 mb-2">
              <span v-for="role in selectedUser.roles" :key="role" class="inline-flex items-center gap-1">
                <UBadge color="primary" variant="subtle">{{ role }}</UBadge>
                <UButton icon="i-heroicons-x-mark" size="2xs" color="red" variant="ghost" @click="revokeRole(role)" />
              </span>
              <span v-if="!selectedUser.roles?.length" class="text-sm text-slate-500">No roles</span>
            </div>
            <div class="flex gap-2">
              <USelectMenu v-model="newRole" :options="grantableRoles" placeholder="Select or type role..." size="sm" class="flex-1" searchable creatable @create="onCreateRole" />
              <UButton size="sm" :disabled="!newRole" @click="grantRole">Grant</UButton>
            </div>
          </div>

          <UDivider />

          <div>
            <h4 class="font-medium mb-2">Permissions</h4>
            <div class="flex flex-wrap gap-2 mb-2">
              <span v-for="perm in selectedUser.permissions" :key="perm" class="inline-flex items-center gap-1">
                <UBadge color="gray" variant="soft">{{ perm }}</UBadge>
                <UButton icon="i-heroicons-x-mark" size="2xs" color="red" variant="ghost" @click="revokePermission(perm)" />
              </span>
              <span v-if="!selectedUser.permissions?.length" class="text-sm text-slate-500">No permissions</span>
            </div>
            <div class="flex gap-2">
              <USelectMenu v-model="newPermission" :options="grantablePermissions" placeholder="Select or type permission..." size="sm" class="flex-1" searchable creatable @create="onCreatePermission" />
              <UButton size="sm" :disabled="!newPermission" @click="grantPermission">Grant</UButton>
            </div>
          </div>

          <UDivider />

          <div class="flex gap-2">
            <UButton v-if="statusLabel(selectedUser.status) !== 'Banned'" color="red" variant="soft" @click="updateStatus('USER_STATUS_BANNED')">Ban User</UButton>
            <UButton v-if="statusLabel(selectedUser.status) !== 'Active'" color="green" variant="soft" @click="updateStatus('USER_STATUS_ACTIVE')">Activate</UButton>
            <UButton color="orange" variant="soft" @click="isResetPasswordOpen = true">Reset Password</UButton>
          </div>
        </div>
      </UCard>
    </UModal>

    <!-- Reset Password Modal -->
    <UModal v-model="isResetPasswordOpen">
      <UCard>
        <template #header>
          <h3 class="text-base font-semibold">Reset Password for {{ selectedUser?.username }}</h3>
        </template>
        <form @submit.prevent="resetPassword" class="space-y-4">
          <UFormGroup label="New Password">
            <UInput v-model="newPassword" type="password" placeholder="Enter new password" />
          </UFormGroup>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="isResetPasswordOpen = false">Cancel</UButton>
            <UButton type="submit" :loading="resetting">Reset</UButton>
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
const search = ref('')
const selectedNamespace = ref('default')
const selectedStatus = ref(null)
const users = ref([])
const page = ref(1)
const pageSize = 20
const totalCount = ref(0)
const totalPages = ref(0)

const isDetailOpen = ref(false)
const selectedUser = ref(null)
const newRole = ref(null)
const newPermission = ref(null)
const availableRoles = ref([])
const availablePermissions = ref([])
const isResetPasswordOpen = ref(false)
const newPassword = ref('')
const resetting = ref(false)

const statusOptions = [
  { label: 'All', value: '' },
  { label: 'Active', value: 'active' },
  { label: 'Inactive', value: 'inactive' },
  { label: 'Banned', value: 'banned' },
]

const columns = [
  { key: 'username', label: 'User' },
  { key: 'email', label: 'Email' },
  { key: 'namespace', label: 'Namespace' },
  { key: 'roles', label: 'Roles' },
  { key: 'permissions', label: 'Permissions' },
  { key: 'status', label: 'Status' },
  { key: 'actions', label: '' }
]

function statusColor(s) {
  const label = statusLabel(s)
  if (label === 'Active') return 'green'
  if (label === 'Banned') return 'red'
  return 'gray'
}

function statusLabel(s) {
  if (s === 'USER_STATUS_ACTIVE' || s === 1) return 'Active'
  if (s === 'USER_STATUS_INACTIVE' || s === 2) return 'Inactive'
  if (s === 'USER_STATUS_BANNED' || s === 3) return 'Banned'
  return String(s)
}

let fetchTimeout = null
function debouncedFetch() {
  clearTimeout(fetchTimeout)
  fetchTimeout = setTimeout(() => { page.value = 1; fetchUsers() }, 300)
}


async function fetchUsers() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page_size: String(pageSize), page: String(page.value), namespace: selectedNamespace.value || 'default' })
    if (search.value) params.set('query', search.value)
    if (selectedStatus.value?.value) params.set('status', selectedStatus.value.value)
    const res = await api.fetch(`/v1/admin/users?${params}`)
    users.value = res.users || []
    totalCount.value = res.total_count || 0
    totalPages.value = res.total_pages || 0
  } catch {
    toast.add({ title: 'Error fetching users', color: 'red' })
  } finally {
    loading.value = false
  }
}

const grantableRoles = computed(() => {
  const assigned = new Set(selectedUser.value?.roles || [])
  return availableRoles.value.filter(r => !assigned.has(r))
})

const grantablePermissions = computed(() => {
  const assigned = new Set(selectedUser.value?.permissions || [])
  return availablePermissions.value.filter(p => !assigned.has(p))
})

function onCreateRole(val) {
  availableRoles.value.push(val)
  newRole.value = val
}

function onCreatePermission(val) {
  availablePermissions.value.push(val)
  newPermission.value = val
}

async function fetchRolesAndPermissions(namespace) {
  try {
    const [rolesRes, permsRes] = await Promise.all([
      api.fetch(`/v1/admin/roles?page_size=100&namespace=${namespace}`),
      api.fetch(`/v1/admin/permissions?page_size=100&namespace=${namespace}`),
    ])
    availableRoles.value = (rolesRes.roles || []).map(r => r.name)
    availablePermissions.value = (permsRes.permissions || []).map(p => p.name)
  } catch {}
}

async function openDetail(row) {
  try {
    const user = await api.fetch(`/v1/admin/users/${row.id}`)
    selectedUser.value = user
    await fetchRolesAndPermissions(user.namespace || selectedNamespace.value)
    isDetailOpen.value = true
  } catch {
    toast.add({ title: 'Error loading user', color: 'red' })
  }
}

async function updateStatus(status) {
  try {
    await api.fetch(`/v1/admin/users/${selectedUser.value.id}/status`, { method: 'PATCH', body: { status } })
    toast.add({ title: 'Status updated', color: 'green' })
    isDetailOpen.value = false
    fetchUsers()
  } catch {
    toast.add({ title: 'Failed to update status', color: 'red' })
  }
}

async function grantRole() {
  if (!newRole.value) return
  const role = typeof newRole.value === 'string' ? newRole.value : newRole.value.label
  try {
    await api.fetch(`/v1/admin/users/${selectedUser.value.id}/roles`, { method: 'POST', body: { roles: [role] } })
    selectedUser.value.roles = [...(selectedUser.value.roles || []), role]
    newRole.value = null
    toast.add({ title: 'Role granted', color: 'green' })
  } catch { toast.add({ title: 'Failed to grant role', color: 'red' }) }
}

async function revokeRole(role) {
  try {
    await api.fetch(`/v1/admin/users/${selectedUser.value.id}/roles?roles=${encodeURIComponent(role)}`, { method: 'DELETE' })
    selectedUser.value.roles = selectedUser.value.roles.filter(r => r !== role)
    toast.add({ title: 'Role revoked', color: 'green' })
  } catch { toast.add({ title: 'Failed to revoke role', color: 'red' }) }
}

async function grantPermission() {
  if (!newPermission.value) return
  const perm = typeof newPermission.value === 'string' ? newPermission.value : newPermission.value.label
  try {
    await api.fetch(`/v1/admin/users/${selectedUser.value.id}/permissions`, { method: 'POST', body: { permissions: [perm] } })
    selectedUser.value.permissions = [...(selectedUser.value.permissions || []), perm]
    newPermission.value = null
    toast.add({ title: 'Permission granted', color: 'green' })
  } catch { toast.add({ title: 'Failed to grant permission', color: 'red' }) }
}

async function revokePermission(perm) {
  try {
    await api.fetch(`/v1/admin/users/${selectedUser.value.id}/permissions?permissions=${encodeURIComponent(perm)}`, { method: 'DELETE' })
    selectedUser.value.permissions = selectedUser.value.permissions.filter(p => p !== perm)
    toast.add({ title: 'Permission revoked', color: 'green' })
  } catch { toast.add({ title: 'Failed to revoke permission', color: 'red' }) }
}

async function resetPassword() {
  resetting.value = true
  try {
    await api.fetch(`/v1/admin/users/${selectedUser.value.id}/reset-password`, { method: 'POST', body: { new_password: newPassword.value } })
    toast.add({ title: 'Password reset', color: 'green' })
    isResetPasswordOpen.value = false
    newPassword.value = ''
  } catch { toast.add({ title: 'Failed to reset password', color: 'red' }) }
  finally { resetting.value = false }
}

onMounted(fetchUsers)
</script>
