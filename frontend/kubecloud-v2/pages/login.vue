<template>
  <div class="h-100 d-flex justify-center align-center">
    <v-card max-width="400" width="100%">
      <v-card-text>
        <v-img
          src="https://staging.myceliumcloud.tf/assets/logo-DY_xoQWE.png"
          width="150"
          height="40"
          class="mx-auto"
        />

        <v-divider class="my-4" />

        <h3 class="text-h4 text-center mb-4">Welcome Back!</h3>

        <p class="mb-10 text-center">Sign in to your Mycelium Cloud account</p>

        <v-form>
          <v-text-field variant="outlined" label="Email Address" />
          <v-text-field variant="outlined" label="Password" type="password" />

          <v-btn type="button" text="Forgot Password?" variant="text" size="small" />

          <v-btn type="submit" block size="large" text="Login" variant="outlined" class="mt-4" />
        </v-form>

        <div class="d-flex align-center justify-center my-5">
          <p>Don't have an account?</p>
          <v-btn to="/register" text="Register" size="small" variant="text" color="primary" />
        </div>

        <v-btn text="Back to Home" to="/" variant="text" block prepend-icon="mdi-arrow-left" />
      </v-card-text>
    </v-card>
  </div>
</template>
