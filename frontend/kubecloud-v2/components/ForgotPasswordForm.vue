<template>
  <div>
    <v-btn @click="$emit('back-to-login')">Back To Login</v-btn>

    <h3 class="text-h4 text-center mb-4">Forgot Password?</h3>
  </div>
</template>

<script setup lang="ts">
defineEmits<{ (e: "back-to-login"): void }>()
</script>
