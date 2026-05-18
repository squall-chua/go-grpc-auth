<template>
  <div class="flex flex-col items-center justify-center py-12 px-4">
    <UCard class="w-full max-w-md p-4">
      <div class="text-center space-y-4">
        <UIcon name="i-heroicons-lock-closed" class="w-12 h-12 text-slate-400 mx-auto" />
        <h2 class="text-2xl font-bold">Reset Password</h2>
        <p class="text-sm text-slate-500">
          Password reset is not yet available through the web interface. Please contact your system administrator to reset your password.
        </p>
        <UButton variant="ghost" to="/login" block>Back to Login</UButton>
      </div>
    </UCard>
  </div>
</template>
