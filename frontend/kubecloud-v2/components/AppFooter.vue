<template>
  <v-footer class="bg-background border-t py-0" app absolute>
    <v-container class="py-9">
      <v-row align="center">
        <v-col>
          <v-img
            src="https://staging.myceliumcloud.tf/assets/logo-DY_xoQWE.png"
            width="130"
            height="34"
          />
        </v-col>

        <v-spacer />
        <v-col cols="auto">
          <v-btn size="small" text="Documentation" to="/docs" variant="plain" />
          <v-btn
            size="small"
            text="Github"
            target="_blank"
            href="https://github.com/codescalers/kubecloud"
            variant="plain"
          />
          <v-btn
            size="small"
            text="Terms & Conditions"
            to="/terms-and-conditions"
            variant="plain"
          />
          <v-btn size="small" text="Privacy Policy" to="/privacy-policy" variant="plain" />
        </v-col>
      </v-row>

      <v-divider class="my-6" />

      <p class="text-primary text-center" :style="{ fontSize: '12px', color: '#60A5FA' }">
        © 2025 Mycelium Cloud. All rights reserved.
      </p>
    </v-container>
  </v-footer>
</template>
