<template>
  <v-row no-gutters class="h-100">
    <v-col cols="6">
      <v-card width="100%" height="100%" class="h-100 d-flex align-center">
        <v-card-text class="px-10">
          <h3 class="text-h4 text-center mb-4">Create Account</h3>

          <p class="mb-10 text-center">Join Mycelium Cloud and start your journey</p>

          <v-form v-model="valid" @submit.prevent="handleSubmit">
            <v-row>
              <v-col cols="6">
                <v-text-field
                  v-model.trim="username"
                  variant="outlined"
                  prepend-inner-icon="mdi-account"
                  label="Username"
                  placeholder="Enter your username"
                  autofocus
                  :rules="[
                    (v) => !!v || 'Username is required',
                    (v) => v.length >= 3 || 'Username must be at least 3 characters',
                    (v) => v.length <= 12 || 'Username must be less than 12 characters',
                  ]"
                />
              </v-col>

              <v-col cols="6">
                <v-text-field
                  v-model.trim="email"
                  prepend-inner-icon="mdi-email"
                  variant="outlined"
                  label="Email Address"
                  placeholder="Enter your email address"
                  :rules="[
                    (v) => !!v || 'Email address is required',
                    (v) => (v.includes('@') && v.includes('.')) || 'Email address must be valid',
                  ]"
                />
              </v-col>

              <v-col cols="12">
                <PasswordInput v-model="password" />
              </v-col>
            </v-row>

            <v-btn
              type="submit"
              block
              size="large"
              text="Create Account"
              prepend-icon="mdi-account-plus"
              variant="outlined"
              class="mt-6"
              :disabled="!valid"
            />
          </v-form>
          <div class="text-caption d-flex justify-center align-center my-4">
            By creating an account you agree to our
            <v-btn text="Terms & Conditions" variant="text" size="x-small" color="primary" /> and
            <v-btn text="Privacy Policy" variant="text" size="x-small" color="primary" />.
          </div>

          <!-- <div class="d-flex align-center justify-center my-5">
            <p>Already have an account?</p>
            <v-btn to="/login" text="Login" size="small" variant="text" color="primary" />
          </div> -->
        </v-card-text>
      </v-card>
    </v-col>

    <v-col cols="6">
      <div class="h-100 w-100 bg-primary d-flex justify-center align-center">KubeCloud</div>
    </v-col>
  </v-row>
</template>

<script lang="ts" setup>
const valid = ref(false)

const username = ref("")
const email = ref("")
const password = ref("")

function handleSubmit() {
  console.log({
    username: username.value,
    email: email.value,
    password: password.value,
  })
}
</script>
