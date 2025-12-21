<template>
  <div>Dashboard Overview</div>
</template>

<script setup lang="ts">
const api = useApi()
onMounted(async () => {
  const { data } = await api.users.getUser()
  console.log(data)
})
</script>
