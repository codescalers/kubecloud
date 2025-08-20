<template>
  <v-container fluid class="h-100">
    <v-row justify="center" align="center" class="h-100">
      <v-col cols="12" sm="10" md="8" lg="6" xl="4">
        <v-card elevation="0">
          <v-card-text class="text-center pa-4 pa-sm-6 pa-md-8">
            <img 
              src="@/assets/icons/maintenance.svg" 
              alt="Maintenance Tools" 
              class="maintenance-icon mb-4"
            />
              
            
            <v-card-title class="text-subtitle-1 text-sm-h5 text-md-h4 font-weight-bold mb-2">
             We'll be back soon
            </v-card-title>
            
            <v-card-text class="text-body-2">
              We’re currently performing <span class="font-weight-bold">maintenance</span> to improve your experience.<br>
              Thanks for your patience — we’ll be back online shortly!
            </v-card-text>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
</script>

<style scoped>
.maintenance-icon {
  width: 85px;
  height: 85px;
  margin-bottom: 16px;
}


</style>

