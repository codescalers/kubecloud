<template>
  <div class="h-100 d-flex justify-center align-center">
    <v-card max-width="700" width="100%">
      <v-card-text>
        <v-img
          src="https://staging.myceliumcloud.tf/assets/logo-DY_xoQWE.png"
          width="150"
          height="40"
          class="mx-auto"
        />

        <v-divider class="my-4" />

        <h3 class="text-h4 text-center mb-4">Create Account</h3>

        <p class="mb-10 text-center">Join Mycelium Cloud and start your journey</p>

        <v-form>
          <v-row>
            <v-col>
              <v-text-field variant="outlined" label="Username" />
            </v-col>
            <v-col>
              <v-text-field variant="outlined" label="Email Address" />
            </v-col>
          </v-row>

          <v-row>
            <v-col>
              <v-text-field variant="outlined" label="Password" type="password" />
            </v-col>
            <v-col>
              <v-text-field variant="outlined" label="Confirm Password" type="password" />
            </v-col>
          </v-row>

          <div>
            <p>Password must contain at least 8 characters, including:</p>
            <ul :style="{ listStyle: 'square' }" class="ml-6 text-caption">
              <li>One uppercase letter (A-Z)</li>
              <li>One lowercase letter (a-z)</li>
              <li>One number (0-9)</li>
              <li>One special character (!@#$%^&*)</li>
            </ul>
          </div>

          <v-btn
            type="submit"
            block
            size="large"
            text="Create Account"
            variant="outlined"
            class="mt-6"
          />
        </v-form>
        <div class="text-caption d-flex justify-center align-center my-4">
          By creating an account you agree to our
          <v-btn text="Terms & Conditions" variant="text" size="x-small" color="primary" /> and
          <v-btn text="Privacy Policy" variant="text" size="x-small" color="primary" />.
        </div>

        <div class="d-flex align-center justify-center my-5">
          <p>Already have an account?</p>
          <v-btn to="/login" text="Login" size="small" variant="text" color="primary" />
        </div>

        <v-btn text="Back to Home" to="/" variant="text" block prepend-icon="mdi-arrow-left" />
      </v-card-text>
    </v-card>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: "plain",
})
</script>
