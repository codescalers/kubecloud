<template>
  <v-footer app name="footer">
    <v-container>
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
          <v-btn text="Documentation" to="/docs" variant="plain" />
          <v-btn
            text="Github"
            target="_blank"
            href="https://github.com/codescalers/kubecloud"
            variant="plain"
          />
          <v-btn text="Terms & Conditions" to="/terms-and-conditions" variant="plain" />
          <v-btn text="Privacy Policy" to="/privacy-policy" variant="plain" />
        </v-col>
      </v-row>

      <v-divider class="my-4" />

      <p class="text-primary text-center">© 2025 Mycelium Cloud. All rights reserved.</p>
    </v-container>
  </v-footer>
</template>
