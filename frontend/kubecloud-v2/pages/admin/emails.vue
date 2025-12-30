<template>
  <div>
    <div class="mb-8">
      <h1 class="text-h4 font-weight-bold">Emails</h1>
      <p class="text-body-1 mt-2 opacity-70">Send emails to all platform users</p>
    </div>

    <v-alert
      type="warning"
      variant="tonal"
      class="mb-8 rounded-lg border-warning border"
      title="Broadcast Warning"
      :style="{ '--v-border-opacity': 0.5 }"
    >
      This action will send an email to
      <strong>all registered users</strong> on the platform immediately. Be aware that this action
      cannot be <strong>undone</strong>. Please double-check the content.
    </v-alert>

    <v-form class="mb-16">
      <v-row>
        <v-col cols="12">
          <p class="text-subtitle-2 opacity-70 mb-2">Email Subject</p>
          <v-text-field
            placeholder="Enter the subject of the email"
            variant="outlined"
            prepend-inner-icon="mdi-alpha-t"
            autofocus
          />
        </v-col>

        <v-col cols="12">
          <div class="mb-2 d-flex align-center justify-space-between">
            <p class="text-subtitle-2 opacity-70">Email Content</p>
            <p class="text-caption opacity-50">Markdown supported</p>
          </div>
          <MarkdownEditor label="Write your message here..." />
        </v-col>

        <v-col cols="12" class="mt-4">
          <p class="text-subtitle-2 opacity-70 mb-2">
            Attachments <span class="text-caption opacity-50">(optional)</span>
          </p>
          <v-card
            class="d-flex align-center justify-center bg-2 border-dashed border-md ga-2"
            :ripple="!files.length"
            @click="open()"
          >
            <div v-if="files.length" class="d-flex align-center ga-4 flex-wrap">
              <p
                v-for="attachment in files"
                :key="attachment.name"
                class="text-subtitle-2 border d-inline-block px-3 py-1 rounded-lg"
                :style="{ backgroundColor: 'rgb(var(--v-bg-2))' }"
                @click.stop="removeFile(attachment)"
              >
                <span class="opacity-70" v-text="attachment.name" />
              </p>
            </div>

            <template v-else>
              <v-avatar variant="tonal" size="80">
                <v-icon icon="mdi-paperclip" />
              </v-avatar>
              <p class="text-subtitle-2 opacity-70">Click to add attachments</p>
            </template>
          </v-card>
        </v-col>

        <v-col cols="12">
          <v-divider class="mt-6" />
        </v-col>

        <v-col cols="12">
          <div class="d-flex justify-space-between align-center">
            <div class="d-flex align-center ga-2">
              <v-icon icon="mdi-information-outline" size="x-small" color="info" />
              <p class="text-subtitle-2 opacity-70">
                <span>Estimated recipients:</span>&nbsp;
                <v-progress-circular v-if="isLoading" indeterminate size="16" width="2" />
                <strong v-else v-text="data?.total_users" />
              </p>
            </div>

            <div class="d-flex ga-2 align-center">
              <v-btn variant="plain">Save Draft</v-btn>

              <v-btn
                type="submit"
                class="btn-form"
                text="Send to All Users"
                append-icon="mdi-send"
                variant="outlined"
              />
            </div>
          </div>
        </v-col>
      </v-row>
    </v-form>
  </div>
</template>

<script setup lang="ts">
const { open, files, removeFile } = useFilesDialog()
const { isLoading, data } = useStats()
</script>
